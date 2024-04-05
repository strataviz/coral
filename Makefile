ENV ?= "dev"
CONTROLLER_TOOLS_VERSION ?= v0.14.0
ENVTEST_VERSION ?= latest
GOLANGCI_LINT_VERSION ?= v1.57.2
KUSTOMIZE_VERSION ?= latest

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS ?= "-s -w -X main.Version=$(VERSION)"

ENVTEST ?= $(LOCALBIN)/setup-envtest
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: build
build:
	@CGO_ENABLED=0 go build -trimpath --ldflags "-s -w -X build.Version=$(VERSION)" -o bin/coral

.PHONY: staging-deploy
staging-deploy:
	@docker build -f Dockerfile . --tag coral:staging
	@kind load docker-image coral:staging --name coral
	@$(KUBECTL) apply -k config/overlays/stage

.PHONY: staging-logs
staging-logs:
	@$(KUBECTL) logs -n coral -l app=coral -f --ignore-errors

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
deps: controller-gen kustomize golangci-lint

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

.PHONY: agent-run
agent-run:
	@$(KUBECTL) apply -k config/overlays/$(ENV)

.PHONY: agent-restart
agent-restart:
	@$(KUBECTL) rollout restart daemonset coral-agent -n coral
	@$(KUBECTL) rollout status daemonset coral-agent -n coral --timeout=120s	

.PHONY: agent-stop
agent-stop:
	@$(KUBECTL) delete ds/coral-agent -n coral

.PHONY: agent-logs
agent-logs:
	@$(KUBECTL) logs -n coral -l app=coral-agent -c agent -f --ignore-errors

.PHONY: exec
exec:
	$(eval POD := $(shell kubectl get pods -n coral -l app=coral -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -n coral -it pod/$(POD) -- bash
