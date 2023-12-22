project_path := $(shell pwd)

GOPKG=github.com/thirdmartini/go-nvme
GOBIN=cmd/nvmed
GOBINCTL=cmd/nvmectl

all: format vet test native

.PHONY: format
format:
	echo "$(project_path)"
	go fmt $(GOPKG)/...

PHONY: native
native: format
	go build $(GOPKG)/$(GOBIN)
	go build $(GOPKG)/$(GOBINCTL)

.PHONY: test
test: format vet
	go test -v -timeout 30s $(GOPKG)/...
	#go test -timeout 30s $(GOPKG)/...

.PHONY: vet
vet: format
	go vet $(GOPKG)/...

