# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates protobuf protobuf-dev bash
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV PATH="/go/bin:${PATH}"
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 \
  && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1 \
  && go install github.com/bufbuild/buf/cmd/buf@v1.50.0 \
  && bash scripts/proto-gen-go.sh --force \
  && bash scripts/sync-amesh-convert.sh --force
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
  -o /out/server ./backend/cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates wget \
  && adduser -D -u 1000 -g 1000 app \
  && mkdir -p /app /data/acme \
  && chown -R 1000:1000 /app /data
WORKDIR /app
COPY --from=build --chown=1000:1000 /out/server /app/server
USER 1000
EXPOSE 8080 9443
ENTRYPOINT ["/app/server"]
