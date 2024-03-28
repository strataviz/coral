package util

import (
	corev1 "k8s.io/api/core/v1"
)

func ContainerNamesMap(containers []corev1.Container) map[string]bool {
	result := make(map[string]bool)
	for _, c := range containers {
		result[c.Name] = true
	}

	return result
}

func ContainerNamesMapInclude(containers []corev1.Container, included []string) map[string]bool {
	result := make(map[string]bool)
	for _, c := range containers {
		for _, name := range included {
			result[c.Name] = c.Name == name
		}
	}

	return result
}

func ContainerNamesMapExclude(containers []corev1.Container, excluded []string) map[string]bool {
	result := make(map[string]bool)
	for _, c := range containers {
		for _, name := range excluded {
			result[c.Name] = c.Name != name
		}
	}

	return result
}

func ContainerImages(containers []corev1.Container) []string {
	var result []string
	for _, c := range containers {
		result = append(result, c.Image)
	}

	return result
}

func ContainerImageInclude(containers []corev1.Container, included []string) []string {
	var result []string
	for _, c := range containers {
		for _, name := range included {
			if c.Name == name {
				result = append(result, c.Image)
			}
		}
	}

	return result
}

func ContainerImageExclude(containers []corev1.Container, excluded []string) []string {
	var result []string
	for _, c := range containers {
		found := false
		for _, name := range excluded {
			if c.Name == name {
				found = true
			}
		}
		if !found {
			result = append(result, c.Image)
		}
	}

	return result
}

// func ContainerInclude(containers []corev1.Container, included []string) []corev1.Container {
// 	var result []corev1.Container
// 	for _, c := range containers {
// 		for _, name := range included {
// 			if c.Name == name {
// 				result = append(result, c)
// 			}
// 		}
// 	}

// 	return result
// }

// func ContainerExclude(containers []corev1.Container, excluded []string) []corev1.Container {
// 	var result []corev1.Container
// 	for _, c := range containers {
// 		found := false
// 		for _, name := range excluded {
// 			if c.Name == name {
// 				found = true
// 			}
// 		}
// 		if !found {
// 			result = append(result, c)
// 		}
// 	}

// 	return result
// }
