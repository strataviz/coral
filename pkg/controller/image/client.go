package image

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetNodes(ctx context.Context, c client.Client, selector labels.Selector) []corev1.Node {
	nodes := new(corev1.NodeList)
	if err := c.List(ctx, nodes, &client.ListOptions{LabelSelector: selector}); err != nil {
		return nodes.Items
	}
	return []corev1.Node{}
}

func GetNodesByImageStatus(ctx context.Context, c client.Client, label string, status string) ([]corev1.Node, error) {
	reqs, err := labels.NewRequirement(label, selection.Equals, []string{status})
	if err != nil {
		return nil, err
	}
	selector := labels.NewSelector().Add(*reqs)

	nodes := GetNodes(ctx, c, selector)
	return nodes, nil
}
