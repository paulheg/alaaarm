# use go image
FROM golang:1.17.1-alpine AS builder

# copy source files to GO HOME
COPY . /go/src/github.com/paulheg/alaaarm

# dependency management
WORKDIR /go/src/github.com/paulheg/alaaarm/cmd/alaaarm/
RUN go mod download

# build
RUN CGO_ENABLED=0 GOOS=linux go build -o server /go/src/github.com/paulheg/alaaarm/cmd/alaaarm/ 

FROM alpine:latest
WORKDIR /
COPY --from=builder /go/src/github.com/paulheg/alaaarm/migration/ /migration
COPY --from=builder /go/src/github.com/paulheg/alaaarm/cmd/alaaarm/server .

ENTRYPOINT [ "./server", "run", "-config=./config/config.json" ]
EXPOSE 3000