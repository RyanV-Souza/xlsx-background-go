package http

import (
	"net/http"
	"time"

	"github.com/RyanV-Souza/xlsx-background-go/internal/queue"
	"github.com/RyanV-Souza/xlsx-background-go/internal/repository"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router          *gin.Engine
	userRepository  *repository.UserRepository
	wagonRepository *repository.WagonRepository
	rabbitmq        *queue.RabbitMQ
}

type GenerateXLSXRequest struct {
	UserID    uint       `json:"userId" binding:"required"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

func NewServer(rabbitmq *queue.RabbitMQ) *Server {
	router := gin.Default()

	server := &Server{
		router:   router,
		rabbitmq: rabbitmq,
	}

	router.POST("/generate", server.generateXlsx)

	return server
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) generateXlsx(c *gin.Context) {
	var req GenerateXLSXRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload := &queue.GenerateXLSXPayload{
		UserID:    req.UserID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	if err := s.rabbitmq.PublishMessage(c.Request.Context(), payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue task"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "XLSX generation started. You will receive an email when it's ready.",
	})
}
