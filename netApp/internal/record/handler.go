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
