# syntax=docker/dockerfile:1

############################
# dev/test image (with toolchain)
############################
FROM golang:1.24 AS devtest

WORKDIR /src

# Cache deps and install tools
COPY go.mod go.sum tools.go ./
RUN go mod download
RUN cat tools.go | grep _ | awk -F'"' '{print $2}' | xargs -tI % go install %

# Copy code
COPY . .

# Default command is tests (CI-friendly). Override in compose/CLI as needed.
CMD ["go", "test", "./...", "-count=1"]

############################
# build binary
############################
FROM golang:1.24 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a static binary for distroless/static
# -trimpath reduces build path leakage
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

############################
# build coverage binary
############################
FROM golang:1.24 AS build-coverage

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -cover -o /out/server-coverage ./cmd/server

############################
# prod coverage image
############################
FROM gcr.io/distroless/static-debian12:nonroot AS prod-coverage

WORKDIR /

COPY --from=build-coverage /out/server-coverage /server

# Environment defaults
ENV ADDR=:8080
ENV EXTERNAL_URL=https://httpbin.org/status/204
ENV GOCOVERDIR=/coverage

EXPOSE 8080

VOLUME /coverage

USER nonroot:nonroot
ENTRYPOINT ["/server"]

############################
# prod image (distroless)
############################
FROM gcr.io/distroless/static-debian12:nonroot AS prod

WORKDIR /

COPY --from=build /out/server /server

# Environment defaults (override at runtime)
ENV ADDR=:8080
ENV EXTERNAL_URL=https://httpbin.org/status/204

EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/server"]
