FROM golang:1.15-alpine AS builder
WORKDIR /
WORKDIR /go/src/poor-man-service-mesh
COPY main.go run.go config.go ./
COPY pkg ./pkg
RUN apk update && apk add git
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go get -v
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go install -v .

FROM alpine
ADD docker-entrypoint.sh /
RUN mkdir /ca && chown 1000:1000 /ca
RUN chmod +x /docker-entrypoint.sh
RUN apk update && apk add curl
USER 1000
COPY --from=builder /go/bin/poor-man-service-mesh /poor-man-service-mesh
ENTRYPOINT [ "/docker-entrypoint.sh" ]
CMD ["/poor-man-service-mesh", "-directory-url", "https://acme.unusual.one/acme/development/directory"]
