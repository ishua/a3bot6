FROM golang:alpine AS build
ARG APP_VERSION=dev
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache git
# keep in sync with github.com/a-h/templ version in a3b-webui/go.mod
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1020
WORKDIR /build
ADD . .
RUN cd a3b-webui && templ generate .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags "-X main.appVersion=$APP_VERSION -extldflags '-static'" -o /build/app /build/a3b-webui/cmd/server

FROM alpine:3.23
ARG APP_VERSION=dev
ENV APP_VERSION=${APP_VERSION}
RUN apk add -U tzdata
COPY --from=build /etc/ssl/certs/ca-certificates.crt \
    /etc/ssl/certs/ca-certificates.crt
COPY --from=build /build/app /app
COPY --from=build /build/a3b-webui/static /static
RUN mkdir /conf
ENTRYPOINT ["/app"]
