package worker

import (
	"strings"
	"sync"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/util"
)

const (
	LabelPrefix = "image.stvz.io/"
)

type NodeImageHashes map[string]stvziov1.ImageState

type NodeState struct {
	DiskPressure    bool              `json:"diskPressure"`
	PidPressure     bool              `json:"pidPressure"`
	Ready           bool              `json:"ready"`
	ImageHashes     map[string]string `json:"images"`
	Labels          map[string]string `json:"labels"`
	AvailableImages map[string]string `json:"availableImages"`
	logger          logr.Logger
	sync.Mutex
}

func NewNodeState(node *corev1.Node) *NodeState {
	n := &NodeState{
		Ready:           false,
		DiskPressure:    false,
		PidPressure:     false,
		ImageHashes:     make(map[string]string),
		AvailableImages: make(map[string]string),
		logger:          logr.Discard(),
	}

	n.Labels = node.Labels
	n.extractConditions(node.Status.Conditions)
	n.extractImageLabels(node.Labels)
	n.extractAvailableImages(node.Status.Images)
	return n
}

func (n *NodeState) GetImageMap() map[string]string {
	m := make(map[string]string)
	for k, v := range n.ImageHashes {
		m[k] = v
	}

	return m
}

// GetUpdatedLabels returns the node labels that have been updated with any new
// images or changed states.  It accepts a map of images as returned by the images
// object which includes the name of the image keyed by the hash of the image.
func (n *NodeState) GetUpdatedLabels(gathered map[string]string) map[string]string {
	labels := make(map[string]string)

	for key, value := range n.Labels {
		// Skip labels that are not image labels.
		if !strings.Contains(key, LabelPrefix) {
			labels[key] = value
			continue
		}

		labelHash := strings.TrimPrefix(key, LabelPrefix)
		if _, ok := gathered[labelHash]; ok {
			if _, ok := n.AvailableImages[labelHash]; ok {
				// If the image is available, set the label to available.
				labels[key] = string(stvziov1.ImageStateAvailable)
			} else {
				// The name isn't in the gathered images, so it needs to
				// be deleted.  This handles the case where an image has
				// been removed from the node either by the collector or
				// manually.  It should only be hit if we haven't updated
				// the node status yet.
				labels[key] = string(stvziov1.ImageStatePending)
			}
		} else {
			// If the image is not in the gathered and it's no longer on the node
			// we can just delete the label.  Otherwise set it up for deletion.
			if _, ok := n.AvailableImages[labelHash]; !ok {
				delete(n.Labels, key)
			} else {
				labels[key] = string(stvziov1.ImageStateDeleting)
			}
		}
	}

	// Add any new images to the labels that aren't yet found in the labels.
	for key := range gathered {
		if _, ok := n.Labels[LabelPrefix+key]; !ok {
			// There's two possibilities here: pending or already available.
			if _, ok := n.AvailableImages[key]; ok {
				labels[LabelPrefix+key] = string(stvziov1.ImageStateAvailable)
			} else {
				labels[LabelPrefix+key] = string(stvziov1.ImageStatePending)
			}
		}
	}

	return labels
}

// IsReady returns true if the node is ready to receive images. In the event
// that the node is not ready, the node has disk pressure as configured, or the
// node has PID pressure as configured, the node will not pull images.
func (n *NodeState) IsReady() bool {
	return n.Ready && !n.DiskPressure && !n.PidPressure
}

// NeedsImage returns true if the image label is marked as pending and the image
// is not found as available in the node image status.
func (n *NodeState) NeedsImage(name string) bool {
	hash := util.ImageHasher(name)
	label := LabelPrefix + hash

	_, available := n.AvailableImages[hash]
	state, hasLabel := n.Labels[label]

	// We don't do anything if the image label is unknown.  This is a safety
	// mechanism to ensure we don't do anything when we are in a bad state.
	if hasLabel && state == string(stvziov1.ImageStateUnknown) {
		return false
	}

	if !available && (hasLabel && state != string(stvziov1.ImageStateDeleting)) {
		return true
	} else if !available && !hasLabel {
		return true
	}

	return false
}

// WithLogger sets the logger for the node object.  By default the logger is set
// to discard.
func (n *NodeState) WithLogger(l logr.Logger) *NodeState {
	n.logger = l
	return n
}

// extractConditions extracts the conditions from the node status and adds them to
// the node object.  The conditions are used to determine if the node is ready to
// receive images.
func (n *NodeState) extractConditions(conditions []corev1.NodeCondition) {
	for _, condition := range conditions {
		switch condition.Type {
		case corev1.NodeReady:
			n.Ready = condition.Status == corev1.ConditionTrue
		case corev1.NodeDiskPressure:
			n.DiskPressure = condition.Status == corev1.ConditionTrue
		case corev1.NodePIDPressure:
			n.PidPressure = condition.Status == corev1.ConditionTrue
		}
	}
}

// extractImageLabels extracts the image labels from the node and adds them to the
// node object.  The image labels are used to track the status of the images on the
// node.
func (n *NodeState) extractImageLabels(labels map[string]string) {
	imageMap := make(map[string]string)
	for k, v := range labels {
		if strings.Contains(k, LabelPrefix) {
			stripped := strings.TrimPrefix(k, LabelPrefix)
			n.ImageHashes[stripped] = string(stvziov1.ImageStateFromString(v))
		}
	}

	n.ImageHashes = imageMap
}

// extractNodeImages extracts the images from the node status and adds them to the
// node object.  I'm not sure if I should just use the image labels or if I should
// get the image from the cri client.  I assume that the status will be updated as
// new images are added to the node, but there may be a delay in the status update
// especially if the API is having issues.
func (n *NodeState) extractAvailableImages(images []corev1.ContainerImage) {
	for _, image := range images {
		for _, name := range image.Names {
			hash := util.ImageHasher(name)
			n.AvailableImages[hash] = name
		}
	}
}
