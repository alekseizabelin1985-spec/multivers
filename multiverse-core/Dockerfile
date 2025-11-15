# build/Dockerfile

# ========== Стадия сборки ==========
FROM golang:1.25-alpine AS builder
ARG SERVICE
RUN echo $SERVICE
# Установка git для go mod download (может потребоваться)
RUN apk add --no-cache git

WORKDIR /app

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .
RUN echo ${SERVICE}
RUN ls
RUN echo "FileLogger: end"

# Собираем бинарник *без* CGO
# Это безопасно, так как мы отключили CGO-зависимости через replace
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/${SERVICE} ./cmd/${SERVICE}

# ───────────────────────

# Можно вернуться к Alpine, так как CGO больше не нужен
FROM alpine:latest
ARG SERVICE
RUN echo $SERVICE
# Установка только необходимых пакетов
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /bin/${SERVICE}  /${SERVICE}
RUN chmod +x /${SERVICE}

CMD ["/${SERVICE}"]