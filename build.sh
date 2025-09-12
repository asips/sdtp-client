#!/bin/bash
set -e
export ver=$(git describe --dirty)
readonly sha=$(git rev-parse --short HEAD)
readonly pkg="github.com/asips/sdtp-client"
go build -ldflags "-X ${pkg}/internal.Version=${ver} -X ${pkg}/internal.GitSHA=${sha}" ./
