export PATH := $(GOPATH)/bin:$(PATH)
export GO111MODULE=on

GO 			:= GOGC=off go
LDFLAGS		:= -s -w
OS_ARCHS	:=darwin:amd64 darwin:arm64 linux:amd64 linux:arm64


## Build
.PHONY: build
build:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" .

## Build All
.PHONY: build-all
build-all:
	@$(foreach n, $(OS_ARCHS),\
		os=$(shell echo "$(n)" | cut -d : -f 1);\
		arch=$(shell echo "$(n)" | cut -d : -f 2);\
		gomips=$(shell echo "$(n)" | cut -d : -f 3);\
		target_suffix=$${os}_$${arch};\
		env CGO_ENABLED=0 GOOS=$${os} GOARCH=$${arch} GOMIPS=$${gomips} go build -trimpath -ldflags "$(LDFLAGS)" -o ./release/doo_cli_$${target_suffix} .;\
	)
