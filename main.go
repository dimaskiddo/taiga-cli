package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// --- Configuration & Types ---
type Config struct {
	TaigaURL    string
	User        string
	Password    string
	ProjectSlug string
}

func main() {
	// 1. Setup Environment
	err := godotenv.Load()
	if err != nil {
		logError("Error: .env file not found or could not be loaded")
		os.Exit(1)
	}

	conf := Config{
		TaigaURL:    os.Getenv("TAIGA_URL"),
		User:        os.Getenv("TAIGA_USER"),
		Password:    os.Getenv("TAIGA_PASSWORD"),
		ProjectSlug: os.Getenv("PROJECT_SLUG"),
	}

	taiga := &TaigaClient{
		BaseURL:    conf.TaigaURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}

	// 2. Authentication
	err = taiga.Login(conf.User, conf.Password)
	if err != nil {
		logError(fmt.Sprintf("Authentication failed. Please check your credentials: %v", err))
		os.Exit(1)
	}
	fmt.Println("Authentication successful")

	if len(os.Args) > 2 && (os.Args[1] == "-pdf" || os.Args[1] == "--pdf") {
		logFile := os.Args[2]

		err := generatePDFReport(logFile, taiga.FullName)
		if err != nil {
			fmt.Printf("Error generating PDF report: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("PDF report generated successfully")
		return
	}

	if len(os.Args) < 2 {
		logError("Error: No input file provided")
		logError("Usage: ./taiga-cli <input_file>")
		os.Exit(1)
	}

	inputFile := os.Args[1]

	// 3. Resolve Project and Status IDs
	projectID, err := taiga.GetProjectID(conf.ProjectSlug)
	if err != nil {
		logError(fmt.Sprintf("Failed to fetch Project ID from Project Slug %s: %v", conf.ProjectSlug, err))
		os.Exit(1)
	}

	statusDoneID, err := taiga.GetStatusID(projectID, "Done")
	if err != nil {
		logError(fmt.Sprintf("Failed to fetch Status for Done ID: %v", err))
		os.Exit(1)
	}

	// 4. Process File
	processInputFile(taiga, inputFile, projectID, statusDoneID)
	fmt.Println("All tasks done. Check logs for details.")
}

func processInputFile(c *TaigaClient, path string, projectID, statusID int) {
	file, err := os.Open(path)
	if err != nil {
		logError(fmt.Sprintf("Error: Input file '%s' not found!", path))
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	if !scanner.Scan() {
		logError("Error: Input file is empty")
		os.Exit(1)
	}

	storyRef := strings.TrimSpace(scanner.Text())
	storyID, err := c.GetStoryID(storyRef, projectID)
	if err != nil {
		logError(fmt.Sprintf("Failed to fetch Story ID for Story Reference ID %s: %v", storyRef, err))
		os.Exit(1)
	}

	_ = os.Mkdir("logs", 0755)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			logError(fmt.Sprintf("Invalid line format: %s", line))
			continue
		}

		subject := strings.TrimSpace(parts[0])
		date := strings.TrimSpace(parts[1])
		startTime := strings.TrimSpace(parts[2])
		spent := strings.TrimSpace(parts[3])

		logName := fmt.Sprintf("logs/created_tasks_%s.log", date[:7])
		if isDuplicate(logName, line) {
			fmt.Printf("Task '%s' already created, skipping.\n", subject)
			continue
		}

		taskID, err := c.CreateTask(subject, projectID, storyID, statusID)
		if err != nil {
			logError(fmt.Sprintf("Failed to create task: %s | Error: %v", subject, err))
			continue
		}

		err = c.UpdateCustomAttributes(taskID, projectID, date, startTime, spent)
		if err != nil {
			logError(fmt.Sprintf("Failed to update custom fields for task ID: %d | Error: %v", taskID, err))
			continue
		}

		appendLog(logName, line)
		fmt.Printf("Subtask '%s' created and the custom fields updated.\n", subject)
	}
}
