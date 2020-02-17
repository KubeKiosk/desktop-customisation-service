FROM golang:latest

WORKDIR /go/src/github.com/kubekiosk/desktop-customisation-service
COPY main.go .
RUN go get github.com/containerd/containerd
RUN go get github.com/gorilla/mux
RUN go build -ldflags "-linkmode external -extldflags -static" -a main.go

FROM scratch
COPY --from=0 /go/src/github.com/kubekiosk/desktop-customisation-service/main /main
CMD ["/main"]