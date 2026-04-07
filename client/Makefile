default: client

GIT_DESCRIBE != git describe --always
GIT_BRANCH != git branch --show-current
GIT_COMMIT != git rev-parse --short=7 HEAD

GOOS ?= 
GOARCH ?= amd64

GOTAGS ?=

GOTAGS ::= nolibopusfile,$(GOTAGS)

CC := gcc
STRIP := strip

CFLAGS ?=
LDFLAGS ?=

CFLAGS += -g -O3 -DNDEBUG

LDFLAGS += -X 'main.GitDescribe=$(GIT_DESCRIBE)'
LDFLAGS += -X 'main.GitBranch=$(GIT_BRANCH)'
LDFLAGS += -X 'main.GitCommit=$(GIT_COMMIT)'

### Find HOST_OS

ifeq ($(OS),Windows_NT)
  HOST_OS ?= Windows
else
  HOST_OS ?= $(shell uname -s 2>/dev/null || echo Unknown)
  ifneq (,$(findstring MINGW,HOST_OS))
    HOST_OS := Windows
  endif
endif

ifeq ($(HOST_OS),Windows)
  GOOS ?= windows
  GOARCH ?= amd64
endif

ifeq ($(HOST_OS),Darwin)
  GOOS ?= darwin
  ifeq ($(GOOS),darwin)
	  GOARCH ?= arm64
  endif
  ifndef BREW_PREFIX
    BREW_PREFIX := $(shell brew --prefix)
  endif
endif

ifeq ($(HOST_OS),Linux)
  machine = $(shell sh -c 'uname -m 2>/dev/null || echo unknown')
  GOOS ?= linux
  ifneq (,$(findstring aarch64,$(machine)))
    GOARCH ?= arm64
  endif

  ifeq ($(GOOS),windows)
	  CROSS ?= x86_64-w64-mingw32-
  endif
endif

### Output file name

OUTDIR ?= out
OUTFILE ?= coop-voicechat

ifeq ($(GOOS),windows)
  OUTFILE ::= $(OUTFILE).exe
endif

OUTPATH := $(OUTDIR)/$(OUTFILE)

### Build

client:
	mkdir -p $(OUTDIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=1 CGO_CFLAGS="$(CFLAGS)" CC=$(CROSS)$(CC) go build -v -o=$(OUTPATH) -ldflags="$(LDFLAGS)" -tags=$(GOTAGS)
	$(CROSS)$(STRIP) -s $(OUTPATH)

clean:
	rm -rf $(OUTDIR)
	go clean