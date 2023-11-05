package buildset

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DesiredState struct {
	Deployment *appsv1.Deployment
}

// TODO: I think we need the queue resource here as well.
func GetDesiredState(observed *ObservedState) (*DesiredState, error) {
	return &DesiredState{
		Deployment: getDesiredDeploymentState(observed),
	}, nil
}

func getDesiredDeploymentState(observed *ObservedState) *appsv1.Deployment {
	expected := observed.buildSet.DeepCopy()

	// TODO: I shouldn't need startup or readiness probes for the watchset
	// workers, but I'll keep them in here and empty for now in case I change
	// my mine.
	var startupProbe = &corev1.Probe{}
	var readinessProbe = &corev1.Probe{}

	var container = corev1.Container{
		Name:    "coral",
		Image:   fmt.Sprintf("%s:%s", *expected.Spec.Image, *expected.Spec.Version),
		Command: strings.Split(*expected.Spec.Command, " "),
		Args: []string{
			"watch",
		},
		Env: []corev1.EnvVar{
			{
				Name: "POD_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
			{
				Name:  "SERVER_NAME",
				Value: "$(POD_NAME)",
			},
		},

		Resources:      *expected.Spec.Resources,
		StartupProbe:   startupProbe,
		ReadinessProbe: readinessProbe,
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            expected.Name,
			Namespace:       expected.Namespace,
			OwnerReferences: getOwnerReferences(observed),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": expected.Name},
			},
			Replicas: expected.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": expected.Name},
				},
				Spec: corev1.PodSpec{
					// TODO: configurable
					TerminationGracePeriodSeconds: &[]int64{10}[0],
					// TODO: Merge containers in case there are some sidecars added
					Containers: []corev1.Container{container},
				},
			},
		},
	}
}

func getOwnerReferences(observed *ObservedState) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion:         observed.buildSet.APIVersion,
			Kind:               observed.buildSet.Kind,
			Name:               observed.buildSet.ObjectMeta.Name,
			UID:                observed.buildSet.ObjectMeta.UID,
			Controller:         &[]bool{true}[0],
			BlockOwnerDeletion: &[]bool{false}[0],
		},
	}
}
