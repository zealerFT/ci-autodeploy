.PHONY: ci-test
ci-test:
	@go test ./... -coverprofile .cover.txt
	@go tool cover -func .cover.txt
	@rm .cover.txt

.PHONY: ci-build
ci-build:
	rm -rf bin
	mkdir -p bin
	CGO_ENABLED=0 go build -buildvcs=false -o bin/autodeploy