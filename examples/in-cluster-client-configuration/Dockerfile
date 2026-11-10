FROM golang:1.26 AS builder
# Allow synced go.mod files to select a newer Go toolchain than the base image.
ENV GOTOOLCHAIN=auto
WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./examples/in-cluster-client-configuration

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/app /app
ENTRYPOINT ["/app"]
