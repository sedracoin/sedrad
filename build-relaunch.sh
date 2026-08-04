#!/usr/bin/env bash
# Reproducible build for the SEDRA relaunch client (branch relaunch/v1.1.0).
# Same toolchain as build-sedrad.sh; adds appBuild="relaunch" version marker.
set -euo pipefail

export PATH="/c/Program Files/Go/bin:$PATH"
export CGO_ENABLED=1
export CC="C:/Users/sdr/AppData/Local/Microsoft/WinGet/Packages/BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe/mingw64/bin/gcc.exe"
export GOFLAGS=-mod=mod

cd "$(dirname "$0")"

LDFLAGS='-X github.com/sedracoin/sedrad/version.appBuild=relaunch'

echo "Go:  $(go version)"
echo "CC:  $("$CC" --version | head -1)"
echo "Building relaunch sedrad, sedrawallet, sedractl ..."
go build -ldflags "$LDFLAGS" -o sedrad.exe .
go build -ldflags "$LDFLAGS" -o sedrawallet.exe ./cmd/sedrawallet
go build -ldflags "$LDFLAGS" -o sedractl.exe ./cmd/sedractl
echo "Done:"
ls -lh sedrad.exe sedrawallet.exe sedractl.exe
./sedrad.exe --version
