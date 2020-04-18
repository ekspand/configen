include .project/go-project.mk
export GO111MODULE=off

# don't echo execution
.SILENT:

.DEFAULT_GOAL := help

.PHONY: *

default: help

all: clean gopath tools generate build covtest

tools:
	echo ${TOOLS_SRC}
	mkdir -p ${TOOLS_PATH}
	@if [ ! -L ${TOOLS_SRC} ]; then ln -sf ${VENDOR_SRC} ${TOOLS_SRC}; fi
	GOPATH=${TOOLS_PATH} go install golang.org/x/tools/cmd/stringer
	GOPATH=${TOOLS_PATH} go install golang.org/x/tools/cmd/gorename
	GOPATH=${TOOLS_PATH} go install golang.org/x/tools/cmd/godoc
	GOPATH=${TOOLS_PATH} go install golang.org/x/tools/cmd/guru
	GOPATH=${TOOLS_PATH} go install golang.org/x/lint/golint
	GOPATH=${TOOLS_PATH} go install github.com/jteeuwen/go-bindata/...
	GOPATH=${TOOLS_PATH} go install github.com/go-phorce/cov-report/cmd/cov-report
	GOPATH=${TOOLS_PATH} go install github.com/mattn/goveralls
	$(call httpsclone,${GITHUB_HOST},stretchr/testify,   ${GOPATH}/src/github.com/stretchr/testify,v1.2.2)

version:
	gofmt -r '"GIT_VERSION" -> "$(GIT_VERSION)"' version/current.template > version/current.go

build:
	echo "Building ${PROJ_NAME}"
	cd ${TEST_DIR} && go build -o ${PROJ_BIN}/${PROJ_NAME} ./cmd/${PROJ_NAME}
	cp ${PROJ_BIN}/${PROJ_NAME} ${TOOLS_BIN}/
