# ---------- build stage ----------
FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# ---------- runtime stage ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=build /out/api ./api
COPY --from=build /out/migrate ./migrate
COPY migrations ./migrations

USER app

EXPOSE 8080

# jalankan migrasi lalu server
CMD ["sh", "-c", "./migrate up && exec ./api"]
