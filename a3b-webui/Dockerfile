FROM golang:alpine as build
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache git
RUN go install github.com/a-h/templ/cmd/templ@latest
WORKDIR /build
ADD . .
RUN templ generate ./a3b-webui/
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags '-extldflags "-static"' -o /build/app /build/a3b-webui/cmd/server

FROM alpine:3.23
RUN apk add -U tzdata
COPY --from=build /etc/ssl/certs/ca-certificates.crt \
    /etc/ssl/certs/ca-certificates.crt
COPY --from=build /build/app /app
COPY --from=build /build/a3b-webui/static /static
RUN mkdir /conf
ENTRYPOINT ["/app"]
