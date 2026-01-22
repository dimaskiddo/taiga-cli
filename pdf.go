package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

// --- PDF Fields ---
type ReportTask struct {
	Subject   string
	Date      string
	Time      string
	SpentMins int
	RawLine   string
}

func generatePDFReport(inputPath string, authorName string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("Error could not open log file: %v", err)
	}
	defer file.Close()

	var tasks []ReportTask
	scanner := bufio.NewScanner(file)
	totalMins := 0

	// 1. Read & Parse Created Tasks Log File
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}

		spent, _ := strconv.Atoi(strings.TrimSpace(parts[3]))

		tasks = append(tasks, ReportTask{
			Subject:   strings.TrimSpace(parts[0]),
			Date:      strings.TrimSpace(parts[1]),
			Time:      strings.TrimSpace(parts[2]),
			SpentMins: spent,
		})

		totalMins += spent
	}

	// 2. Sort The Created Tasks By The Activity Date
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Date != tasks[j].Date {
			return tasks[i].Date > tasks[j].Date
		}

		return tasks[i].Time > tasks[j].Time
	})

	// 3. Determine Period
	periodLabel := "Unknown"
	if len(tasks) > 0 {
		// Parse the date string (e.g., "2026-01-02")
		// Since the file is monthly, any task's date is sufficient.
		t, err := time.Parse("2006-01-02", tasks[0].Date)
		if err == nil {
			periodLabel = t.Format("January 2006")
		}
	}

	// 4. Initialize PDF
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Report Title
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(44, 62, 80) // Dark Slate Blue
	pdf.Cell(0, 15, "Worklog Activity Report")
	pdf.Ln(12)

	// Report Metadata
	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(100, 100, 100) // Grey
	pdf.Cell(0, 10, fmt.Sprintf("Author: %s", authorName))
	pdf.Ln(5)
	pdf.Cell(0, 10, fmt.Sprintf("Month: %s", periodLabel))
	pdf.Ln(12)

	// Helper for Table Headers
	drawHeader := func() {
		pdf.SetFont("Arial", "B", 11)
		pdf.SetFillColor(44, 62, 80)    // Dark Blue
		pdf.SetTextColor(255, 255, 255) // White
		pdf.SetDrawColor(200, 200, 200) // Light Grey
		pdf.SetLineWidth(0.3)

		h := 12.0
		// Headers: Subject | Date | Time | Duration
		pdf.CellFormat(100, h, "Activity Subject", "1", 0, "L", true, 0, "")
		pdf.CellFormat(30, h, "Date", "1", 0, "C", true, 0, "")
		pdf.CellFormat(20, h, "Time", "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, h, "Duration", "1", 0, "C", true, 0, "")
		pdf.Ln(-1)
	}

	drawHeader()

	// 5. Process Rows
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(50, 50, 50)
	fill := false

	for _, task := range tasks {
		// Calculate Duration String
		hours := task.SpentMins / 60
		mins := task.SpentMins % 60
		durationFmt := fmt.Sprintf("%dh %02dm", hours, mins)

		// Dynamic Height Calculation
		lineHeight := 6.0
		subjectWidth := 100.0

		lines := pdf.SplitLines([]byte(task.Subject), subjectWidth)
		numLines := float64(len(lines))
		rowHeight := (numLines * lineHeight) + 4.0

		// Page Break Check
		if pdf.GetY()+rowHeight > 270 {
			pdf.AddPage()
			drawHeader()

			pdf.SetFont("Arial", "", 10)
			pdf.SetTextColor(50, 50, 50)
		}

		// Zebra Striping
		if fill {
			pdf.SetFillColor(245, 245, 245)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		currentY := pdf.GetY()
		currentX := pdf.GetX()

		// Draw Subject Cell (First Column)
		pdf.CellFormat(subjectWidth, rowHeight, "", "1", 0, "", true, 0, "")

		// Reset Cursor to Draw Text Inside The Box
		pdf.SetXY(currentX, currentY+2) // +2 Top Padding
		pdf.MultiCell(subjectWidth, lineHeight, task.Subject, "", "L", false)

		// Draw Fixed Columns (Right Side)
		// Move Cursor to The Right of The Subject Column
		pdf.SetXY(currentX+subjectWidth, currentY)

		pdf.CellFormat(30, rowHeight, task.Date, "1", 0, "C", true, 0, "")
		pdf.CellFormat(20, rowHeight, task.Time, "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, rowHeight, durationFmt, "1", 0, "C", true, 0, "")

		// Move to Next Line
		pdf.SetXY(currentX, currentY+rowHeight)

		fill = !fill
	}

	// 6. Total Footer
	pdf.Ln(8)
	pdf.SetFillColor(230, 230, 230)
	pdf.SetFont("Arial", "B", 12)
	totalHours := float64(totalMins) / 60.0

	// Create a Ssummary Box
	pdf.CellFormat(0, 12, fmt.Sprintf("   Total Time Spent: %.2f Hours", totalHours), "1", 1, "L", true, 0, "")

	// 7. Save File
	outName := strings.Replace(inputPath, ".log", ".pdf", 1)
	if !strings.HasSuffix(outName, ".pdf") {
		outName += ".pdf"
	}

	return pdf.OutputFileAndClose(outName)
}
