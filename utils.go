package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

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
