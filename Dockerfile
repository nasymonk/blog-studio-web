FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.23-alpine AS go-builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/blog-studio-web ./cmd/server

FROM alpine:3.20
ARG HUGO_VERSION=0.147.0
ARG TARGETARCH
RUN apk add --no-cache ca-certificates wget tar libc6-compat libstdc++ libgcc \
  && case "${TARGETARCH}" in amd64) HUGO_ARCH=amd64 ;; arm64) HUGO_ARCH=arm64 ;; *) echo "unsupported arch: ${TARGETARCH}" && exit 1 ;; esac \
  && wget -O /tmp/hugo.tar.gz "https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_extended_${HUGO_VERSION}_linux-${HUGO_ARCH}.tar.gz" \
  && tar -xzf /tmp/hugo.tar.gz -C /usr/local/bin hugo \
  && rm /tmp/hugo.tar.gz
WORKDIR /app
COPY --from=go-builder /out/blog-studio-web /usr/local/bin/blog-studio-web
COPY --from=web /src/web/dist /app/web/dist
ENV PORT=8080 \
    BASE_PATH=/studio \
    BLOG_STUDIO_BLOG_ROOT=/blog \
    BLOG_STUDIO_DATA_ROOT=/data \
    BLOG_STUDIO_STATIC_DIR=/app/web/dist
EXPOSE 8080
ENTRYPOINT ["blog-studio-web"]
