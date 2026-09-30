# Production image: one image, two binaries — the API server and the console.
# Both are built from the same composition root, so the console always matches
# the deployed code. Production access is:
#   docker compose -f docker-compose.prod.yml exec api /app/console          # read-only
#   docker compose -f docker-compose.prod.yml exec api /app/console -write   # writes allowed
FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags='-s -w' -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -ldflags='-s -w' -o /out/console ./cmd/console

FROM alpine:3.21

RUN adduser -D -u 10001 app

COPY --from=build /out/api     /app/api
COPY --from=build /out/console /app/console

USER app
WORKDIR /app
EXPOSE 8080

ENTRYPOINT ["/app/api"]
