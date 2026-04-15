# -----------------------------
# 1. Build Phase
# -----------------------------
FROM golang:1.26.2-alpine3.22 AS build

# Install all tools needed for client + Go + templ
RUN apk --no-cache add gcc g++ make git nodejs npm bash wget

WORKDIR /go/src/app

# Copy everything for build
COPY . .


# -----------------------------
# Client Build
# -----------------------------
WORKDIR /go/src/app/client

RUN npm install --legacy-peer-deps
RUN npm run build

# -----------------------------
# Views Build (templ)
# -----------------------------
WORKDIR /go/src/app/views

RUN go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate

# -----------------------------
# GeoIP Download
# -----------------------------
ARG MAXMIND_LICENSE_KEY

RUN mkdir -p /go/src/app/internal/geoip && \
    cd /go/src/app/internal/geoip && \
    if [ -n "$MAXMIND_LICENSE_KEY" ]; then \
        echo "Downloading GeoIP database..." && \
        wget -q -O GeoLite2.tar.gz \
        "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&license_key=${MAXMIND_LICENSE_KEY}&suffix=tar.gz" && \
        tar -xzf GeoLite2.tar.gz && \
        mv GeoLite2-Country_*/GeoLite2-Country.mmdb . && \
        rm -rf GeoLite2-Country_* GeoLite2.tar.gz && \
        echo "GeoIP DB installed"; \
    else \
        echo "No MAXMIND_LICENSE_KEY provided, skipping GeoIP download"; \
    fi

# -----------------------------
# Go Build
# -----------------------------
WORKDIR /go/src/app

# Prepare Go modules
RUN go mod tidy

RUN GOOS=linux go build -ldflags="-s -w" -o ./bin/urx ./cmd/server/*.go

# -----------------------------
# Release Phase
# -----------------------------
FROM alpine:3.22 AS release

RUN apk update && apk upgrade && apk --no-cache add ca-certificates

WORKDIR /go/bin

COPY --from=build /go/src/app/bin /go/bin
COPY --from=build /go/src/app/static /go/bin/static
COPY --from=build /go/src/app/sql /go/bin/sql
COPY --from=build /go/src/app/internal/geoip /go/bin/geoip

EXPOSE 3373

ENTRYPOINT ["/go/bin/urx", "--port", "3373"]
