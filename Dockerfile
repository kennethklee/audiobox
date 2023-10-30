# Build Backend
# =============
FROM golang:1.21.3-alpine AS gobuilder
ARG VERSION="(untracked)"
ENV PATH="/app:${PATH}"
WORKDIR /app

# Install debug dependencies
RUN apk add --no-cache sqlite \
    && go install github.com/cosmtrek/air@v1.45.0

COPY . .
RUN CGO_ENABLED=0 go build -v -mod=vendor -ldflags "-s -w -X main.Version=$VERSION" -tags timetzdata -o audiobox

# for faster local builds
ENV GOCACHE="/app/.cache"

# development mode
EXPOSE 8090
HEALTHCHECK --start-period=5s --retries=2 --interval=30s CMD audiobox healthcheck
CMD ["/go/bin/air", "-c", "/app/air.toml"]



# Run App
# =======
FROM scratch
ENV PATH="/app:${PATH}"
ENV TMPDIR="/app/pb_data/.tmp"
WORKDIR /app

COPY --from=gobuilder /app/audiobox /app/audiobox

ENV APP_ENV="production"
EXPOSE 8090
VOLUME /app/pb_data
HEALTHCHECK --start-period=5s --retries=2 --interval=30s CMD audiobox healthcheck
CMD ["audiobox", "serve", "--http", "0.0.0.0:8090"]
