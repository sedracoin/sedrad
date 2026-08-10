#!/usr/bin/env bash
# Reproducible build for the SEDRA v2.0.0 relaunch client (branch fork/v2.0.0).
# Builds the full node + tooling set with the kHeavyHash cgo toolchain.
set -euo pipefail

export PATH="/c/Program Files/Go/bin:$PATH"
export CGO_ENABLED=1
export CC="C:/Users/sdr/AppData/Local/Microsoft/WinGet/Packages/BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe/mingw64/bin/gcc.exe"
export GOFLAGS=-mod=mod

cd "$(dirname "$0")"

# appBuild MUST contain only [A-Za-z0-9-]; dots are rejected by version.checkAppBuild.
LDFLAGS='-X github.com/sedracoin/sedrad/version.appBuild=relaunch'

OUT="${1:-.}"        # optional output dir (e.g. dist/win); defaults to repo root
mkdir -p "$OUT"

echo "Go:  $(go version)"
echo "CC:  $("$CC" --version | head -1)"
echo "Out: $OUT"
echo "Building v2.0.0 relaunch: sedrad, sedrawallet, sedractl, sedraminer, genkeypair ..."
go build -ldflags "$LDFLAGS" -o "$OUT/sedrad.exe" .
go build -ldflags "$LDFLAGS" -o "$OUT/sedrawallet.exe" ./cmd/sedrawallet
go build -ldflags "$LDFLAGS" -o "$OUT/sedractl.exe"    ./cmd/sedractl
go build -ldflags "$LDFLAGS" -o "$OUT/sedraminer.exe"  ./cmd/sedraminer
go build -ldflags "$LDFLAGS" -o "$OUT/genkeypair.exe"  ./cmd/genkeypair
echo "Done:"
ls -lh "$OUT"/sedrad.exe "$OUT"/sedrawallet.exe "$OUT"/sedractl.exe "$OUT"/sedraminer.exe "$OUT"/genkeypair.exe
"$OUT/sedrad.exe" --version
