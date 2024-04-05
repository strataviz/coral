// Copyright 2024 Coral Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
