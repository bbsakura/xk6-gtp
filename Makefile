#brunch name version
VERSION := $(shell git rev-parse --abbrev-ref HEAD)

build: gen
	xk6 build --with github.com/takehaya/xk6-gtp@latest=$(shell pwd)

## lint
.PHONY: lint
lint:
	@for pkg in $$(go list ./...): do \
		golint --set_exit_status $$pkg || exit $$?; \
	done

.PHONY: codecheck
codecheck:
	test -z "$(gofmt -s -l . | tee /dev/stderr)"
	go vet ./...

.PHONY: clean
clean:
	rm k6

.PHONY: gen
gen:
	go generate pkg/gtpv2/constant.go
