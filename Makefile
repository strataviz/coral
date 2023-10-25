CRD_OPTIONS ?= "crd:maxDescLen=0,generateEmbeddedObjectMeta=true"
RBAC_OPTIONS ?= "rbac:roleName=coral-role"
WEBHOOK_OPTIONS ?= "webhook"
OUTPUT_OPTIONS ?= "output:artifacts:config=config/base/crd"
ENV ?= "dev"

CONTROLLER_TOOLS_VERSION ?= v0.13.0
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

###
### Generators
###
.PHONY: codegen
codegen: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./pkg/apis/..."

.PHONY: manifests
manifests:
	$(CONTROLLER_GEN) $(CRD_OPTIONS) $(RBAC_OPTIONS) $(WEBHOOK_OPTIONS) paths="./pkg/..."

.PHONY: generate
generate: codegen manifests

###
### Build, install, run, and clean
###
.PHONY: install
install: generate
	@$(KUSTOMIZE) build config/overlays/$(ENV) | envsubst | kubectl apply -f -

.PHONY: uninstall
uninstall:
	kubectl delete -k config/overlays/$(ENV)

.PHONY: run
run:
	$(eval POD := $(shell kubectl get pods -n coral -l app=coral -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -n coral -it pod/$(POD) -- bash -c "go run main.go -zap-log-level=8 -skip-insecure-verify=true"

.PHONY: exec
exec:
	$(eval POD := $(shell kubectl get pods -n coral -l app=coral -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -n coral -it pod/$(POD) -- bash

###
### Individual dep installs were copied out of kubebuilder testdata makefiles.
###
.PHONY: deps
deps: controller-gen kustomize

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: kustomize
kustomize: $(KUSTOMIZE)
$(KUSTOMIZE): $(LOCALBIN)
	@curl -sSLo ./scripts/install_kustomize.sh "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
	@chmod +x ./scripts/install_kustomize.sh
	@./scripts/install_kustomize.sh $(KUSTOMIZE_VERSION) $(LOCALBIN)

.PHONY: clean
clean:
	@kubectl delete -k config/overlays/$(ENV)
	@rm -f $(LOCALBIN)/*
