# Исследование производительности веб-сервиса на Golang

## 1. Инициализируем локальный репозиторий

```shell
git init
```

Также создадим файл `.gitignore`, чтобы некоторые файлы не отслеживались и не загружались в удаленный репозиторий

```text
# .gitignore

.env
.idea
```

## 2. Создадим файл .env с переменными окружения

```text
# .env

PG_HOST=localhost
PG_PORT=5432
PG_USER=postgres
PG_PASSWORD=postgres
PG_DATABASE=postgres
```

Также продублируем эти данные в файл `.env.example`, который будет загружаться в удаленный репозиторий в отличие от оригинального файла `.env` и будет служить примером для заполнения оригинального файла

## 3. Поднимем Postgres с помощью Docker Compose

Предварительно необходимо установить [Docker Desktop](https://www.docker.com/products/docker-desktop/). Далее следует создать файл `docker-compose.yml`, в котором будет приведено описание образов, поднимаемых с помощью Docker Compose.

```yaml
# docker-compose.yaml

services:
  postgres:
    image: postgres:latest
    container_name: postgres
    env_file:
      - ./.env
    environment:
      - POSTGRES_PASSWORD=${PG_PASSWORD}
      - POSTGRES_USER=${PG_USER}
      - POSTGRES_DB=${PG_DATABASE}
    ports:
      - "5432:5432"
    healthcheck:
      test: /usr/bin/pg_isready
      interval: 10s
      timeout: 10s
      retries: 5
    restart: unless-stopped
```

## 4. Создание Makefile для удобного управления нашим приложением

```make
# Makefile

.PHONY: up down

up:
	docker compose -f docker-compose.yaml up -d --build

down:
	docker compose -f docker-compose.yaml down -v
```

## 5. Создание веб-сервиса на Golang

Создадим проект со следующей структурой

```text
netApp
├── Dockerfile
├── cmd
│   └── main.go
├── configs
│   └── config.go
├── entrypoint.sh
├── go.mod
├── go.sum
├── internal
│   └── record
│       ├── handler.go
│       ├── model.go
│       └── repository.go
├── migrations
│   └── auto.go
└── pkg
    ├── db
    │   └── db.go
    ├── request
    │   ├── decode.go
    │   ├── handle.go
    │   └── validate.go
    └── response
        └── response.go

10 directories, 15 files
```

Начнем со вспомогательных фукнций

```go
// netApp/pkg/response/response.go

package response

import (
	"encoding/json"
	"log"
	"net/http"
)

func Json(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Fatalf("Error while encoding response: %v", err)
	}
}
```

```go
// netApp/pkg/request/decode.go

package request

import (
	"encoding/json"
	"io"
)

func Decode[T any](body io.Reader) (T, error) {
	var payload T
	err := json.NewDecoder(body).Decode(&payload)
	if err != nil {
		return payload, err
	}
	return payload, nil
}
```

```go
// netApp/pkg/request/validate.go

package request

import "github.com/go-playground/validator"

func IsValid[T any](payload T) error {
	validate := validator.New()
	err := validate.Struct(payload)
	return err
}
```

```go
// netApp/pkg/request/handle.go

package request

import (
	"classroomWebGolang/pkg/response"
	"net/http"
)

func HandleBody[T any](w *http.ResponseWriter, r *http.Request) (*T, error) {
	body, err := Decode[T](r.Body)
	if err != nil {
		response.Json(*w, err.Error(), http.StatusBadRequest)
		return nil, err
	}
	err = IsValid[T](body)
	if err != nil {
		response.Json(*w, err.Error(), http.StatusBadRequest)
		return nil, err
	}
	return &body, nil
}
```

Перед созданием вспомогательной функции для работы с базой данных следует прописать пакет конфигураций

```go
// netApp/configs/config.go

package configs

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	Db DbConfig
}

type DbConfig struct {
	Dsn string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return &Config{
		Db: DbConfig{
			Dsn: os.Getenv("DB_DSN"),
		},
	}
}
```

```go
// netApp/pkg/db/db.go

package db

import (
	"classroomWebGolang/configs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Db struct {
	*gorm.DB
}

func NewDb(conf *configs.Config) (*Db, error) {
	db, err := gorm.Open(postgres.Open(conf.Db.Dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Db{db}, nil
}
```

Затем пропишем модель

```go
// netApp/internal/record/model.go

package record

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"math/rand"
)

type Record struct {
	ID          uuid.UUID
	Name        string
	Age         int
	Address     string
	PhoneNumber string
	*gorm.Model
}

func NewRecord() *Record {
	return &Record{
		ID:          uuid.New(),
		Name:        faker.Name(),
		Age:         rand.Int(),
		Address:     gofakeit.Address().Address,
		PhoneNumber: gofakeit.Phone(),
	}
}
```

После этого настроим миграции

```go
// netApp/migrations/auto.go

package main

import (
	"classroomWebGolang/internal/record"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db, err := gorm.Open(postgres.Open(os.Getenv("DB_DSN")), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	err = db.AutoMigrate(&record.Record{})
	if err != nil {
		log.Fatalf("Error creating record: %v", err)
	}
}
```

Вслед за этим можем реализовать репозиторий

```go
// netApp/internal/record/repository.go

package record

import "classroomWebGolang/pkg/db"

type RecordRepository struct {
	Database *db.Db
}

func NewRecordRepository(db *db.Db) *RecordRepository {
	return &RecordRepository{Database: db}
}

func (r *RecordRepository) CreateRecord(Record *Record) (*Record, error) {
	result := r.Database.Create(Record)
	if result.Error != nil {
		return nil, result.Error
	}
	return Record, nil
}

func (r *RecordRepository) GetRecords() ([]Record, error) {
	var records []Record
	result := r.Database.Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	return records, nil
}
```

Наконец, можем реализовать обработчики

```go
// netApp/internal/record/handler.go

package record

import (
	"classroomWebGolang/configs"
	"classroomWebGolang/pkg/response"
	"log"
	"net/http"
)

type RecordHandlerDeps struct {
	RecordRepository *RecordRepository
	Config           *configs.Config
}

type RecordHandler struct {
	RecordRepository *RecordRepository
	Config           *configs.Config
}

func NewRecordHandler(router *http.ServeMux, deps *RecordHandlerDeps) {
	handler := &RecordHandler{
		RecordRepository: deps.RecordRepository,
		Config:           deps.Config,
	}

	router.HandleFunc("POST /person", handler.CreateRecord())
	router.HandleFunc("GET /person", handler.GetRecords())
}

func (h *RecordHandler) CreateRecord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("CreateRecord")
		record := NewRecord()
		createRecord, err := h.RecordRepository.CreateRecord(record)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		response.Json(w, createRecord, http.StatusCreated)
	}
}

func (h *RecordHandler) GetRecords() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("GetRecords")
		records, err := h.RecordRepository.GetRecords()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		response.Json(w, records, http.StatusOK)
	}
}
```

Финальный штрих - реализация логики точки входа в приложение

```go
// netApp/cmd/main.go

package main

import (
	"classroomWebGolang/configs"
	"classroomWebGolang/internal/record"
	"classroomWebGolang/pkg/db"
	"log"
	"net/http"
)

func main() {
	conf := configs.LoadConfig()

	db, err := db.NewDb(conf)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("DB_DSN is %s\n", conf.Db.Dsn)

	router := http.NewServeMux()

	recordRepository := record.NewRecordRepository(db)

	record.NewRecordHandler(router, &record.RecordHandlerDeps{RecordRepository: recordRepository, Config: conf})

	server := http.Server{
		Addr:    ":8000",
		Handler: router,
	}

	log.Println("Server is listening on port 8000")
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
```

Теперь осталось прописать файлы `.env` и `.env.example`

```text
# netApp/.env

DB_DSN="host=postgres user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
```

Создать `entrypoint.sh` для запуска начальной миграции

```shell
# netApp/entrypoint.sh

#!/bin/sh

set -e

echo "Running migrations..."
./build/migrate || { echo "Migration failed"; exit 1; }

echo "Starting the application..."
exec ./build/main
```

И последнее - Dockerfile

```shell
# netApp/Dockerfile

FROM golang:alpine

WORKDIR /app

COPY . .

RUN go build -o build/migrate migrations/auto.go
RUN go build -o build/main cmd/main.go

COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/bin/sh", "/app/entrypoint.sh"]
```

Теперь мы можем запустить наш сервис и базу данных

```shell
make up
```

## 6. Проводим нагрузочное тестирование веб-сервиса на Python с помощью [bombardier](https://github.com/codesenberg/bombardier)

```bash
bombardier -c 10 -n 1000 127.0.0.1:8000/person/ -m POST -H 'accept: application/json' -H 'Content-Type: application/json'

#Statistics        Avg      Stdev        Max
#  Reqs/sec      2294.11     546.93    2795.37
#  Latency        4.34ms     2.82ms    25.94ms
#  HTTP codes:
#    1xx - 0, 2xx - 1000, 3xx - 0, 4xx - 0, 5xx - 0
#    others - 0
#  Throughput:     1.21MB/s

bombardier -c 10 -n 1000 127.0.0.1:8000/person/ -m GET -H 'accept: application/json' -H 'Content-Type: application/json'

#Statistics        Avg      Stdev        Max
#  Reqs/sec       574.09     126.59    1058.09
#  Latency       17.46ms     5.48ms    47.51ms
#  HTTP codes:
#    1xx - 0, 2xx - 1000, 3xx - 0, 4xx - 0, 5xx - 0
#    others - 0
#  Throughput:   155.96MB/s
```

# 7. Результаты

| Метрика | POST запрос | GET запрос |
|---------|-------------|------------|
| Reqs/sec (среднее) | 2294.11 | 574.09 |
| Latency (среднее) | 4.34ms | 17.46ms |
| Latency (макс.) | 25.94ms | 47.51ms |
| HTTP коды 2xx | 1000 | 1000 |
| Throughput | 1.21MB/s | 155.96MB/s |
