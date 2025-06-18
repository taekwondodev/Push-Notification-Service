FROM --platform=$BUILDPLATFORM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build \
    -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -trimpath \
    -o server ./cmd/server/main.go

FROM scratch

COPY --from=builder /app/server /server

USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/server"]