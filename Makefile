include .project/go-project.mk

# don't echo execution
.SILENT:

.DEFAULT_GOAL := help

.PHONY: *

default: help

all: clean gopath tools generate build test

gettools:
	mkdir -p ${TOOLS_SRC}
	$(call gitclone,${GITHUB_HOST},golang/tools,             ${TOOLS_SRC}/golang.org/x/tools,                  2226533658007779ffd629b495a088530c84dc50)
	$(call gitclone,${GITHUB_HOST},golang/lint,              ${TOOLS_SRC}/github.com/golang/lint,              3ea3fa98a8104b2c8f8a7bffaebc7e54dddf99e1)
	$(call gitclone,${GITHUB_HOST},jteeuwen/go-bindata,      ${TOOLS_SRC}/github.com/jteeuwen/go-bindata,      v3.0.7)
	$(call gitclone,${GITHUB_HOST},jstemmer/go-junit-report, ${TOOLS_SRC}/github.com/jstemmer/go-junit-report, 385fac0ced9acaae6dc5b39144194008ded00697)
	$(call gitclone,${GITHUB_HOST},go-phorce/cov-report,     ${TOOLS_SRC}/github.com/go-phorce/cov-report,     master)

tools: gettools
	GOPATH=${TOOLS_PATH} go install golang.org/x/tools/cmd/stringer
	GOPATH=${TOOLS_PATH} go install golang.org/x/tools/cmd/gorename
	GOPATH=${TOOLS_PATH} go install golang.org/x/tools/cmd/godoc
	GOPATH=${TOOLS_PATH} go install golang.org/x/tools/cmd/guru
	GOPATH=${TOOLS_PATH} go install github.com/golang/lint/golint
	GOPATH=${TOOLS_PATH} go install github.com/jteeuwen/go-bindata/...
	GOPATH=${TOOLS_PATH} go install github.com/jstemmer/go-junit-report
	GOPATH=${TOOLS_PATH} go install github.com/go-phorce/cov-report/cmd/cov-report

getdevtools:
	$(call gitclone,${GITHUB_HOST},golang/tools,                ${GOPATH}/src/golang.org/x/tools,                  master)
	$(call gitclone,${GITHUB_HOST},derekparker/delve,           ${GOPATH}/src/github.com/derekparker/delve,        master)
	$(call gitclone,${GITHUB_HOST},uudashr/gopkgs,              ${GOPATH}/src/github.com/uudashr/gopkgs,           master)
	$(call gitclone,${GITHUB_HOST},nsf/gocode,                  ${GOPATH}/src/github.com/nsf/gocode,               master)
	$(call gitclone,${GITHUB_HOST},rogpeppe/godef,              ${GOPATH}/src/github.com/rogpeppe/godef,           master)
	$(call gitclone,${GITHUB_HOST},acroca/go-symbols,           ${GOPATH}/src/github.com/acroca/go-symbols,        master)
	$(call gitclone,${GITHUB_HOST},ramya-rao-a/go-outline,      ${GOPATH}/src/github.com/ramya-rao-a/go-outline,   master)
	$(call gitclone,${GITHUB_HOST},ddollar/foreman,             ${GOPATH}/src/github.com/ddollar/foreman,          master)
	$(call gitclone,${GITHUB_HOST},sqs/goreturns,               ${GOPATH}/src/github.com/sqs/goreturns,            master)
	$(call gitclone,${GITHUB_HOST},karrick/godirwalk,           ${GOPATH}/src/github.com/karrick/godirwalk,        master)
	$(call gitclone,${GITHUB_HOST},pkg/errors,                  ${GOPATH}/src/github.com/pkg/errors,               master)

devtools: getdevtools
	go get -u github.com/golang/dep/cmd/dep
	go install golang.org/x/tools/cmd/fiximports
	go install golang.org/x/tools/cmd/goimports
	go install github.com/derekparker/delve/cmd/dlv
	go install github.com/uudashr/gopkgs/cmd/gopkgs
	go install github.com/nsf/gocode
	go install github.com/rogpeppe/godef
	go install github.com/acroca/go-symbols
	go install github.com/ramya-rao-a/go-outline
	go install github.com/sqs/goreturns

version:
	gofmt -r '"GIT_VERSION" -> "$(GIT_VERSION)"' version/current.template > version/current.go

build:
	echo "Building ${PROJ_NAME}"
	cd ${TEST_DIR} && go build -o ${PROJ_BIN}/${PROJ_NAME} ./cmd/${PROJ_NAME}
	cp ${PROJ_BIN}/${PROJ_NAME} ${TOOLS_BIN}/
