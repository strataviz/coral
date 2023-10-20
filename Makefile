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
install: generate tls
	kubectl apply -k config/overlays/$(ENV)

.PHONY: uninstall
uninstall:
	kubectl delete -k config/overlays/$(ENV)

.PHONY: tls
tls:
ifeq ($(ENV), "dev")
	@./scripts/gen-certs.sh
endif

.PHONY: run
run:
	$(eval POD := $(shell kubectl get pods -n coral -l app=coral -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -n coral -it pod/$(POD) -- bash -c "go run main.go -zap-log-level=8"

.PHONY: exec
exec:
	$(eval POD := $(shell kubectl get pods -n coral -l app=coral -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -n coral -it pod/$(POD) -- bash

.PHONY: clean
clean: kind-clean
	@rm -f $(LOCALBIN)/*

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
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi

###
### Local development environment
###
.PHONY: dev
dev: kind-start kind-load install

.PHONY: kind-start
kind-start:
	@./scripts/kind-start.sh

.PHONY: kind-clean
kind-clean:
	@kind delete cluster --name=strataviz
