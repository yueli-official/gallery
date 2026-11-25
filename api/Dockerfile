FROM golang:1.26.10-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
ENV GOPRIVATE=github.com/yueli-official/*
RUN --mount=type=secret,id=yueli_git_auth \
    if [ -s /run/secrets/yueli_git_auth ]; then \
        export GIT_CONFIG_COUNT=1 \
            GIT_CONFIG_KEY_0=http.https://github.com/yueli-official/.extraheader \
            GIT_CONFIG_VALUE_0="$(cat /run/secrets/yueli_git_auth)"; \
    fi; \
    go mod download

COPY api ./api
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gallery-api ./cmd/gallery
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gallery-healthcheck ./cmd/healthcheck
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gallery-bindingcheck ./cmd/bindingcheck
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gallery-bootstrap ./cmd/bootstrap
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gallery-imagekeys ./cmd/imagekeys
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gallery-asset-references ./cmd/asset-references

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
COPY --from=build /out/gallery-imagekeys /usr/local/bin/gallery-imagekeys
COPY --from=build /out/gallery-asset-references /usr/local/bin/gallery-asset-references
COPY --chown=nonroot:nonroot manifest ./manifest

ENV GF_SERVER_ADDRESS=0.0.0.0:8091
ENV OTEL_TRACES_EXPORTER=none

EXPOSE 8091
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/gallery-api"]
