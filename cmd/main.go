package main

import (
	"log"

	"github.com/RyanV-Souza/xlsx-background-go/http"
	"github.com/RyanV-Souza/xlsx-background-go/internal/database"
	"github.com/RyanV-Souza/xlsx-background-go/internal/queue"
	"github.com/RyanV-Souza/xlsx-background-go/internal/repository"
	"github.com/RyanV-Souza/xlsx-background-go/internal/worker"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Failed to load .env file:", err)
	}

	dbConfig := &database.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "docker",
		Password: "docker",
		DBName:   "xlsx",
	}

	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	rabbitmq, err := queue.NewRabbitMQ("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer rabbitmq.Close()

	userRepo := repository.NewUserRepository(db)
	wagonRepo := repository.NewWagonRepository(db)

	worker := worker.NewWorker(userRepo, wagonRepo, rabbitmq)
	go func() {
		if err := worker.Start(); err != nil {
			log.Fatal("Failed to start worker:", err)
		}
	}()

	server := http.NewServer(rabbitmq)
	if err := server.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
