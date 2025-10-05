FROM golang:latest AS build
WORKDIR /app

# Copy dan download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code dan build binary
COPY . .
RUN go build -o tickitz ./cmd/main.go

# --- Stage runtime ---
FROM debian:bookworm-slim
WORKDIR /app

# Copy hasil build
COPY --from=build /app/tickitz .


COPY --from=build /app/public ./public


RUN mkdir -p /app/public && chmod -R 755 /app/public

EXPOSE 9001
CMD ["./tickitz"]
