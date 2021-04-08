######################################################### {COPYRIGHT-TOP} ###
# Licensed Materials - Property of IBM
# 5900-AEO
#
# Copyright IBM Corp. 2020, 2021
#
# US Government Users Restricted Rights - Use, duplication, or
# disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
######################################################### {COPYRIGHT-END} ###
# Current Operator version
VERSION ?= 0.0.1
# Default bundle image tag
BUNDLE_IMG ?= controller-bundle:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,crdVersions=v1"

# The Operator Image
REGISTRY ?= us.icr.io/abp-scratchpad
OPERATOR_IMAGE_NAME := iaf-demo-cartridge
FLINK_PROCESSOR_IMAGE_NAME := iaf-demo-flink-processor
IMAGE_VERSION ?= latest

# Image URL to use all building/pushing image targets
IMG ?=${REGISTRY}/${OPERATOR_IMAGE_NAME}:${IMAGE_VERSION}

# Check if req params provided
ifndef API_KEY
$(error    ===>> REQUIRED: API_KEY not provided)
endif

ifndef GIT_TOKEN
$(error    ===>> REQUIRED: GIT_TOKEN not provided)
endif


# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	# cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

purge: manifests kustomize
	# cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG} --build-arg GIT_TOKEN=${GIT_TOKEN}

# Push the docker image
docker-push:
	docker push ${IMG}

all-docker: docker-build docker-push

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests
	operator-sdk generate kustomize manifests -q
	$(KUSTOMIZE) build config/manifests \
		| hack/images.sh search_and_replace_images_for_kustomize \
		| operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	hack/add-copyright.sh
	operator-sdk bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	@echo "\n\n\n\t======> Bundle-Build\n\n"
	@docker build -f bundle.Dockerfile -t ${REGISTRY}/${OPERATOR_IMAGE_NAME}-bundle:${IMAGE_VERSION} .

bundle-push:
	@echo "\n\n\n\t======> Bundle-Push\n\n"
	docker push ${REGISTRY}/${OPERATOR_IMAGE_NAME}-bundle:${IMAGE_VERSION}

image-registry:
	@scripts/registry-setup.sh ${API_KEY}
	@echo "\n\n\n\t======> IBM Cloud Image Registry Configured\n\n"

all-bundle: bundle bundle-build bundle-push

catalog: manifests
	$(eval CATALOG_IMG=$(REGISTRY)/$(OPERATOR_IMAGE_NAME)-catalog:$(IMAGE_VERSION))
	$(eval BUNDLE_IMG=$(REGISTRY)/$(OPERATOR_IMAGE_NAME)-bundle:$(IMAGE_VERSION))

	opm index add --bundles $(BUNDLE_IMG) --container-tool docker --tag $(CATALOG_IMG)
	docker push $(CATALOG_IMG)

build: image-registry all-docker all-bundle catalog

# Removes the iafdemo CRs, CRD, Subscription, CSV, and CatalogSource
clean-demo:
	$(eval CATALOG_IMG=$(REGISTRY)/$(OPERATOR_IMAGE_NAME)-catalog:$(IMAGE_VERSION))

	oc delete --ignore-not-found -A --all iafdemo.democartridge.ibm.com || true
	oc delete --ignore-not-found subscription iaf-demo-cartridge-operator
	oc delete --ignore-not-found csv iaf-demo-cartridge.v0.0.1
	sed s#@image_placeholder@#$(CATALOG_IMG)# config/samples/catalogsource.yaml \
		| oc delete --ignore-not-found -f -
	oc delete --ignore-not-found $(shell oc get crd -o name | grep "democartridge.ibm.com") || true

	@echo "\n\t======> Demo Cleanup Complete\n"

# Creates the demo cartridge catalog and the subscription to that operator.
# This will trigger a new installation from OLM, and will then deploy the sample IAFDemo CR.
deploy-demo:
	$(eval CATALOG_IMG=$(REGISTRY)/$(OPERATOR_IMAGE_NAME)-catalog:$(IMAGE_VERSION))

	sed s#@image_placeholder@#$(CATALOG_IMG)# config/samples/catalogsource.yaml \
		| oc apply -f -
	oc apply -f config/samples/subscription.yaml
	until oc wait pod -l app.kubernetes.io/name=iaf-demo-cartridge-operator --for=condition=Ready --timeout=300s; \
		do echo "No pod yet" ; sleep 15; done
	oc apply -f config/samples/democartridge_v1_iafdemo.yaml

	# Workaround for EventStreams RBAC. Ignore the failures: those just mean it already exists.
	oc create clusterrole iaf-event-stream-workaround --verb=get,delete,create --resource=clusterrolebindings || true
	oc create clusterrolebinding iaf-event-stream-workaround-binding --clusterrole=iaf-event-stream-workaround --serviceaccount=ibm-common-services:ibm-events-operator || true

	@echo "\n\t======> Deploy Complete\n"
