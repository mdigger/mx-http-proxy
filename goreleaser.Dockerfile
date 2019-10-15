FROM alpine:latest as alpine
RUN apk add -U --no-cache ca-certificates

FROM scratch
LABEL maintainer="sedykh@gmail.com" \
org.label-schema.name="MX HTTP API" \
org.label-schema.description="MX HTTP REST API Proxy service" \
org.label-schema.vendor="xyzrd.com" \
org.label-schema.vcs-url="https://github.com/mdigger/mx-http-proxy" \
org.label-schema.schema-version="1.0"
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY mx-http-proxy /
ENV PORT="8000" MX="" LOG=""
EXPOSE ${PORT}
ENTRYPOINT ["/mx-http-proxy"]
