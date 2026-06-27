FROM golang:alpine as build
ARG APP_VERSION=dev
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache git
RUN apk add --no-cache build-base
WORKDIR /build
ADD . .
RUN CGO_ENABLED=1 GOOS=linux \
go build -ldflags "-X main.appVersion=$APP_VERSION -extldflags '-static'" -o app /build/mcore/cmd/app

FROM alpine:3.23
ARG APP_VERSION=dev
ENV APP_VERSION=${APP_VERSION}
RUN apk add -U tzdata
COPY --from=build /etc/ssl/certs/ca-certificates.crt \
/etc/ssl/certs/ca-certificates.crt
COPY --from=build /build/app /app
RUN mkdir /conf
RUN mkdir /data
EXPOSE 8080

ENTRYPOINT ["/app"]
