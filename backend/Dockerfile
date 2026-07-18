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
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
  -o /out/migrate-ai-resources ./backend/cmd/migrate-ai-resources
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
  -o /out/migrate-worker-config-documents ./backend/cmd/migrate-worker-config-documents
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
  -o /out/worker-definition-sync ./backend/cmd/worker-definition-sync

FROM alpine:3.19
RUN apk add --no-cache ca-certificates wget \
  && adduser -D -u 1000 -g 1000 app \
  && mkdir -p /app /data/acme \
  && chown -R 1000:1000 /app /data
WORKDIR /app
COPY --from=build --chown=1000:1000 /out/server /app/server
COPY --from=build --chown=1000:1000 /out/migrate-ai-resources /app/migrate-ai-resources
COPY --from=build --chown=1000:1000 /out/migrate-worker-config-documents /app/migrate-worker-config-documents
COPY --from=build --chown=1000:1000 /out/worker-definition-sync /app/worker-definition-sync
COPY --from=build --chown=1000:1000 /src/config/worker-types /app/config/worker-types
USER 1000
EXPOSE 8080 9443
ENTRYPOINT ["/app/server"]
