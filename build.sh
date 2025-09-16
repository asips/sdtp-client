#!/bin/bash
set -e
export ver=$(git describe --dirty)
readonly sha=$(git rev-parse --short HEAD)
readonly pkg="github.com/asips/sdtp-client"

#!/bin/bash
#
# Build script for local or dev builds. Production or release builds should be
# handled via the SLA Go Releaser GitHub Action.
#
set -e
export builddir=./build

commit_date=$(git log --date=iso8601-strict -1 --pretty=%ct)
commit=$(git rev-parse HEAD)
version=$(git describe --tags --always --dirty | cut -c2-)
tree_state=$(if git diff --quiet; then echo "clean"; else echo "dirty"; fi)

function build() {
  export GOOS=$1
  export GOARCH=$2
  export CGO_ENABLED=0

  binname=sdtp-${GOOS}-${GOARCH}
  if [[ $GOOS == "windows" ]]; then
    binname=${binname}.exe
  fi

  go build -o ${builddir}/${binname} -ldflags "-X ${pkg}/internal.Version=${ver} -X ${pkg}/internal.GitSHA=${sha}" 
}

mkdir -pv $builddir
rm -fv $builddir/*

build linux amd64
build windows amd64
build darwin amd64
build darwin arm64
