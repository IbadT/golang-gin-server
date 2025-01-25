# Используем многоэтапную сборку для уменьшения размера итогового образа
FROM golang:1.21-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o todo-app ./cmd/main.go

# Финальный образ
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем собранное приложение
COPY --from=builder /app/todo-app .
COPY .env ./

# Копируем статические файлы (если есть)
COPY ./configs ./configs

# Указываем команду для запуска
CMD ["./todo-app"]
