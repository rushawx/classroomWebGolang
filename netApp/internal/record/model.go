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
