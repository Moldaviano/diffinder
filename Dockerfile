FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/seed   ./cmd/seed

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 1000 app
USER app
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /out/seed   /app/seed
EXPOSE 8080
ENTRYPOINT ["/app/server"]
