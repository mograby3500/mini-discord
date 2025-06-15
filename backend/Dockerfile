# Stage 1 - Build
FROM golang:1.24-alpine as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . ./

WORKDIR /app/cmd/api

RUN go build -o /server

# Stage 2 - Run
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /server .

EXPOSE 8080
CMD ["./server"]
