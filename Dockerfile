# Multi-stage build of the SEDRA relaunch node (cgo: go-muhash / go-secp256k1).
# Build context = this repo (relaunch/sedrad). Produces sedrad + wallet + ctl.
FROM golang:1.23-bookworm AS build
ENV CGO_ENABLED=1 GOFLAGS=-mod=mod
WORKDIR /src
# gcc/g++ are present in the bookworm golang image; needed for the cgo packages.
COPY . .
ARG APPBUILD=relaunch
RUN go build -ldflags "-X github.com/sedracoin/sedrad/version.appBuild=${APPBUILD}" -o /out/sedrad . \
 && go build -ldflags "-X github.com/sedracoin/sedrad/version.appBuild=${APPBUILD}" -o /out/sedrawallet ./cmd/sedrawallet \
 && go build -ldflags "-X github.com/sedracoin/sedrad/version.appBuild=${APPBUILD}" -o /out/sedractl ./cmd/sedractl

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/* \
 && useradd -m -u 10001 sedra
COPY --from=build /out/sedrad /out/sedrawallet /out/sedractl /usr/local/bin/
USER sedra
WORKDIR /home/sedra
VOLUME ["/home/sedra/.sedrad"]
# P2P must be reachable; RPC (22110) is intentionally NOT EXPOSEd here — compose
# keeps it on the internal network only (the gRPC RPC is unauthenticated).
EXPOSE 22111
ENTRYPOINT ["sedrad"]
