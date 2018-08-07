ARG VERSION
ARG COMMIT
ARG DATE

# FROM golang:alpine AS builder
FROM golang:rc-alpine AS builder
ARG VERSION
ARG COMMIT
ARG DATE
RUN apk --no-cache add git
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go install -i -ldflags "-w -s -X main.version=$VERSION -X main.commit=$COMMIT -X main.buildDate=$DATE" \
-a -installsuffix cgo ./... && \
echo "mdigger:x:1000:1000::/app:" > passwd

FROM scratch
ARG VERSION
ARG COMMIT
ARG DATE
LABEL version=${VERSION:-"dev"} commit=${COMMIT} date=${DATE} \
maintainer="dmitrys@xyzrd.com" company="xyzrd.com"
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/mxhttp /app/
COPY --from=builder /app/passwd /etc/
USER mdigger
ENV PORT="8000" MX="" PATH="/app"
EXPOSE ${PORT}
ENTRYPOINT ["/app/mxhttp"]
# CMD ["--help"]
