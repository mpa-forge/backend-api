FROM golang:1.24.12-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY cmd ./cmd

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/backend-api ./cmd/api-placeholder

FROM alpine:3.22 AS runtime

RUN adduser -D -u 10001 appuser

COPY --from=build /out/backend-api /backend-api

USER appuser

EXPOSE 8080

ENV HTTP_PORT=8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
	CMD wget -q -O /dev/null http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/backend-api"]
