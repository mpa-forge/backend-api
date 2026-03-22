FROM golang:1.25.1-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/backend-api ./cmd/api

FROM alpine:3.22 AS runtime

RUN adduser -D -u 10001 appuser

COPY --from=build /out/backend-api /backend-api

USER appuser

EXPOSE 8080

ENV HTTP_PORT=8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
	CMD wget -q -O /dev/null "http://127.0.0.1:${HTTP_PORT}/healthz" || exit 1

ENTRYPOINT ["/backend-api"]
