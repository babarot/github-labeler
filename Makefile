APPNAME = github-labeler
VERSION = 0.1.0
APPDIR  = $(APPNAME)_$(VERSION)_$(GOOS)_$(GOARCH)
DISTDIR = dist

.PHONY: all
all: help

.PHONY: build
build:
	go build

.PHONY: pack
pack:
	mkdir -p $(DISTDIR)/$(APPDIR)
	go build -o $(DISTDIR)/$(APPDIR)/$(APPNAME)
	tar cfvz $(DISTDIR)/$(APPDIR).tar.gz $(DISTDIR)/$(APPDIR)
	rm -rf $(DISTDIR)/$(APPDIR)

.PHONY: crossbuild
crossbuild:
	@$(MAKE) pack GOOS=windows GOARCH=amd64 SUFFIX_EXE=.exe
	@$(MAKE) pack GOOS=windows GOARCH=386   SUFFIX_EXE=.exe
	@$(MAKE) pack GOOS=linux   GOARCH=amd64
	@$(MAKE) pack GOOS=linux   GOARCH=386
	@$(MAKE) pack GOOS=darwin  GOARCH=amd64
	@$(MAKE) pack GOOS=darwin  GOARCH=386

.PHONY: release
release:
	ghr -username $$USER -repository $(APPNAME) $(VERSION) $(DISTDIR)

.PHONY: test
test:
	go test -v -parallel=4 ./...

help: ## Self-documented Makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
