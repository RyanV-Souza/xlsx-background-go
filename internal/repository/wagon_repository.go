package repository

import (
	"time"

	"github.com/RyanV-Souza/xlsx-background-go/internal/model"
	"gorm.io/gorm"
)

type WagonRepository struct {
	db *gorm.DB
}

func NewWagonRepository(db *gorm.DB) *WagonRepository {
	return &WagonRepository{db: db}
}

func (r *WagonRepository) GetByDateRange(startDate, endDate time.Time) ([]model.Wagon, error) {
	var wagons []model.Wagon
	result := r.db.Where("start_date BETWEEN ? AND ?", startDate, endDate).Find(&wagons)
	if result.Error != nil {
		return nil, result.Error
	}

	return wagons, nil
}
