# syntax=docker/dockerfile:1
# Production image for the Go API: static binary on distroless (no shell,
# runs as non-root). Build from this directory:  docker build -t mindflow-api .
FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
ENV PORT=3000 NODE_ENV=production
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/server"]
