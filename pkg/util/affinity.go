package util

import (
	corev1 "k8s.io/api/core/v1"
)

func MergeRequirements(a, b []*corev1.NodeSelectorRequirement) []*corev1.NodeSelectorRequirement {
	results := make(map[string]*corev1.NodeSelectorRequirement)

	for _, r := range a {
		results[r.Key] = r
	}

	for _, r := range b {
		results[r.Key] = r
	}

	var resultSlice []*corev1.NodeSelectorRequirement
	for _, r := range results {
		resultSlice = append(resultSlice, r)
	}

	return resultSlice
}
