package model

import (
	"time"
)

type User struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Wagon struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Code      string    `json:"code"`
	Status    string    `json:"status"`
	StartDate time.Time `json:"startDate" sql:"type:timestamp(3)"`
	EndDate   time.Time `json:"endDate" sql:"type:timestamp(3)"`
}
