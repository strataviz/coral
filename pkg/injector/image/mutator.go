package image

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"stvz.io/coral/pkg/util"
)

type Action struct {
	MutatePullPolicy bool
	MutateSelectors  bool
}

type Mutator struct {
	log logr.Logger

	policy    bool
	selectors bool

	include []string
	exclude []string

	kind string
	obj  client.Object
}

func NewMutator(log logr.Logger) *Mutator {
	return &Mutator{
		log: log,
	}
}

func (m *Mutator) FromReq(req admission.Request, decoder *admission.Decoder) error {
	m.kind = req.Kind.Kind

	obj, err := util.ObjectFromKind(m.kind)
	if err != nil {
		return err
	}

	err = decoder.Decode(req, obj)
	if err != nil {
		return err
	}
	m.obj = obj

	annotation, ok := obj.GetAnnotations()["image.stvz.io/inject"]
	if !ok || annotation == "" {
		return nil
	}

	if included, ok := obj.GetAnnotations()["image.stvz.io/included"]; ok {
		m.include = strings.Split(included, ",")
	} else {
		m.include = []string{}
	}

	if excluded, ok := obj.GetAnnotations()["image.stvz.io/excluded"]; ok {
		m.exclude = strings.Split(excluded, ",")
	} else {
		m.exclude = []string{}
	}

	parts := strings.Split(annotation, ",")
	for _, part := range parts {
		switch part {
		case "pull-policy":
			m.policy = true
		case "selectors":
			m.selectors = true
		}
	}

	return nil
}

func (m *Mutator) Managed() bool {
	return m.policy || m.selectors
}

func (m *Mutator) Mutate(req admission.Request) admission.Response {
	obj := m.mutate(m.obj)

	o, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, o)
}

func (m *Mutator) mutate(obj client.Object) client.Object {
	// If we are a pod or a replicaset and have a reference to an object that we should
	// already be managing, then just allow it through as we'll be updating the templates
	// in the other objects.
	if m.kind == "Pod" || m.kind == "ReplicaSet" {
		if ref := m.obj.GetOwnerReferences(); len(ref) > 0 {
			return obj
		}
	}

	// Otherwise, handle the policy
	switch m.kind {
	case "CronJob":
		o := m.obj.(*batchv1.CronJob)
		o.Spec.JobTemplate.Spec.Template.Spec = m.manage(o.Spec.JobTemplate.Spec.Template.Spec)
	case "DaemonSet":
		o := m.obj.(*appsv1.DaemonSet)
		o.Spec.Template.Spec = m.manage(o.Spec.Template.Spec)
	case "Deployment":
		o := m.obj.(*appsv1.Deployment)
		o.Spec.Template.Spec = m.manage(o.Spec.Template.Spec)
	case "Job":
		o := m.obj.(*batchv1.Job)
		o.Spec.Template.Spec = m.manage(o.Spec.Template.Spec)
	case "ReplicaSet":
		o := m.obj.(*appsv1.ReplicaSet)
		o.Spec.Template.Spec = m.manage(o.Spec.Template.Spec)
	case "ReplicationController":
		o := m.obj.(*corev1.ReplicationController)
		o.Spec.Template.Spec = m.manage(o.Spec.Template.Spec)
	case "StatefulSet":
		o := m.obj.(*appsv1.StatefulSet)
		o.Spec.Template.Spec = m.manage(o.Spec.Template.Spec)
	case "Pod":
		o := m.obj.(*corev1.Pod)
		o.Spec = m.manage(o.Spec)
	}

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations["image.stvz.io/injected"] = "true"
	obj.SetAnnotations(annotations)

	return obj
}

func (m *Mutator) manage(spec corev1.PodSpec) corev1.PodSpec {
	if m.policy {
		spec = m.manageImagePullPolicy(spec)
	}

	if m.selectors {
		spec = m.manageSelectors(spec)
	}

	return spec
}

func (m *Mutator) manageImagePullPolicy(spec corev1.PodSpec) corev1.PodSpec {
	var containerMap map[string]bool

	// see if we can pull out the include/exclude.
	if len(m.include) > 0 {
		containerMap = util.ContainerNamesMapInclude(spec.Containers, m.include)
	} else if len(m.exclude) > 0 {
		containerMap = util.ContainerNamesMapExclude(spec.Containers, m.exclude)
	} else {
		containerMap = util.ContainerNamesMap(spec.Containers)
	}

	containers := []corev1.Container{}
	for _, container := range spec.Containers {
		if containerMap[container.Name] {
			container.ImagePullPolicy = corev1.PullNever
		}
		containers = append(containers, container)
	}

	spec.Containers = containers

	return spec
}

func (m *Mutator) manageSelectors(spec corev1.PodSpec) corev1.PodSpec {
	containers := spec.Containers

	selectors := spec.NodeSelector
	if selectors == nil {
		selectors = make(map[string]string)
	}

	// see if we can pull out the include/exclude.
	var images []string
	if len(m.include) > 0 {
		images = util.ContainerImageInclude(containers, m.include)
	} else if len(m.exclude) > 0 {
		images = util.ContainerImageExclude(containers, m.exclude)
	} else {
		images = util.ContainerImages(containers)
	}

	for _, image := range images {
		selectors[util.ImageLabelKey(util.ImageHasher(image))] = "available"
	}

	spec.NodeSelector = selectors

	return spec
}
