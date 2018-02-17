.PHONY: test build build-no-modules modules

all: test build

test:
	go test ./...

build:
	go build -o bin/yobot main.go

build-no-modules:
	CGO_ENABLED=0 go build -o bin/yobot main.go

modules:
	@echo "Building modules"
	@rm -f ./modules/*.so
	@for m in ./modules/*; do \
		cd "$$m"; \
		echo "Building $$(basename $$m).so"; \
		go build -buildmode=plugin -o "../$$(basename $$m).so" .; \
		cd ../..; \
	done
