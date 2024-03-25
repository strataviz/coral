package worker

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/util"
)

// Images represents and tracks the images that the node requires and
// their states.
type Images struct {
	// Perhaps we use a map with resource hash as key?
	// resources []stvziov1.Image
	// Logger for logging.
	logger logr.Logger
	// Limit images to a specific namespace.
	namespace string
	// images is a map of images that the node is tracking.  A map is used
	// so we can deduplicate images.
	images map[string]bool
}

// NewImages creates a new images object.
func NewImages() *Images {
	return &Images{
		namespace: "",
		images:    make(map[string]bool),
	}
}

// WithLogger sets the logger for the images object.
func (i *Images) WithLogger(log logr.Logger) *Images {
	i.logger = log
	return i
}

// WithNamespace sets the namespace that will constrain searches for images.
func (i *Images) WithNamespace(ns string) *Images {
	i.namespace = ns
	return i
}

// Included checks if the node labels match the selectors.
func (i *Images) Included(nodeLabels map[string]string, selectors []stvziov1.NodeSelector) (bool, error) {
	s := labels.NewSelector()
	for _, selector := range selectors {
		req, err := labels.NewRequirement(selector.Key, selector.Operator, selector.Values)
		if err != nil {
			return false, err
		}
		s = s.Add(*req)
	}

	return s.Matches(labels.Set(nodeLabels)), nil
}

// GetImages retrieves all of the images that the system knows about.  We currently
// grab all of them and then filter out the ones that do not match the node labels.
// Optimally I would like to reverse that and filter out the images that don't match
// the node labels, but that doesn't seem possible.
func (i *Images) GetImages(ctx context.Context, cli client.Client, labels map[string]string) error {
	imgs := new(stvziov1.ImageList)
	err := cli.List(ctx, imgs, &client.ListOptions{
		Namespace: i.namespace,
	})
	if err != nil {
		return err
	}

	for _, img := range imgs.Items {
		if !img.DeletionTimestamp.IsZero() {
			// If we are pending deletion, we don't want the image.
			continue
		}

		included, err := i.Included(labels, img.Spec.Selector)
		if err != nil {
			return err
		}

		if included {
			i.AddImage(img)
		}
	}

	return nil
}

// AddImage adds an image to the state tracker.
func (i *Images) AddImage(image stvziov1.Image) {
	for _, im := range image.Spec.Images {
		for _, tag := range im.Tags {
			i.images[fmt.Sprintf("%s:%s", *im.Name, tag)] = true
		}
	}
}

// List returns the list of images.
func (i *Images) List() []string {
	l := make([]string, 0, len(i.images))
	for k := range i.images {
		l = append(l, k)
	}
	return l
}

func (i *Images) HashedMap() map[string]string {
	h := make(map[string]string)
	for k := range i.images {
		h[util.ImageHasher(k)] = k
	}
	return h
}
