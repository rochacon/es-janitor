FROM golang:1.11-alpine AS builder
RUN apk add ca-certificates git
RUN grep nobody /etc/passwd > /etc/passwd.nobody

ENV GO111MODULE on
COPY go.mod /go/src/github.com/rochacon/es-janitor/
COPY go.sum /go/src/github.com/rochacon/es-janitor/
WORKDIR /go/src/github.com/rochacon/es-janitor/
RUN go mod download

COPY . /go/src/github.com/rochacon/es-janitor/
RUN CGO_ENABLED=0 go install -v -ldflags "-X main.VERSION=$(git describe --abbrev=10 --always --dirty)" .

FROM scratch
COPY --from=builder /etc/passwd.nobody /etc/passwd
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=builder /go/bin/janitor /es-janitor
USER nobody
ENTRYPOINT ["/es-janitor"]
