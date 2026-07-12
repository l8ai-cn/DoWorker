FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
  -o /out/marketplace ./marketplace/cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates \
  && adduser -D -u 1000 -g 1000 app \
  && mkdir -p /app \
  && chown -R 1000:1000 /app
WORKDIR /app
COPY --from=build --chown=1000:1000 /out/marketplace /app/marketplace
USER 1000
EXPOSE 8080
ENTRYPOINT ["/app/marketplace"]
