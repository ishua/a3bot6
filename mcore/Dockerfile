FROM golang:alpine as build
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache git
RUN apk add --no-cache build-base
WORKDIR /build
ADD . .
RUN CGO_ENABLED=0 GOOS=linux \
go build -ldflags '-extldflags "-static"' -o app /build/mcore/cmd/app

FROM alpine:3.21
RUN apk add -U tzdata
COPY --from=build /etc/ssl/certs/ca-certificates.crt \
/etc/ssl/certs/ca-certificates.crt
COPY --from=build /build/app /app
RUN mkdir /conf
EXPOSE 8080

ENTRYPOINT ["/app"]