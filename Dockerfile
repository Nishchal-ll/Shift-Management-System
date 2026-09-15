FROM golang:1.24-alpine

WORKDIR /app

# Install git, build-base, and air for live reloading
RUN apk add --no-cache git build-base && \
    go install github.com/air-verse/air@v1.61.7

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]