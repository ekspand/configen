# configen

Tool to generate a configuration system, define your config types, and have everything code generated from it. Supports parameter overrides based on hostname, or from an environment variable.
Expecting GOPATH location is github.com/go-phorce/configen

## Usage

    go install github.com/go-phorce/configen/cmd/configen

The tool can be used directly

    configen -c <config_def.json> [-d outdirDir]

Or as part of `go generate`

    go generate ./...

## Build

Before openning VSCODE or running make, run once:
    . ./vscode.sh

* `make get` fetches the pinned dependencies from repos
* `make devtools` get the dev tools for local development in VSCODE
* `make build` build the executable tool
* `make test` run the tests
* `make testshort` runs the tests skipping the end-to-end tests and the code coverage reporting
* `make covtest` runs the tests with end-to-end and the code coverage reporting
* `make coverage` view the code coverage results from the last make test run.
* `make generate` runs go generate to update any code generated files
* `make fmt` runs go fmt on the project.
* `make lint` runs the go linter on the project.

run `make get` once, then run `make build` or `make test` as needed.

First run:

    make all

Subsequent builds:

    make build

Tests:

    make test

Optionally run golang race detector with test targets by setting RACE flag:

    make test RACE=true

Review coverage report:

    make covtest coverage

### Current Travis-CI build status

[![Build Status](https://travis-ci.org/go-phorce/configen.svg?branch=master)](https://travis-ci.org/go-phorce/configen)