ENV ?= "dev"

CONTROLLER_TOOLS_VERSION ?= v0.13.0
ENVTEST_VERSION ?= latest
GOLANGCI_LINT_VERSION ?= v1.54.2
KUSTOMIZE_VERSION ?= latest

ENVTEST ?= $(LOCALBIN)/setup-envtest
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

###
### Generators
###
CRD_OPTIONS ?= "crd:maxDescLen=0,generateEmbeddedObjectMeta=true"
RBAC_OPTIONS ?= "rbac:roleName=coral-role"
WEBHOOK_OPTIONS ?= "webhook"
OUTPUT_OPTIONS ?= "output:artifacts:config=config/base/crd"

.PHONY: codegen
codegen: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./pkg/apis/..."

.PHONY: manifests
manifests:
	$(CONTROLLER_GEN) $(CRD_OPTIONS) $(RBAC_OPTIONS) $(WEBHOOK_OPTIONS) paths="./pkg/..."

.PHONY: generate
generate: codegen manifests

###
### Build, install, test, and clean
###
.PHONY: install
install: deps generate
	@$(KUSTOMIZE) build config/overlays/$(ENV) | envsubst | kubectl apply -f -

.PHONY: uninstall
uninstall:
	kubectl delete -k config/overlays/$(ENV)

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: vet
vet:
	@go vet ./...

.PHONY: test
test: deps
	@KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
	go test ./pkg/... -test.v

.PHONY: lint
lint: golangci-lint
	@$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint
	@$(GOLANGCI_LINT) run --fix

.PHONY: clean
clean: depsclean
	@-kubectl delete -k config/overlays/$(ENV)
	@-kind delete cluster --name coral

###
### Individual dep installs were copied out of kubebuilder testdata makefiles.
###
.PHONY: deps
deps: controller-gen kustomize envtest golangci-lint

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN): $(LOCALBIN)
	@GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: kustomize
kustomize: $(KUSTOMIZE)
$(KUSTOMIZE): $(LOCALBIN)
	@GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

.PHONY: envtest
envtest: $(ENVTEST)
$(ENVTEST): $(LOCALBIN)
	@GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@$(ENVTEST_VERSION)
	$(LOCALBIN)/setup-envtest use --bin-dir $(LOCALBIN) 

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(LOCALBIN)
	@GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}

.PHONY: depsclean
depsclean:
# Fix permissions on the envtest assets, otherwise they can't be deleted.
	@-chmod -R +w $(LOCALBIN)/k8s/*
	@-rm -rf $(LOCALBIN)

###
### Local development
###
.PHONY: localdev
localdev:
	@./scripts/kind-start.sh
	@$(KUBECTL) apply -k config/cert-manager
	@$(KUBECTL) wait --for=condition=available --timeout=120s deploy -l app.kubernetes.io/group=cert-manager -n cert-manager
	@$(KUBECTL) apply -k config/overlays/$(ENV)

.PHONY: run
run:
	$(eval POD := $(shell kubectl get pods -n coral -l app=coral -o=custom-columns=:metadata.name --no-headers))
	@$(KUBECTL) exec -n coral -it pod/$(POD) -- bash -c "go run main.go controller --log-level=8 --skip-insecure-verify"

.PHONY: run-worker
run-worker:
	@$(KUBECTL) apply -k config/overlays/$(ENV)

.PHONY: restart-workers
restart-worker:
	@$(KUBECTL) rollout restart daemonset coral-worker -n coral
	@$(KUBECTL) rollout status daemonset coral-worker -n coral --timeout=120s	

.PHONY: stop-workers
stop-workers:
	@$(KUBECTL) delete ds/coral-worker -n coral

.PHONY: worker-logs
worker-logs:
	@$(KUBECTL) logs -n coral -l app=coral-worker -c worker -f --ignore-errors

.PHONY: exec
exec:
	$(eval POD := $(shell kubectl get pods -n coral -l app=coral -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -n coral -it pod/$(POD) -- bash
