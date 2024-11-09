.PHONY: build
build:
	go build -trimpath -buildmode=plugin -o .local/thumbnailizer.so .

.PHONY: build-linux
build-linux:
	docker run --name strawhouse-builder --rm -v $(PWD):/opt/ golang:1.23.1-bookworm \
		/bin/bash -c "cd /opt/ && go build -trimpath -buildmode=plugin -o .local/thumbnailizer-linux.so ."