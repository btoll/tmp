FROM golang:1.17 as build

WORKDIR /go/src/gotest
ADD go.mod main.go /go/src/gotest/

RUN go build -o /go/src/gotest

FROM gcr.io/distroless/base
COPY --from=build /go/src/gotest/gotest /usr/bin
ENTRYPOINT ["gotest"]

