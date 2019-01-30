package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Users struct {
	gorm.Model
	Name  string
	Email string `gorm:"not null;not null;unique_index"`
}
