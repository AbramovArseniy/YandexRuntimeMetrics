
ARG GOLANG_IMAGE_TAG="1.19-alpine3.17"
ARG ALPINE_IMAGE_TAG="3.17"

FROM golang:$GOLANG_IMAGE_TAG as go
RUN apk update && \
    apk add gcc-go musl-dev openssl-dev ca-certificates && \
    update-ca-certificates

FROM go as mod_cache
WORKDIR /app
COPY go.* .
RUN go mod download

FROM go as test
COPY --from=mod_cache $GOPATH $GOPATH
WORKDIR /app
COPY . .
RUN go test -race -v ./...

FROM test as server-build
RUN go build -race -v -o ./bin/server ./cmd/server/...

FROM test as agent-build
RUN go build -race -v -o ./bin/agent ./cmd/agent/...

FROM alpine:$ALPINE_IMAGE_TAG as server
WORKDIR /app
COPY --from=server-build /app/bin/server .
EXPOSE 8080
ENV ADDRESS="127.0.0.1:8080"
ENV STORE_INTERVAL="300s"
ENV STORE_FILE="/tmp/devops-metrics-db.json"
ENV RESTORE="true"
ENV KEY=""
ENV DATABASE_DSN=""
VOLUME /tmp
ENTRYPOINT /app/server

FROM alpine:$ALPINE_IMAGE_TAG as agent
COPY --from=agent-build /app/bin/agent /usr/local/bin/agent
ENV ADDRESS="127.0.0.1:8080"
ENV POLL_INTERVAL="10s"
ENV REPORT_INTERVAL="2s"
ENV KEY=""
ENV RATE_LIMIT=""
ENTRYPOINT agent
