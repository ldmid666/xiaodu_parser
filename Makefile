VERSION = v0.0.2
PROJECT_NAME := parser_xiaodu
OUTPUT_DIR := $(CURDIR)
BUILDTIME = $(shell date "+%Y%m%d")
LDFLAGS = -mod=vendor -ldflags "-w -s -X 'main.Version=${VERSION}' -X 'main.Build=${BUILDTIME}'"
# VERSION = $(shell git describe --tags `git rev-list --tags --max-count=1`)_$(shell git rev-parse --short HEAD)
GITVERSION = $(shell git rev-parse --short HEAD)
export GO111MODULE := on
export GOPROXY := https://gocenter.io

target ?= mac

ifeq ($(target), mac)
    ENV := env GOOS=darwin GOARCH=amd64
else ifeq ($(target), linux)
    ENV := env GOOS=linux GOARCH=amd64
else ifeq ($(target), arm)
    ENV := env GOOS=linux GOARCH=arm
else
$(error Invalid target)
endif
GOVER := $(shell go version)
# NEEDGOVER := 1.11.
# VER := $(findstring $(NEEDGOVER),$(GOVER))
# ifeq ($(VER),)
#     $(error need go version is not set)
# endif
$(info GOVER: $(GOVER))
$(info target: $(target))
$(info VERSION: $(VERSION))
$(info )

output:
# @表示不显示这行命令，但是还是会显示结果
	@echo "Compiling source"
	$(ENV) go build $(LDFLAGS)  -o $(OUTPUT_DIR)/$(PROJECT_NAME)

run:
	./${PROJECT_NAME}

clean:
	rm -f $(OUTPUT_DIR)/$(PROJECT_NAME)*

release:
	git checkout release
	git merge develop
	git push origin release
	git checkout develop
mod:
	rm -rf $(CURDIR)/vendor $(CURDIR)/go.mod $(CURDIR)/go.sum
	go mod init $(PROJECT_NAME)
	go mod tidy
	go mod vendor
	