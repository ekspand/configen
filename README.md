# configen

Tool to generate a configuration system, define your config types, and have everything code gen'd from it. Supports parameter overrides based on hostname, or from an environment variable.
Expecting GOPATH location is github.com/ekspand/configen

## Usage

	configen -c <config_def.json> [-d outdirDir]

## Build

    go build .

## Test

    go vet .
    go test .

