# Should be started with:
# docker run -ti -p 2121-2130:2121-2130 fclairamb/ftpserver

# Preparing the build environment
FROM golang:1.23-alpine AS builder
ENV GOFLAGS="-mod=readonly"
RUN apk add --update --no-cache bash ca-certificates curl git && update-ca-certificates
RUN mkdir -p /workspace
WORKDIR /workspace

# Building
COPY . .
RUN CGO_ENABLED=0 go build -mod=readonly -ldflags='-w -s' -v -o ftpserver

# Preparing the final image
FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 2121-2130
COPY --from=builder /workspace/ftpserver /bin/ftpserver
ENTRYPOINT [ "/bin/ftpserver" ]
