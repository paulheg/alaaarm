# use go image
FROM golang AS builder

# copy source files to GO HOME
COPY . /go/src/github.com/paulheg/alaaarm

# dependency management
WORKDIR /go/src/github.com/paulheg/alaaarm
RUN go mod init

# build
RUN CGO_ENABLED=0 GOOS=linux go build -o server -i /go/src/github.com/paulheg/alaaarm/cmd/alaaarm/ 

FROM alpine:latest
WORKDIR /root/
COPY --from=builder ./go/src/github.com/paulheg/alaaarm/cmd/alaaarm/server .
ENTRYPOINT [ "server" ]
EXPOSE 8080