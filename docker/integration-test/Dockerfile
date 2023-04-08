FROM golang:1.19-bullseye

WORKDIR /workdir
COPY go.mod go.sum /workdir/
RUN go mod download
COPY . .
ENTRYPOINT ["go", "test", "-v", "-tags", "integration", "." ]
