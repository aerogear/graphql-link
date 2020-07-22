FROM golang:alpine as build
RUN apk --no-cache add ca-certificates

# In case we want to build the gateway in docker....
#WORKDIR /go/src/app
#COPY . .
#RUN CGO_ENABLED=0 go-wrapper install -ldflags '-extldflags "-static"'

FROM scratch
# FROM alpine
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ADD bin /bin
ADD etc /etc
WORKDIR /etc/graphql-gw
EXPOSE 8080/tcp
ENTRYPOINT ["/bin/graphql-gw", "serve"]