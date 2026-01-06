package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

type TaigaClient struct {
	BaseURL    string
	Token      string
	UserID     int
	HTTPClient *http.Client
}

// --- API Response Structures ---
type AuthResponse struct {
	AuthToken string `json:"auth_token"`
	ID        int    `json:"id"`
}

type IDResponse struct {
	ID int `json:"id"`
}

type CustomAttr struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	if len(os.Args) < 2 {
		logError("Error: No input file provided")
		logError("Usage: ./taiga-cli <input_file>")
		os.Exit(1)
	}

	inputFile := os.Args[1]

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

// --- API Helper Methods ---
func (c *TaigaClient) request(method, path string, payload interface{}, target interface{}) error {
	var body io.Reader
	if payload != nil {
		b, _ := json.Marshal(payload)
		body = bytes.NewBuffer(b)
	}

	req, _ := http.NewRequest(method, c.BaseURL+path, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("status %d: %s", res.StatusCode, string(bodyBytes))
	}

	if target != nil {
		return json.NewDecoder(res.Body).Decode(target)
	}

	return nil
}

func (c *TaigaClient) Login(user, pass string) error {
	payload := map[string]string{"type": "normal", "username": user, "password": pass}

	var res AuthResponse
	err := c.request("POST", "/api/v1/auth", payload, &res)
	if err != nil {
		return err
	}

	c.Token = res.AuthToken
	c.UserID = res.ID

	return nil
}

func (c *TaigaClient) GetProjectID(slug string) (int, error) {
	var res IDResponse

	err := c.request("GET", "/api/v1/projects/by_slug?slug="+slug, nil, &res)

	return res.ID, err
}

func (c *TaigaClient) GetStoryID(ref string, projectID int) (int, error) {
	var res IDResponse

	url := fmt.Sprintf("/api/v1/userstories/by_ref?ref=%s&project=%d", ref, projectID)
	err := c.request("GET", url, nil, &res)

	return res.ID, err
}

func (c *TaigaClient) GetStatusID(projectID int, name string) (int, error) {
	var list []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	err := c.request("GET", fmt.Sprintf("/api/v1/task-statuses?project=%d", projectID), nil, &list)
	if err != nil {
		return 0, err
	}

	for _, s := range list {
		if s.Name == name {
			return s.ID, nil
		}
	}

	return 0, fmt.Errorf("status %s not found", name)
}

func (c *TaigaClient) CreateTask(subject string, projectID, storyID, statusID int) (int, error) {
	payload := map[string]interface{}{
		"subject":     subject,
		"assigned_to": c.UserID,
		"status":      statusID,
		"project":     projectID,
		"user_story":  storyID,
		"is_blocked":  false,
		"is_closed":   true,
	}

	var res IDResponse
	err := c.request("POST", "/api/v1/tasks", payload, &res)

	return res.ID, err
}

func (c *TaigaClient) UpdateCustomAttributes(taskID, projectID int, date, start, spent string) error {
	var attrs []CustomAttr
	err := c.request("GET", fmt.Sprintf("/api/v1/task-custom-attributes?project=%d", projectID), nil, &attrs)
	if err != nil {
		return err
	}

	fieldMap := make(map[string]string)
	for _, a := range attrs {
		switch a.Name {
		case "Activity Date":
			fieldMap[fmt.Sprint(a.ID)] = date
		case "Start Time":
			fieldMap[fmt.Sprint(a.ID)] = start
		case "Total Time Spent":
			fieldMap[fmt.Sprint(a.ID)] = spent
		}
	}

	payload := map[string]interface{}{
		"attributes_values": fieldMap,
		"version":           1,
	}

	return c.request("PATCH", fmt.Sprintf("/api/v1/tasks/custom-attributes-values/%d", taskID), payload, nil)
}

// --- Utils ---
func isDuplicate(filepath, entry string) bool {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return false
	}

	return strings.Contains(string(content), strings.TrimSpace(entry))
}

func appendLog(filepath, entry string) {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString(entry + "\n")
}

func logError(message string) {
	_ = os.Mkdir("logs", 0755)

	f, err := os.OpenFile("logs/error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open error log: %v\n", err)
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] %s\n", timestamp, message)

	fmt.Fprint(os.Stderr, logLine)
	f.WriteString(logLine)
}
