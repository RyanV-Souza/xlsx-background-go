package main

import (
	"github.com/RyanV-Souza/xlsx-background-go/internal/database"
)

func main() {
	dbConfig := &database.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "docker",
		Password: "docker",
		DBName:   "xlsx",
	}

	_, err := database.Connect(dbConfig)
	if err != nil {
		panic(err)
	}

}
