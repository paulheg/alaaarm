# use go image
FROM --platform=$BUILDPLATFORM golang:latest AS builder
ARG TARGETOS
ARG TARGETARCH

# copy source files to GO HOME
COPY . /go/src/github.com/paulheg/alaaarm

# dependency management
WORKDIR /go/src/github.com/paulheg/alaaarm/cmd/alaaarm/
RUN go mod download

# build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o server /go/src/github.com/paulheg/alaaarm/cmd/alaaarm/ 

FROM alpine:latest
WORKDIR /
COPY --from=builder /go/src/github.com/paulheg/alaaarm/migration/ /migration
COPY --from=builder /go/src/github.com/paulheg/alaaarm/localizations/ /localizations
COPY --from=builder /go/src/github.com/paulheg/alaaarm/web/templates /web/templates
COPY --from=builder /go/src/github.com/paulheg/alaaarm/cmd/alaaarm/server .

RUN mkdir config && ./server config create -config=./config/config.json

ENV PORT=3000
EXPOSE 3000

ENTRYPOINT [ "./server", "run", "-config=./config/config.json" ]