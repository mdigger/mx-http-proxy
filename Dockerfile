ARG VERSION
ARG COMMIT
ARG DATE

# FROM golang:rc-alpine AS builder
FROM golang:alpine AS builder
ARG VERSION
ARG COMMIT
ARG DATE
ENV ldflags="-w -s -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${DATE}"
RUN echo "mdigger:x:1000:1000::/app:" > /tmp/passwd
RUN ["apk", "--no-cache", "add", "git"]
WORKDIR /go/src/github.com/mdigger/mx-http-proxy
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
COPY . .
RUN ["go", "get", "-d", "github.com/shurcooL/vfsgen", \
"github.com/shurcooL/httpfs/filter", "./..."]
RUN ["go", "generate"]
RUN go install -i -ldflags "${ldflags}" -a -installsuffix cgo ./...

FROM scratch
ARG VERSION
ARG COMMIT
ARG DATE
LABEL version=${VERSION:-"dev"} commit=${COMMIT} date=${DATE} \
maintainer="dmitrys@xyzrd.com" company="xyzrd.com"
WORKDIR /app
COPY --from=builder /tmp/passwd /etc/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/mx-http-proxy /app/
USER mdigger
ENV PORT="8000" MX="" LOG="" PATH="/app"
EXPOSE ${PORT}
ENTRYPOINT ["/app/mx-http-proxy"]
# CMD ["--help"]
