BIN_DIR    := $(CURDIR)/bin
BUILD_DIR  := $(CURDIR)/build
TOOLS_DIR  := $(CURDIR)/tools
SCRIPT_DIR := $(CURDIR)/scripts

PROTO_DIR := proto
PROTO_FILES = $(shell find $(PROTO_DIR)/ -name "*.proto" -print)
GENERATED_SOURCES = $(patsubst %.proto,%.pb.go,$(PROTO_FILES)) $(patsubst %.proto,%_grpc.pb.go,$(PROTO_FILES))
SOURCES = $(shell find . -name "tools" -prune -o \( -name "*.go" \! -name "*.pb.go" -o -name "go.mod" \) -print)

PROTOC          := $(BIN_DIR)/protoc
PROTOC_VERSION  := 21.4
PROTOC_CHECKSUM := d51e8f030162f08823a4738ab0ac00bee537e30b583a562e6962dbb040d86736

PROTOC_GEN_GO         := $(BIN_DIR)/protoc-gen-go
PROTOC_GEN_GO_VERSION := v1.28.1
PROTOC_GEN_GO_PACKAGE := google.golang.org/protobuf/cmd/protoc-gen-go

PROTOC_GEN_GO_GRPC         := $(BIN_DIR)/protoc-gen-go-grpc
PROTOC_GEN_GO_GRPC_VERSION := v1.2.0
PROTOC_GEN_GO_GRPC_PACKAGE := google.golang.org/grpc/cmd/protoc-gen-go-grpc

GOFLAGS = -trimpath -mod=readonly -ldflags '-extldflags=-static'
GOENV = CGO_ENABLED=0

KIND         := $(BIN_DIR)/kind
KIND_PACKAGE := sigs.k8s.io/kind
KIND_VERSION := v0.17.0
KIND_CLUSTER_NAME := metropolis
# Obtained from https://github.com/kubernetes-sigs/kind/releases/tag/v0.17.0
KIND_NODE_VERSION := v1.25.3@sha256:f52781bc0d7a19fb6c405c2af83abfeb311f130707a0e219175677e366cc45d1
KIND_CLUSTER_NAME := grpctimeout

JSONNET         := $(BIN_DIR)/jsonnet
JSONNET_PACKAGE := github.com/google/go-jsonnet/cmd/jsonnet
JSONNET_VERSION := v0.18.0

KUBECTL := $(BIN_DIR)/kubectl
KUBECTL_VERSION := v1.25.3
# curl -L -s "https://dl.k8s.io/${KUBECTL_VERSION}/bin/linux/amd64/kubectl.sha256"
KUBECTL_CHECKSUM := f57e568495c377407485d3eadc27cda25310694ef4ffc480eeea81dea2b60624

SERVER_BINARY := $(BUILD_DIR)/grpcserver
CLIENT_BINARY := $(BUILD_DIR)/grpcclient

.PHONY: build
build: $(SERVER_BINARY) $(CLIENT_BINARY) | $(BUILD_DIR)

#--- template for go install ----------------
# Usage: $(eval $(call GO_INSTALL,package/path/to/command,VERSION))
define GO_INSTALL
$(BIN_DIR)/$(notdir $(1)): $(TOOLS_DIR)/$(notdir $(1))-$(2)/$(notdir $(1)) | $(BIN_DIR)
	ln -sf $$< $$@

$(TOOLS_DIR)/$(notdir $(1))-$(2): | $(TOOLS_DIR)
	mkdir -p $$@

$(TOOLS_DIR)/$(notdir $(1))-$(2)/$(notdir $(1)): $(TOOLS_DIR)/$(notdir $(1))-$(2)
	$(GOENV) GOBIN=$$< go install $(GOFLAGS) $(1)@$(2)
endef

#--- Code generation by protoc --------------------
$(TOOLS_DIR)/protoc-$(PROTOC_VERSION)/bin/protoc: | $(TOOLS_DIR)
	$(SCRIPT_DIR)/install-protoc $(PROTOC_VERSION) $(PROTOC_CHECKSUM)

$(PROTOC): $(TOOLS_DIR)/protoc-$(PROTOC_VERSION)/bin/protoc | $(BIN_DIR)
	ln -sf $< $@

$(eval $(call GO_INSTALL,$(PROTOC_GEN_GO_PACKAGE),$(PROTOC_GEN_GO_VERSION)))
$(eval $(call GO_INSTALL,$(PROTOC_GEN_GO_GRPC_PACKAGE),$(PROTOC_GEN_GO_GRPC_VERSION)))

$(PROTO_DIR)/%.pb.go: $(PROTO_DIR)/%.proto $(PROTOC) $(PROTOC_GEN_GO)
	$(PROTOC) \
		-I "$(PROTO_DIR)" \
		--plugin=protoc-gen-go=$(PROTOC_GEN_GO) \
		--go_out $(dir $@) \
		--go_opt paths=source_relative \
		"$<"

$(PROTO_DIR)/%_grpc.pb.go: $(PROTO_DIR)/%.proto $(PROTOC) $(PROTOC_GEN_GO_GRPC)
	$(PROTOC) \
		-I "$(PROTO_DIR)" \
		--plugin=protoc-gen-go-grpc=$(PROTOC_GEN_GO_GRPC) \
		--go-grpc_out $(dir $@) \
		--go-grpc_opt paths=source_relative \
		"$<"

#--- Build go binary ------------------
.PRECIOUS: $(GENERATED_SOURCES)
$(BUILD_DIR)/%: $(SOURCES) $(GENERATED_SOURCES) | $(BUILD_DIR)
	$(GOENV) go build $(GOFLAGS) -o $@ ./cmd/$*

.PHONY: image
image: $(SERVER_BINARY) $(CLIENT_BINARY)
	DOCKER_BUILDKIT=1 docker build . -t grpc-contour-timeout:latest

#--- kind cluster ---------------------------------
$(eval $(call GO_INSTALL,$(KIND_PACKAGE),$(KIND_VERSION)))
$(eval $(call GO_INSTALL,$(JSONNET_PACKAGE),$(JSONNET_VERSION)))

$(TOOLS_DIR)/kubectl-$(KUBECTL_VERSION)/kubectl: | $(TOOLS_DIR)
	$(SCRIPT_DIR)/install-kubectl $(KUBECTL_VERSION) $(KUBECTL_CHECKSUM)

$(KUBECTL): $(TOOLS_DIR)/kubectl-$(KUBECTL_VERSION)/kubectl | $(BIN_DIR)
	ln -sf $< $@

.PHONY: cluster-up
cluster-up: $(KIND) $(KUBECTL)
	$(KIND) create cluster --name $(KIND_CLUSTER_NAME) --image kindest/node:$(KIND_NODE_VERSION) --config kind-cluster.yaml --wait 180s
	if [ $$($(KUBECTL) config current-context) != "kind-$(KIND_CLUSTER_NAME)" ]; then \
		$(KUBECTL) config use-context kind-$(KIND_CLUSTER_NAME); \
	fi
	$(KUBECTL) apply -f https://projectcontour.io/quickstart/contour.yaml
	$(KUBECTL) wait --timeout=120s --for=condition=available --all deployments -n projectcontour

.PHONY: cluster-down
cluster-down: $(KIND)
	$(KIND) delete cluster --name $(KIND_CLUSTER_NAME)

.PHONY: load-image
load-image: image 
	$(KIND) load docker-image --name $(KIND_CLUSTER_NAME) grpc-contour-timeout:latest 

.PHONY: configure-context
configure-context: $(KUBECTL)
	@if [ $$($(KUBECTL) config current-context) != "kind-$(KIND_CLUSTER_NAME)" ]; then \
		$(KUBECTL) config use-context kind-$(KIND_CLUSTER_NAME); \
	fi

.PHONY: apply
apply: $(JSONNET) $(KUBECTL) configure-context
	$(JSONNET) -y -S kubernetes/main.jsonnet | $(KUBECTL) apply -f - 
	$(KUBECTL) wait --timeout=120s --for=condition=available --all deployments

.PHONY: rollout
rollout: $(KUBECTL) configure-context
	$(KUBECTL) rollout restart deploy grpcserver
	$(KUBECTL) rollout status -w deploy grpcserver

.PHONY: port-forward
port-forward: configure-context
	$(KUBECTL) -n projectcontour port-forward service/envoy 8888:80

.PHONY: manifest
manifest: $(JSONNET)
	$(JSONNET) -y -S kubernetes/main.jsonnet

#--- Misc -----------------------------------------
$(BUILD_DIR) $(TOOLS_DIR) $(BIN_DIR):
	mkdir -p "$@"

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)/*
	rm -f metropolis.tar

.PHONY: clean-all
clean-all: clean
	rm -rf $(BIN_DIR)/*
	rm -rf $(TOOLS_DIR)/*
