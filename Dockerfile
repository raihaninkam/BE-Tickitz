FROM golang:latest AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o tickitz ./cmd/main.go

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /app/tickitz .
EXPOSE 8080
CMD ["./tickitz"]

