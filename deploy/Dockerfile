# syntax=docker/dockerfile:1

############################
#        BUILD STAGE       #
############################
ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src
RUN apk add --no-cache git ca-certificates

# если в go.mod указан более новый тулчейн — пусть докачает
ENV GOTOOLCHAIN=auto
ENV CGO_ENABLED=0

# 1) тянем зависимости (кешируем слой)
COPY go.mod go.sum ./
RUN go mod download

# 2) копируем весь проект (включая web/, internal/, и т.п.)
COPY . .

# 3) собираем бинарь
RUN --mount=type=cache,target=/root/.cache/go-build \
    echo "=== LS /src ===" && ls -la && \
    echo "=== LS /src/cmd ===" && (ls -la cmd || true) && \
    if [ -f main.go ]; then \
        echo "Building from root (main.go found)"; \
        go build -ldflags="-s -w" -o /out/bookcollect ./; \
    elif [ -d cmd ]; then \
        echo "Building from ./cmd/..."; \
        go build -ldflags="-s -w" -o /out/bookcollect ./cmd/...; \
    else \
        echo "No main.go in root and no cmd/ directory. Project layout unexpected."; \
        exit 1; \
    fi

# Копируем статику и шаблоны в артефакты
COPY web /out/web
COPY uploads /out/uploads



# 4) складываем статику в артефакты
RUN mkdir -p /out/web /out/uploads && \
    cp -r web /out/web || true

############################
#       RUNTIME STAGE      #
############################
FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=build /out/bookcollect /app/bookcollect
COPY --from=build /out/web        /app/web
COPY --from=build /out/uploads    /app/uploads


# переменные окружения по умолчанию
ENV PORT=8080
ENV APP_HTTPS=0

EXPOSE 8080
ENTRYPOINT ["/app/bookcollect"]
