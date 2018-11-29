# makefile to cross compile the printmaps commandline client
# list all cross compile possibilities: go tool dist list
# 
# makefile adapted from this example:
# http://stackoverflow.com/documentation/go/1020/cross-compilation#t=201703051136361578518
#
# version 1.0.0 - 2017/06/02: initial release
# version 1.1.0 - 2018/11/29: builds linux (64 bit) only

appname := printmaps_webservice
sources := $(wildcard *.go)

build = GOOS=$(1) GOARCH=$(2) go build -o build/$(appname)$(3)
tar = cd build && tar -cvzf $(appname)_$(1)_$(2).tar.gz $(appname)$(3) && rm $(appname)$(3)
zip = cd build && zip $(appname)_$(1)_$(2).zip $(appname)$(3) && rm $(appname)$(3)

# .PHONY: all linux clean

all: data.go linux

clean:
	rm -rf build/

# ----- copy shared code -----
data.go: ../data.go
	cp ../data.go data.go

# ----- linux builds -----
linux: build/$(appname)_linux_arm64.tar.gz build/$(appname)_linux_amd64.tar.gz

build/$(appname)_linux_amd64.tar.gz: $(sources)
	$(call build,linux,amd64,)
	$(call tar,linux,amd64)

build/$(appname)_linux_arm64.tar.gz: $(sources)
	$(call build,linux,arm64,)
	$(call tar,linux,arm64)
