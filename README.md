# configen

[![Build Status](https://travis-ci.com/go-phorce/configen.svg?branch=master)](https://travis-ci.com/go-phorce/configen)
[![Coverage Status](https://coveralls.io/repos/github/go-phorce/configen/badge.svg?branch=master)](https://coveralls.io/github/go-phorce/configen?branch=master)

Tool to generate a configuration system, define your config types, and have everything code generated from it. Supports parameter overrides based on hostname, or from an environment variable.

## Usage

    go install github.com/go-phorce/configen/cmd/configen

The tool can be used directly

    configen -c <config_def.json> [-d outdirDir]

Or as part of `go generate`

    go generate ./...

## Dependencies

    go get github.com/juju/errors

## Contribution

* `make all` complete build and test
* `make build` build the executable tool
* `make test` run the tests
* `make testshort` runs the tests skipping the end-to-end tests and the code coverage reporting
* `make covtest` runs the tests with end-to-end and the code coverage reporting
* `make coverage` view the code coverage results from the last make test run.
* `make generate` runs go generate to update any code generated files
* `make fmt` runs go fmt on the project.
* `make lint` runs the go linter on the project.

run `make all` once, then run `make build` or `make test` as needed.

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
