GO       ?= go
TOOL_MOD ?= -modfile tool.go.mod
TOOL     ?= $(GO) tool $(TOOL_MOD)

-include Makefile.local

all: test

.PHONY: build install test docs-dev
tools-install:
	test -f tool.go.mod || head -n3 go.mod >> tool.go.mod
	$(GO) get $(TOOL_MOD) -tool github.com/mgechev/revive@latest
	$(GO) get $(TOOL_MOD) -tool github.com/segmentio/golines@latest
	$(GO) get $(TOOL_MOD) -tool mvdan.cc/gofumpt@latest
	$(GO) mod tidy $(TOOL_MOD)

format:
	find -iname '*.go' -print0 | xargs -0 $(TOOL) golines --max-len 80 -w --shorten-comments
	find -iname '*.go' -print0 | xargs -0 $(TOOL) gofumpt -w

test:
	$(GO) test -v .
	$(TOOL) revive -config revive.toml -formatter friendly ./...

lint:
	$(TOOL) revive -config revive.toml -formatter friendly ./...
