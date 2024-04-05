package image

import (
	"encoding/json"
	"maps"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
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
		o, _ := m.obj.(*batchv1.CronJob)
		o.Spec.JobTemplate.Spec.Template.Spec = m.manage(o.Spec.JobTemplate.Spec.Template.Spec)
	case "DaemonSet":
		o, _ := m.obj.(*appsv1.DaemonSet)
		o.Spec.Template.Spec = m.manage(o.Spec.Template.Spec)
	case "Deployment":
		o, _ := m.obj.(*appsv1.Deployment)
		o.Spec.Template.Spec = m.manage(o.Spec.Template.Spec)
	case "Job":
		o, _ := m.obj.(*batchv1.Job)
		o.Spec.Template.Spec = m.manage(o.Spec.Template.Spec)
	case "ReplicaSet":
		o, _ := m.obj.(*appsv1.ReplicaSet)
		o.Spec.Template.Spec = m.manage(o.Spec.Template.Spec)
	case "ReplicationController":
		o, _ := m.obj.(*corev1.ReplicationController)
		o.Spec.Template.Spec = m.manage(o.Spec.Template.Spec)
	case "StatefulSet":
		o, _ := m.obj.(*appsv1.StatefulSet)
		o.Spec.Template.Spec = m.manage(o.Spec.Template.Spec)
	case "Pod":
		o, _ := m.obj.(*corev1.Pod)
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
	var containers []corev1.Container
	switch {
	case len(m.include) > 0:
		containers = util.ModifyFunc(spec.Containers, m.include, func(c corev1.Container, n string) corev1.Container {
			if c.Name == n {
				c.ImagePullPolicy = corev1.PullNever
			}
			return c
		})
	case len(m.exclude) > 0:
		containers = util.ModifyFunc(spec.Containers, m.exclude, func(c corev1.Container, n string) corev1.Container {
			if c.Name != n {
				c.ImagePullPolicy = corev1.PullNever
			}
			return c
		})
	default:
		for _, c := range spec.Containers {
			c.ImagePullPolicy = corev1.PullNever
			containers = append(containers, c)
		}
	}

	spec.Containers = containers

	return spec
}

func (m *Mutator) manageSelectors(spec corev1.PodSpec) corev1.PodSpec {
	selectors := spec.NodeSelector
	if selectors == nil {
		selectors = make(map[string]string)
	}

	// Ensure we are removing any existing selectors.  There may be a quicker
	// way to do this, but the I don't expect the number of selectors to be
	// large.
	maps.DeleteFunc(selectors, func(k string, v string) bool {
		return strings.HasPrefix(k, stvziov1.LabelPrefix)
	})

	var containers []corev1.Container
	// Include will always take precedence over exclude.
	switch {
	case len(m.include) > 0:
		containers = util.FilterFunc(spec.Containers, m.include, func(c corev1.Container, n string) bool {
			return c.Name == n
		})
	case len(m.exclude) > 0:
		containers = util.FilterFunc(spec.Containers, m.exclude, func(c corev1.Container, n string) bool {
			return c.Name != n
		})
	default:
		containers = spec.Containers
	}

	for _, c := range containers {
		selectors[stvziov1.HashedImageLabelKey(c.Image)] = "available"
	}

	spec.NodeSelector = selectors

	return spec
}
