# makefile to cross compile the printmaps commandline client
# list all cross compile possibilities: go tool dist list
# 
# makefile adapted from this example:
# http://stackoverflow.com/documentation/go/1020/cross-compilation#t=201703051136361578518
#
# version 1.0.0 - 2017/06/02: initial release
# version 1.1.0 - 2018/12/01: app renamed
# version 1.2.0 - 2018/12/10: refactoring (data.go removed)
# version 1.3.0 - 2021/06/12: darwin arm64 added

appname := printmaps
sources := $(wildcard *.go)

build = GOOS=$(1) GOARCH=$(2) go build -o build/$(appname)$(3)
tar = cd build && tar -cvzf $(appname)_$(1)_$(2).tar.gz $(appname)$(3) && rm $(appname)$(3)
zip = cd build && zip $(appname)_$(1)_$(2).zip $(appname)$(3) && rm $(appname)$(3)

# .PHONY: all windows darwin linux clean

all: linux darwin windows 

clean:
	rm -rf build/

# ----- linux builds -----
linux: build/$(appname)_linux_arm.tar.gz build/$(appname)_linux_arm64.tar.gz build/$(appname)_linux_386.tar.gz build/$(appname)_linux_amd64.tar.gz

build/$(appname)_linux_386.tar.gz: $(sources)
	$(call build,linux,386,)
	$(call tar,linux,386)

build/$(appname)_linux_amd64.tar.gz: $(sources)
	$(call build,linux,amd64,)
	$(call tar,linux,amd64)

build/$(appname)_linux_arm.tar.gz: $(sources)
	$(call build,linux,arm,)
	$(call tar,linux,arm)

build/$(appname)_linux_arm64.tar.gz: $(sources)
	$(call build,linux,arm64,)
	$(call tar,linux,arm64)

# ----- darwin (macOS) builds -----
darwin: build/$(appname)_darwin_amd64.tar.gz build/$(appname)_darwin_arm64.tar.gz

build/$(appname)_darwin_amd64.tar.gz: $(sources)
	$(call build,darwin,amd64,)
	$(call tar,darwin,amd64)

build/$(appname)_darwin_arm64.tar.gz: $(sources)
	$(call build,darwin,arm64,)
	$(call tar,darwin,arm64)

# ----- windows builds -----
windows: build/$(appname)_windows_386.zip  build/$(appname)_windows_amd64.zip

build/$(appname)_windows_386.zip: $(sources)
	$(call build,windows,386,.exe)
	$(call zip,windows,386,.exe)

build/$(appname)_windows_amd64.zip: $(sources)
	$(call build,windows,amd64,.exe)
	$(call zip,windows,amd64,.exe)
