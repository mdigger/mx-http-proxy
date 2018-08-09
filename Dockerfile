ARG version="develop"
ARG commit
ARG date
ARG LOG
ARG MX
ARG PORT="8000"

# FROM golang:rc-alpine AS builder
FROM golang:alpine AS builder
ARG version
ARG commit
ARG date
RUN ["apk", "--no-cache", "add", "git"]
WORKDIR /go/src/github.com/mdigger/mx-http-proxy
COPY . .
RUN ["go", "get", "-d", "github.com/shurcooL/vfsgen", \
"github.com/shurcooL/httpfs/filter", "./..."]
RUN ["go", "generate"]
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
ENV ldflags="-w -s -X main.version=${version} -X main.commit=${commit} -X main.date=${date}"
RUN go install -i -ldflags "${ldflags}" -a -installsuffix cgo

FROM scratch
ARG version
ARG commit
ARG date
ARG LOG
ARG MX
ARG PORT
LABEL maintainer="dmitrys@xyzrd.com" \
org.label-schema.name="MX HTTP API" \
org.label-schema.description="MX HTTP REST API Proxy service" \
org.label-schema.vendor="xyzrd.com" \
org.label-schema.version=${version} \
org.label-schema.build-date=${date} \
org.label-schema.vcs-url="https://github.com/mdigger/mx-http-proxy" \
org.label-schema.vcs-ref=${commit} \
org.label-schema.schema-version="1.0"
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/mx-http-proxy /
ENV PORT="${PORT}" MX="${MX}" LOG="${LOG}"
EXPOSE ${PORT}
VOLUME ["/letsEncrypt.cache", "/certs"]
ENTRYPOINT ["/mx-http-proxy"]
# CMD ["--help"]
