package database

import (
	"fmt"
	"game-knowledge/backend/models"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() error {
	database, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		return err
	}

	err = database.AutoMigrate(&models.User{}, &models.Match{}, &models.Prediction{})
	if err != nil {
		fmt.Println("AutoMigrate failed, error occurred: ", err)
		return err
	}
	DB = database
	return nil
}
