package agent

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

// Node represents a kubernetes node with helpers for interacting with coral
// specific data.
type Node struct {
	availableImages       map[string]string
	conditionReady        bool
	conditionDiskPressure bool
	conditionPIDPressure  bool
	corev1.Node
}

// GetNode retrieves a node from the cache.
func GetNode(ctx context.Context, n string, c client.Client) (*Node, error) {
	node := corev1.Node{}
	err := c.Get(ctx, client.ObjectKey{Name: n}, &node)

	wrapped := Node{Node: *node.DeepCopy()}
	wrapped.gatherInfo()

	return &wrapped, err
}

// HasImage returns true if the image is available on the node (i.e. is registered in
// the node status).
func (n *Node) HasImage(image string) bool {
	_, ok := n.availableImages[image]
	return ok
}

// IsReady returns true if the node is ready.
func (n *Node) IsReady() bool {
	return n.conditionReady && n.conditionDiskPressure && n.conditionPIDPressure
}

// Refresh retrieves the latest node information from the cache.
func (n *Node) Refresh(ctx context.Context, c client.Client) error {
	node := corev1.Node{}
	err := c.Get(ctx, client.ObjectKey{Name: n.Name}, &node)
	if err != nil {
		return err
	}

	n.Node = *node.DeepCopy()
	n.gatherInfo()
	return nil
}

// StatusUpdate updates the node status.
func (n *Node) StatusUpdate(ctx context.Context, c client.Client) error {
	return c.Status().Update(ctx, &n.Node)
}

// UpdateLabels updates the node labels.
func (n *Node) UpdateLabels(ctx context.Context, c client.Client, labels map[string]string) error {
	n.Labels = labels
	return c.Update(ctx, &n.Node)
}

// gatherInfo retrieves the node information for the wrapped node.
func (n *Node) gatherInfo() {
	n.getImages()
	n.setCondition()
}

// getImages extracts the images from the node status and stores them in the
// availableImages map for easy lookup.
func (n *Node) getImages() {
	images := make(map[string]string)
	for _, image := range n.Status.Images {
		for _, name := range image.Names {
			// TODO: update the util method.
			images[name] = stvziov1.HashedImageLabelKey(name)
		}
	}

	n.availableImages = images
}

// setCondition sets the node conditions for easy lookup.
func (n *Node) setCondition() {
	for _, condition := range n.Status.Conditions {
		switch condition.Type {
		case corev1.NodeReady:
			n.conditionReady = condition.Status == corev1.ConditionTrue
		case corev1.NodeDiskPressure:
			n.conditionDiskPressure = condition.Status == corev1.ConditionFalse
		case corev1.NodePIDPressure:
			n.conditionPIDPressure = condition.Status == corev1.ConditionFalse
		}
	}
}
