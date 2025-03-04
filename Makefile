IMAGE ?= "flux-suspension-exporter"
TAG ?= "dev"

# Makefile
all: setup hooks

# requires `nvm use --lts` or `nvm use node`
.PHONY: setup
setup: 
	npm install -g @commitlint/config-conventional @commitlint/cli  


.PHONY: hooks
hooks:
	@git config --local core.hooksPath .githooks/

.PHONY: kind
kind:
	kind create cluster --name kind-fse --config kind/config.yaml

.PHONY: kind-delete
kind-delete:
	kind delete cluster --name kind-fse

.PHONY: kind-load
kind-load:
	docker build . -t $(IMAGE):$(TAG)
	kind load docker-image --name kind-fse $(IMAGE):$(TAG)

.PHONY: apply
apply:
	kustomize build k8s/ | kubectl apply -f -