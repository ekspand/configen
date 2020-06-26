include .project/gomod-project.mk
export GO111MODULE=on
BUILD_FLAGS=-mod=vendor
COVERAGE_EXCLUSIONS="/rt\.go|/bindata\.go|_test.go|_mock.go|templates.go"

# don't echo execution
.SILENT:

.DEFAULT_GOAL := help

.PHONY: *

default: help

all: clean tools generate build covtest

#
# clean produced files
#
clean:
	go clean
	rm -rf \
		${COVPATH} \
		${PROJ_BIN}

tools:
	go install golang.org/x/tools/cmd/stringer
	go install golang.org/x/tools/cmd/gorename
	go install golang.org/x/tools/cmd/godoc
	go install golang.org/x/tools/cmd/guru
	go install golang.org/x/lint/golint
	go install github.com/omeid/go-resources/cmd/resources
	go install github.com/go-phorce/cov-report/cmd/cov-report
	go install github.com/mattn/goveralls
	go install github.com/stretchr/testify

version:
	gofmt -r '"GIT_VERSION" -> "$(GIT_VERSION)"' version/current.template > version/current.go

build:
	echo "Building ${PROJ_NAME} with ${BUILD_FLAGS}"
	go build ${BUILD_FLAGS} -o ${PROJ_BIN}/${PROJ_NAME} ./cmd/${PROJ_NAME}
