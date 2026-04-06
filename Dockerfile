# -----------------------------
# 1. Build Phase
# -----------------------------
FROM golang:1.26.1-alpine3.22 AS build

# Install all tools needed for client + Go + templ
RUN apk --no-cache add gcc g++ make git nodejs npm bash

WORKDIR /go/src/app

# Copy everything for build
COPY . .

# Prepare Go modules
RUN go mod tidy


# -----------------------------
# Client Build
# -----------------------------
WORKDIR /go/src/app/client

RUN npm install
RUN npm run build

# -----------------------------
# Views Build (templ)
# -----------------------------
WORKDIR /go/src/app/views

RUN go install github.com/a-h/templ/cmd/templ@0.3.857
RUN templ generate

# -----------------------------
# Go Build
# -----------------------------
WORKDIR /go/src/app

RUN GOOS=linux go build -ldflags="-s -w" -o ./bin/gosot ./cmd/server/*.go

# -----------------------------
# Release Phase
# -----------------------------
FROM alpine:3.22 AS release

RUN apk update && apk upgrade && apk --no-cache add ca-certificates

WORKDIR /go/bin

COPY --from=build /go/src/app/bin /go/bin
COPY --from=build /go/src/app/static /go/bin/static
COPY --from=build /go/src/app/sql /go/bin/sql

EXPOSE 3388

ENTRYPOINT ["/go/bin/gosot", "--port", "3388"]
