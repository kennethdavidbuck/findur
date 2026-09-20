# syntax=docker/dockerfile:1.7
FROM golang:1.27.1-alpine3.23 AS backend-build
ARG BUILD_SHA
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN case "$BUILD_SHA" in (*[!0-9a-f]*|'') exit 1;; esac && test "${#BUILD_SHA}" -eq 40
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X github.com/kennethdavidbuck/findur/backend/internal/platform/buildinfo.SHA=${BUILD_SHA}" -o /out/findur ./cmd/findur \
 && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot AS backend
ENV APP_ENV=production PORT=10000 MIGRATIONS_URL=file:///app/db/migrations
WORKDIR /app
COPY --from=backend-build --chown=nonroot:nonroot /out/findur /app/findur
COPY --from=backend-build --chown=nonroot:nonroot /out/migrate /app/migrate
COPY --chown=nonroot:nonroot backend/db/migrations /app/db/migrations
USER nonroot:nonroot
EXPOSE 10000
ENTRYPOINT ["/app/findur"]

FROM node:26.9.0-alpine3.23 AS frontend-build
ARG BUILD_SHA
ENV VITE_BUILD_SHA=${BUILD_SHA} VITE_REQUIRE_BUILD_SHA=true
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN case "$BUILD_SHA" in (*[!0-9a-f]*|'') exit 1;; esac && test "${#BUILD_SHA}" -eq 40 && npm run build

FROM nginx:1.29.5-alpine3.23 AS frontend
COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=frontend-build /src/frontend/dist /usr/share/nginx/html
EXPOSE 8080
