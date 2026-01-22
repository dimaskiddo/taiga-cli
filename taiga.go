package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// --- API Response Structures ---
type TaigaClient struct {
	BaseURL    string
	Token      string
	UserID     int
	FullName   string
	HTTPClient *http.Client
}

type AuthResponse struct {
	AuthToken string `json:"auth_token"`
	ID        int    `json:"id"`
	FullName  string `json:"full_name"`
}

type IDResponse struct {
	ID int `json:"id"`
}

type CustomAttr struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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
	c.FullName = res.FullName

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
