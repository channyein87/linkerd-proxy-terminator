FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.22 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG VERSION

ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /go/src/github.com/channyein87/linkerd-proxy-terminator

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY .  .

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -ldflags "-s -w -X main.Version=${VERSION}" \
  -a -o /usr/bin/proxy-terminator .

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static:nonroot

LABEL org.opencontainers.image.source=https://github.com/channyein87/linkerd-proxy-terminator

WORKDIR /
COPY --from=builder /usr/bin/proxy-terminator /
USER nonroot:nonroot

CMD ["/proxy-terminator"]