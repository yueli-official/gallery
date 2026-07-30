FROM golang:1.25.12-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY api ./api
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gallery-api ./cmd/gallery
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gallery-healthcheck ./cmd/healthcheck
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gallery-bindingcheck ./cmd/bindingcheck
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gallery-bootstrap ./cmd/bootstrap

FROM migrate/migrate:v4.19.1 AS migration
COPY manifest/sql/migrations /migrations

FROM gcr.io/distroless/static-debian12:nonroot AS bindingcheck
COPY --from=build /out/gallery-bindingcheck /usr/local/bin/gallery-bindingcheck
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/gallery-bindingcheck"]

FROM gcr.io/distroless/static-debian12:nonroot AS bootstrap
COPY --from=build /out/gallery-bootstrap /usr/local/bin/gallery-bootstrap
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/gallery-bootstrap"]

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=build /out/gallery-api /usr/local/bin/gallery-api
COPY --from=build /out/gallery-healthcheck /usr/local/bin/gallery-healthcheck
COPY --chown=nonroot:nonroot manifest ./manifest

ENV GF_SERVER_ADDRESS=0.0.0.0:8091
ENV OTEL_TRACES_EXPORTER=none

EXPOSE 8091
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/gallery-api"]
