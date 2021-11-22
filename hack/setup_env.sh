#!/usr/bin/env bash

set -o nounset
set -o pipefail
set -o errexit
set -o xtrace

function print_help() {
  ALL_FUNCS="test_env|print_help"
  echo "Usage: bash ${0} (${ALL_FUNCS})"
}

function test_env() {

    go get github.com/onsi/ginkgo/ginkgo@v1.16.4 \
        golang.org/x/tools/cmd/goimports@v0.1.5 \
        github.com/golang/mock/mockgen@v1.5.0 \
        github.com/vektra/mockery/.../@v1.1.2 \
        gotest.tools/gotestsum@v1.6.3 \
        github.com/axw/gocov/gocov \
        sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.2 \
        github.com/AlekSi/gocov-xml@v0.0.0-20190121064608-3a14fb1c4737

    go mod vendor
}

declare -F $@ || (print_help && exit 1)

"$@"
