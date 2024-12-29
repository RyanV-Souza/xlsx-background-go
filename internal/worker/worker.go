package worker

import (
	"fmt"
	"os"
	"time"

	"github.com/RyanV-Souza/xlsx-background-go/internal/queue"
	"github.com/RyanV-Souza/xlsx-background-go/internal/repository"
	"github.com/go-mail/mail"
	"github.com/xuri/excelize/v2"
)

const TaskGenerateXLSX = "xlsx:generate"

type GenerateXLSXPayload struct {
	UserID    uint       `json:"userId"`
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
}

type Worker struct {
	userRepo  *repository.UserRepository
	wagonRepo *repository.WagonRepository
	rabbitmq  *queue.RabbitMQ
}

func NewWorker(userRepo *repository.UserRepository, wagonRepo *repository.WagonRepository, rabbitmq *queue.RabbitMQ) *Worker {
	return &Worker{
		userRepo:  userRepo,
		wagonRepo: wagonRepo,
		rabbitmq:  rabbitmq,
	}
}

func (w *Worker) Start() error {
	return w.rabbitmq.ConsumeMessages(w.handleGenerateXLSXTask)
}

func (w *Worker) handleGenerateXLSXTask(payload *queue.GenerateXLSXPayload) error {
	user, err := w.userRepo.GetByID(payload.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	wagons, err := w.wagonRepo.GetByDateRange(*payload.StartDate, *payload.EndDate)
	if err != nil {
		return fmt.Errorf("failed to get wagons: %v", err)
	}

	f := excelize.NewFile()
	defer f.Close()

	f.DeleteSheet("Sheet1")

	userSheet := "User Info"
	f.NewSheet(userSheet)
	f.SetCellValue(userSheet, "A1", "User ID")
	f.SetCellValue(userSheet, "B1", "Name")
	f.SetCellValue(userSheet, "C1", "Email")

	f.SetCellValue(userSheet, "A2", user.ID)
	f.SetCellValue(userSheet, "B2", user.Name)
	f.SetCellValue(userSheet, "C2", user.Email)

	wagonSheet := "Wagons"
	f.NewSheet(wagonSheet)
	f.SetCellValue(wagonSheet, "A1", "Wagon ID")
	f.SetCellValue(wagonSheet, "B1", "Code")
	f.SetCellValue(wagonSheet, "C1", "Status")
	f.SetCellValue(wagonSheet, "D1", "Start Date")
	f.SetCellValue(wagonSheet, "E1", "End Date")

	for i, wagon := range wagons {
		row := i + 2
		f.SetCellValue(wagonSheet, fmt.Sprintf("A%d", row), wagon.ID)
		f.SetCellValue(wagonSheet, fmt.Sprintf("B%d", row), wagon.Code)
		f.SetCellValue(wagonSheet, fmt.Sprintf("C%d", row), wagon.Status)
		f.SetCellValue(wagonSheet, fmt.Sprintf("D%d", row), wagon.StartDate.Format("2006-01-02 15:04:05"))
		f.SetCellValue(wagonSheet, fmt.Sprintf("E%d", row), wagon.EndDate.Format("2006-01-02 15:04:05"))
	}

	fileName := fmt.Sprintf("report_%d_%s.xlsx", user.ID, time.Now().Format("20060102150405"))
	if err := f.SaveAs(fileName); err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}
	defer func() {
		fmt.Println("Removing file...")
		if err := os.Remove(fileName); err != nil {
			fmt.Printf("Failed to remove file: %v\n", err)
		}
	}()

	if err := sendEmail(user.Email, fileName); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func sendEmail(to string, filePath string) error {
	email := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")

	m := mail.NewMessage()
	m.SetHeader("From", email)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Your XLSX Report")
	m.SetBody("text/plain", "Please find your requested report attached.")
	m.Attach(filePath)

	d := mail.NewDialer("smtp.gmail.com", 587, email, password)
	return d.DialAndSend(m)
}
