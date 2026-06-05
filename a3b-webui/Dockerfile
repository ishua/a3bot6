FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Generate templ templates
RUN templ generate

# Build
RUN CGO_ENABLED=0 go build -o /build/a3b-webui ./cmd/server

FROM alpine:3.23

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /build/a3b-webui .

EXPOSE 8090

ENTRYPOINT ["./a3b-webui"]
