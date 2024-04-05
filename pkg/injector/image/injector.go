// Copyright 2023 StrataViz
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package image

// +kubebuilder:docs-gen:collapse=Apache License

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:docs-gen:collapse=Go imports

// +kubebuilder:webhook:verbs=create;update,path=/mutate-stvz-io-v1-image-injector,mutating=true,failurePolicy=fail,groups=apps,resources=cronjobs;daemonsets;deployments;jobs;replicasets;replicationcontrollers;statefulsets,versions=v1,name=minjector.image.stvz.io,admissionReviewVersions=v1,sideEffects=none

type Injector struct {
	client.Client
	cache   cache.Cache
	decoder *admission.Decoder
	log     logr.Logger

	// default webhook action as config value
	defaultAction admission.Response
}

// SetupWebhookWithManager adds webhook for BuildSet.
func SetupWebhookWithManager(mgr ctrl.Manager) error {
	i := &Injector{
		Client:        mgr.GetClient(),
		cache:         mgr.GetCache(),
		decoder:       admission.NewDecoder(mgr.GetScheme()),
		defaultAction: admission.Allowed(""),
		log:           mgr.GetLogger().WithName("image-injector"),
	}

	mgr.GetWebhookServer().Register("/mutate-stvz-io-v1-image-injector", &webhook.Admission{
		Handler: i,
	})

	return nil
}

func (i *Injector) Handle(ctx context.Context, req admission.Request) admission.Response {
	// TODO: change back to filtering based on the annotations.  We can do that here now
	// instead of forcing us to use labels on the resources.

	logger := log.FromContext(ctx)
	logger.Info("handling request", "req", req)

	mutator := NewMutator(i.log)
	if err := mutator.FromReq(req, i.decoder); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// If we are not managing the object, then we should just allow it through
	if !mutator.Managed() {
		admission.Allowed("")
	}

	// Run the mutators
	return mutator.Mutate(req)
}

var _ admission.Handler = &Injector{}
