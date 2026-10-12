FROM --platform=$BUILDPLATFORM golang:1.27 AS builder

WORKDIR /src

ARG MODULE_VERSION=main

RUN --mount=type=cache,target=/go/pkg/mod/,sharing=locked \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download

ARG TARGETARCH
ARG TARGETOS

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build -o /echonetlite-exporter

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=builder /echonetlite-exporter /echonetlite-exporter

ENTRYPOINT ["/echonetlite-exporter"]
