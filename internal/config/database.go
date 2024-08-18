package config

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func GetDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("database"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	return db
}
