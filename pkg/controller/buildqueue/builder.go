package buildqueue

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type DesiredState struct {
	StatefulSet     *appsv1.StatefulSet
	HeadlessService *corev1.Service
	Service         *corev1.Service
	ConfigMap       *corev1.ConfigMap
}

// TODO: I think we need the queue resource here as well.
func GetDesiredState(observed *ObservedState) (*DesiredState, error) {
	config := NewNatsConfig(observed.buildQueue.Name, *observed.buildQueue.Spec.Replicas)

	return &DesiredState{
		StatefulSet:     getDesiredStatefulSetState(observed, config),
		HeadlessService: getDesiredHeadlessServiceState(observed),
		Service:         getDesiredServiceState(observed),
		ConfigMap:       getConfigMapState(observed, config),
	}, nil
}

func getDesiredStatefulSetState(observed *ObservedState, config *NatsConfig) *appsv1.StatefulSet {
	expected := observed.buildQueue.DeepCopy()

	var startupProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(DefaultNatsPort),
			},
		},
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
		FailureThreshold:    10,
		SuccessThreshold:    1,
	}

	var readinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/healthz?js-server-only=true",
				Port: intstr.FromInt(DefaultNatsMonitorPort),
			},
		},
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
		FailureThreshold:    3,
		SuccessThreshold:    1,
	}

	var container = corev1.Container{
		Name:  "nats",
		Image: fmt.Sprintf("docker.io/nats:%s", *expected.Spec.Version),
		Lifecycle: &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"nats-server",
						"-sl=ldm=/var/run/nats/nats.pid",
					},
				},
			},
		},
		Args: []string{
			"--config",
			"/etc/nats/nats.conf",
			"--name",
			"$(POD_NAME)",
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
		// TODO: add additional ports when we add clustering support
		Ports: []corev1.ContainerPort{
			{ContainerPort: DefaultNatsMonitorPort, Name: "monitor"},
			{ContainerPort: DefaultNatsPort, Name: "nats"},
			{ContainerPort: DefaultNatsClusterPort, Name: "cluster"},
		},
		Resources:      *expected.Spec.Resources,
		StartupProbe:   startupProbe,
		ReadinessProbe: readinessProbe,
	}

	// Emptydir volume for the pid file in case we add the hot
	// reload feature.
	var pidVolume = corev1.Volume{
		Name: "pid",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      "pid",
		MountPath: "/var/run/nats",
	})

	var configVolume = corev1.Volume{
		Name: "config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: fmt.Sprintf("%s-nats", expected.Name),
				},
			},
		},
	}
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      "config",
		MountPath: "/etc/nats",
	})

	var claims = make([]corev1.PersistentVolumeClaim, 0)
	// TODO: I may need an init container for sysctl settings
	if expected.Spec.Volume != nil {
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      "data",
			MountPath: "/data/jetstream",
		})
		// TODO: May need to add an init container to chown the volume
		claims = append(claims, *expected.Spec.Volume)
		// TODO: Add additional volumes for certs and config
	}

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-nats", expected.Name),
			Namespace:       expected.Namespace,
			OwnerReferences: getOwnerReferences(observed),
			Annotations: map[string]string{
				"config/checksum": config.Checksum(),
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"buildqueue": expected.Name},
			},
			ServiceName:         fmt.Sprintf("%s-nats-headless", expected.Name),
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Replicas:            expected.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"buildqueue": expected.Name},
				},
				Spec: corev1.PodSpec{
					// TODO: configurable
					TerminationGracePeriodSeconds: &[]int64{10}[0],
					// TODO: Merge containers in case there are some sidecars added
					Containers: []corev1.Container{container},
					Volumes: []corev1.Volume{
						pidVolume,
						configVolume,
					},
				},
			},
			VolumeClaimTemplates: claims,
		},
	}
}

func getDesiredHeadlessServiceState(observed *ObservedState) *corev1.Service {
	expected := observed.buildQueue.DeepCopy()

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-nats-headless", expected.Name),
			Namespace:       expected.Namespace,
			OwnerReferences: getOwnerReferences(observed),
		},
		Spec: corev1.ServiceSpec{
			PublishNotReadyAddresses: true,
			Ports: []corev1.ServicePort{
				{Port: DefaultNatsMonitorPort, Name: "monitor"},
				{Port: DefaultNatsPort, Name: "nats"},
			},
			Selector:  map[string]string{"buildqueue": expected.Name},
			ClusterIP: corev1.ClusterIPNone,
		},
	}

	// Add the cluster port if we have more than one replica
	if *expected.Spec.Replicas > 1 {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Port: DefaultNatsClusterPort, Name: "cluster",
		})
	}

	return svc
}

func getDesiredServiceState(observed *ObservedState) *corev1.Service {
	expected := observed.buildQueue.DeepCopy()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-nats", expected.Name),
			Namespace:       expected.Namespace,
			OwnerReferences: getOwnerReferences(observed),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: 4222, Name: "nats"},
			},
			Selector: map[string]string{"buildqueue": expected.Name},
		},
	}
}

func getConfigMapState(observed *ObservedState, config *NatsConfig) *corev1.ConfigMap {
	expected := observed.buildQueue.DeepCopy()

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-nats", expected.Name),
			Namespace:       expected.Namespace,
			OwnerReferences: getOwnerReferences(observed),
		},
		Data: config.GetConfigMapData(),
	}
}

func getOwnerReferences(observed *ObservedState) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion:         observed.buildQueue.APIVersion,
			Kind:               observed.buildQueue.Kind,
			Name:               observed.buildQueue.ObjectMeta.Name,
			UID:                observed.buildQueue.ObjectMeta.UID,
			Controller:         &[]bool{true}[0],
			BlockOwnerDeletion: &[]bool{false}[0],
		},
	}
}
