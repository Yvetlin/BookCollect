# syntax=docker/dockerfile:1

########### BUILD ###########
FROM golang:1.24-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates

# 1) Модули
COPY go.mod .
RUN go mod download

# 2) Исходники (вся папка проекта, т.к. build context = ..)
COPY . .
RUN go mod tidy

# 3) Сборка
ENV CGO_ENABLED=0
RUN go build -o /out/bookcollect ./cmd

# 4) Ресурсы: пустая uploads + КОПИРУЕМ всю web (static + templates)
RUN mkdir -p /out/uploads && touch /out/uploads/.keep
RUN cp -r web /out/

########## RUNTIME ##########
FROM gcr.io/distroless/static-debian12
WORKDIR /app

COPY --from=build /out/bookcollect /app/bookcollect
COPY --from=build /out/web        /app/web
COPY --from=build /out/uploads    /app/uploads

ENV PORT=8080
ENV APP_HTTPS=0
EXPOSE 8080

ENTRYPOINT ["/app/bookcollect"]
