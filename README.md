# L0 LABS (Nats streaming + PostgresSQL)

## **Тестовое задание L0 проекта ЛАБС**

[Видео с демонстрацией работоспособности сервиса](https://drive.google.com/file/d/1Ka_RcJKCzs0z2lBweMyRZ_xAlvgyKWDy/view?usp=sharing)

## Стэк:
- Golang
- Postgres
- Nats streaming
- React
- Docker

## Запуск сервиса:
1. `docker-compose up -d` запускает контейнеры
2. Сервис автоматически соберется и запустится
3. Веб-интерфейс доступен по http://localhost:8080
4. Для отправки тестового сообщения: `go run cmd/publisher/publish.go`