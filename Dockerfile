
#build stage
FROM golang:alpine AS builder
RUN apk --no-cache add git
WORKDIR /go/src/github.com/mdigger/mx-http-proxy
COPY . .
RUN go get -d -v github.com/shurcooL/vfsgen github.com/shurcooL/httpfs/filter ./...
RUN go generate
RUN CGO_ENABLED=0 GOOS=linux go install -v -ldflags '-w -s' -a -installsuffix cgo ./...

# COPY .git .
# RUN GIT_COMMIT=$(git rev-list -1 HEAD) && \
#     go build -ldflags "-X main.GitCommit=$GIT_COMMIT"

#final stage
FROM scratch 
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/mx-http-proxy ./mx-http-proxy
ENTRYPOINT ["./mx-http-proxy", "-http=:8000"]
LABEL Name="mx-http-proxy" Version="0.0.1"
EXPOSE 8000
