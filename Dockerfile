# Should be started with:
# docker run -ti -p 2121-2130:2121-2130 fclairamb/ftpserver

# Preparing the build environment. The binary is cross-compiled on the build platform, which is much faster than
# building it in an emulated environment.
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder
ENV GOFLAGS="-mod=readonly" CGO_ENABLED=0
RUN apk add --update --no-cache ca-certificates && update-ca-certificates
WORKDIR /workspace

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Building
COPY . .
ARG TARGETOS TARGETARCH TARGETVARIANT
ARG VERSION=dev COMMIT="" BUILD_DATE=""
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    if [ "$TARGETARCH" = "arm" ]; then export GOARM="${TARGETVARIANT#v}"; fi && \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -v -o ftpserver \
      -ldflags="-w -s -X main.BuildVersion=${VERSION} -X main.Commit=${COMMIT} -X main.BuildDate=${BUILD_DATE}"

# The directories the server writes to, owned by the unprivileged user (65532 is the distroless "nonroot" id)
RUN install -d -o 65532 -g 65532 /rootfs/app && install -d -m 1777 -o 65532 -g 65532 /rootfs/tmp

# Preparing the final image
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# The default configuration serves the /tmp directory
COPY --from=builder --chown=65532:65532 /rootfs/tmp /tmp
COPY --from=builder --chown=65532:65532 /rootfs/app /app
USER 65532:65532
WORKDIR /app
EXPOSE 2121-2130
COPY --from=builder /workspace/ftpserver /bin/ftpserver
ENTRYPOINT [ "/bin/ftpserver" ]
