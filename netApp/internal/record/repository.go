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
