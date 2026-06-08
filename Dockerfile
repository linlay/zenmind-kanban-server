FROM golang:1.26 AS build

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/zenmind-kanban-server ./cmd/server

FROM alpine:3.22

WORKDIR /app
COPY --from=build /out/zenmind-kanban-server /app/zenmind-kanban-server
EXPOSE 8080
RUN adduser -D -H -u 10001 appuser && mkdir -p /data && chown -R appuser:appuser /data
USER appuser
HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/app/zenmind-kanban-server"]
