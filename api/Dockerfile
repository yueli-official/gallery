FROM golang:1.25.12-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY api ./api
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gallery-api ./cmd/gallery

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=build /out/gallery-api /usr/local/bin/gallery-api
COPY --chown=nonroot:nonroot manifest ./manifest

ENV GF_SERVER_ADDRESS=0.0.0.0:8091
ENV OTEL_TRACES_EXPORTER=none

EXPOSE 8091
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/gallery-api"]
