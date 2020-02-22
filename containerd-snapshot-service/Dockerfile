FROM golang:latest AS agent

WORKDIR /go/src/github.com/kubekiosk/desktop-customisation-service
COPY main.go .
RUN go get github.com/containerd/containerd
RUN go get github.com/gorilla/mux
RUN go build -ldflags "-linkmode external -extldflags -static" -a main.go

FROM golang:latest AS deps 
RUN go get github.com/containerd/containerd
WORKDIR /go/src/github.com/containerd/containerd
RUN go build -ldflags "-linkmode external -extldflags -static" -a /go/src/github.com/containerd/containerd/cmd/ctr/main.go

FROM scratch
COPY ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=agent /go/src/github.com/kubekiosk/desktop-customisation-service/main /snapshot-agent
COPY --from=deps /go/src/github.com/containerd/containerd/main /ctr
ENV PATH=/
CMD ["/snapshot-agent"]