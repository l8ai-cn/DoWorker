# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates protobuf protobuf-dev bash
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 \
  && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1 \
  && bash scripts/proto-gen-go.sh --force
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
  -o /out/relay ./relay/cmd/relay

FROM alpine:3.19
RUN apk add --no-cache ca-certificates wget \
  && adduser -D -u 1000 -g 1000 app \
  && mkdir -p /app /data/acme \
  && chown -R 1000:1000 /app /data
WORKDIR /app
COPY --from=build --chown=1000:1000 /out/relay /app/relay
USER 1000
EXPOSE 8090
ENTRYPOINT ["/app/relay"]
