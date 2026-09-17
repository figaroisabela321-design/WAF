# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY waf-control/go.mod waf-control/go.sum ./
RUN go mod download
COPY waf-control/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/waf-control ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/waf-control /app/waf-control
# openapi is embedded; keep a copy for operators
COPY --from=builder /src/docs/openapi.yaml /app/docs/openapi.yaml
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/waf-control"]
