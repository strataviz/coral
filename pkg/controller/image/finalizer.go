package image

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/util"
)

const (
	Finalizer = "image.stvz.io/finalizer"
)

type finalizer struct {
	client client.Client
	image  *stvziov1.Image
}

func (f *finalizer) reconcile(ctx context.Context) (bool, error) {
	logger := log.FromContext(ctx)

	if f.image.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(f.image, Finalizer) {
			logger.V(8).Info("adding finalizer", "finalizer", Finalizer)
			return true, f.add(ctx)
		}
	} else {
		if controllerutil.ContainsFinalizer(f.image, Finalizer) {
			logger.V(8).Info("finishing cleanup", "finalizer", Finalizer)
			return true, f.finish(ctx)
		}
	}

	return false, nil
}

// TODO: the finalizer is getting in the way by updating the resource and causing the update
// later on to fail due ot the resource being modified.
func (f *finalizer) add(ctx context.Context) error {
	controllerutil.AddFinalizer(f.image, Finalizer)
	return f.client.Update(ctx, f.image)
}

func (f *finalizer) finish(ctx context.Context) error {
	logger := log.FromContext(ctx)
	// TODO: there's a race here where the workers are looking for the images to disappear
	// from the nodes before we remove the finalizer.  We can get around this by setting the
	// labels from here, but that goes back to the original responsibilities issue that I was
	// trying to solve by moving everything to the worker.  Could I check for a finalizer
	// state in the worker?
	for _, image := range f.image.Spec.Images {
		for _, tag := range image.Tags {
			key := util.ImageLabelKey(tag)
			nodes := new(corev1.NodeList)

			reqs, err := labels.NewRequirement(key, selection.Exists, nil)
			if err != nil {
				return err
			}
			selector := labels.NewSelector().Add(*reqs)

			err = f.client.List(ctx, nodes, &client.ListOptions{LabelSelector: selector})
			if err != nil {
				return err
			}

			// If there are nodes that still have the image present, then we don't delete
			// the finalizer.  This will keep the image resource around so the node worker
			// that has not yet removed the images can use the information contained in the
			// resource to do so.  This does have the side effect of potentially not deleting
			// the image resource if the node is in a bad state or the worker is unable to
			// clean.  This is a tradeoff that we are willing to make in this case as it
			// allows us to visualize that the entire cluster is in a consistent state.
			logger.V(8).Info("checking nodes for image", "image", key, "nodes", len(nodes.Items))
			if len(nodes.Items) > 0 {
				return nil
			}
		}
	}
	return f.remove(ctx)
}

func (f *finalizer) remove(ctx context.Context) error {
	controllerutil.RemoveFinalizer(f.image, Finalizer)
	return f.client.Update(ctx, f.image)
}
