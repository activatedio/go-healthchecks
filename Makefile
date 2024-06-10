.PHONY: test goland release

VERSION ?= $(shell cat ./.version)

clean:
	go clean -testcache

test:
	go test ./...

goland:
	nix-shell goland.nix

release:
	go install github.com/activatedio/go-release@v0.0.9
	go-release perform
	git push origin
	git push origin --tags
	GOPROXY=proxy.golang.org go list -m github.com/activatedio/go-healthchecks@$(VERSION)
