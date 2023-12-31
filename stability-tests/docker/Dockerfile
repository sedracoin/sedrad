ARG SEDRAD_IMAGE
ARG SEDRAMINER_IMAGE

FROM ${SEDRAD_IMAGE} as sedrad
FROM ${SEDRAMINER_IMAGE} as sedraminer

FROM golang:1.19-alpine

RUN mkdir -p /go/src/github.com/sedracoin/sedrad

WORKDIR /go/src/github.com/sedracoin/sedrad

RUN apk add bash build-base git

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

COPY --from=sedrad /app/ /app/
COPY --from=sedraminer /app/ /app/
ENV PATH="/app:${PATH}"

WORKDIR /go/src/github.com/sedracoin/sedrad/stability-tests

RUN go install ./...

ENTRYPOINT ["./run/run.sh"]