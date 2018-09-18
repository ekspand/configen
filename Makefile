include .project/go-project.mk

# don't echo execution
.SILENT:

.DEFAULT_GOAL := help

.PHONY: *

default: help

all: clean gopath tools generate build covtest

gettools:
	mkdir -p ${TOOLS_PATH}/src
	$(call httpsclone,${GITHUB_HOST},golang/tools,             ${TOOLS_PATH}/src/golang.org/x/tools,                  release-branch.go1.11)
	$(call httpsclone,${GITHUB_HOST},jteeuwen/go-bindata,      ${TOOLS_PATH}/src/github.com/jteeuwen/go-bindata,      6025e8de665b31fa74ab1a66f2cddd8c0abf887e)
	$(call httpsclone,${GITHUB_HOST},jstemmer/go-junit-report, ${TOOLS_PATH}/src/github.com/jstemmer/go-junit-report, 385fac0ced9acaae6dc5b39144194008ded00697)
	$(call httpsclone,${GITHUB_HOST},go-phorce/cov-report,     ${TOOLS_PATH}/src/github.com/go-phorce/cov-report,     master)
	$(call httpsclone,${GITHUB_HOST},golang/lint,              ${TOOLS_PATH}/src/golang.org/x/lint,                   06c8688daad7faa9da5a0c2f163a3d14aac986ca)
	#$(call httpsclone,${GITHUB_HOST},golangci/golangci-lint,   ${TOOLS_PATH}/src/github.com/golangci/golangci-lint,   master)
	$(call httpsclone,${GITHUB_HOST},mattn/goveralls,          ${TOOLS_PATH}/src/github.com/mattn/goveralls,          88fc0d50edb2e4cf09fe772457b17d6981826cff)
	$(call httpsclone,${GITHUB_HOST},stretchr/testify,         ${GOPATH}/src/github.com/stretchr/testify,             master)

tools: gettools
	GOPATH=${TOOLS_PATH} go install golang.org/x/tools/cmd/stringer
	GOPATH=${TOOLS_PATH} go install golang.org/x/tools/cmd/gorename
	GOPATH=${TOOLS_PATH} go install golang.org/x/tools/cmd/godoc
	GOPATH=${TOOLS_PATH} go install golang.org/x/tools/cmd/guru
	GOPATH=${TOOLS_PATH} go install github.com/jteeuwen/go-bindata/...
	GOPATH=${TOOLS_PATH} go install github.com/jstemmer/go-junit-report
	GOPATH=${TOOLS_PATH} go install github.com/go-phorce/cov-report/cmd/cov-report
	GOPATH=${TOOLS_PATH} go install golang.org/x/lint/golint
	#GOPATH=${TOOLS_PATH} go install github.com/golangci/golangci-lint/cmd/golangci-lint
	GOPATH=${TOOLS_PATH} go install github.com/mattn/goveralls

version:
	gofmt -r '"GIT_VERSION" -> "$(GIT_VERSION)"' version/current.template > version/current.go

build:
	echo "Building ${PROJ_NAME}"
	cd ${TEST_DIR} && go build -o ${PROJ_BIN}/${PROJ_NAME} ./cmd/${PROJ_NAME}
	cp ${PROJ_BIN}/${PROJ_NAME} ${TOOLS_BIN}/
