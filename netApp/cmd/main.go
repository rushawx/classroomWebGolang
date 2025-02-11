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
