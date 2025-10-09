package pdf

import (
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// Generator provides utilities for PDF generation
type Generator struct {
	pdf *gofpdf.Fpdf
}

// NewGenerator creates a new PDF generator with standard settings
func NewGenerator() *Generator {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	return &Generator{
		pdf: pdf,
	}
}

// SetTitle sets the document title
func (g *Generator) SetTitle(title string) {
	g.pdf.SetTitle(title, false)
}

// SetAuthor sets the document author
func (g *Generator) SetAuthor(author string) {
	g.pdf.SetAuthor(author, false)
}

// AddHeader adds a header with title and subtitle
func (g *Generator) AddHeader(title, subtitle string) {
	g.pdf.SetFont("Arial", "B", 20)
	g.pdf.SetTextColor(41, 128, 185) // Blue color
	g.pdf.CellFormat(0, 10, title, "", 1, "C", false, 0, "")

	if subtitle != "" {
		g.pdf.SetFont("Arial", "", 12)
		g.pdf.SetTextColor(127, 140, 141) // Gray color
		g.pdf.CellFormat(0, 6, subtitle, "", 1, "C", false, 0, "")
	}

	g.pdf.Ln(5)
	g.pdf.SetTextColor(0, 0, 0) // Reset to black
}

// AddSectionTitle adds a section title
func (g *Generator) AddSectionTitle(title string) {
	g.pdf.Ln(5)
	g.pdf.SetFont("Arial", "B", 14)
	g.pdf.SetTextColor(52, 73, 94) // Dark blue
	g.pdf.CellFormat(0, 8, title, "", 1, "L", false, 0, "")
	g.pdf.SetTextColor(0, 0, 0) // Reset to black
	g.pdf.Ln(2)
}

// AddText adds regular text content
func (g *Generator) AddText(text string) {
	g.pdf.SetFont("Arial", "", 11)
	g.pdf.MultiCell(0, 6, text, "", "L", false)
}

// AddKeyValue adds a key-value pair with bold key
func (g *Generator) AddKeyValue(key, value string) {
	g.pdf.SetFont("Arial", "B", 11)
	g.pdf.Cell(50, 6, key+":")
	g.pdf.SetFont("Arial", "", 11)
	g.pdf.Cell(0, 6, value)
	g.pdf.Ln(6)
}

// AddTable adds a table with headers and rows
func (g *Generator) AddTable(headers []string, rows [][]string, colWidths []float64) {
	// Header
	g.pdf.SetFont("Arial", "B", 11)
	g.pdf.SetFillColor(41, 128, 185)  // Blue background
	g.pdf.SetTextColor(255, 255, 255) // White text

	for i, header := range headers {
		width := colWidths[i]
		g.pdf.CellFormat(width, 8, header, "1", 0, "C", true, 0, "")
	}
	g.pdf.Ln(-1)

	// Reset colors for rows
	g.pdf.SetTextColor(0, 0, 0)
	g.pdf.SetFont("Arial", "", 10)

	// Rows with alternating colors
	fill := false
	for _, row := range rows {
		if fill {
			g.pdf.SetFillColor(240, 240, 240) // Light gray
		} else {
			g.pdf.SetFillColor(255, 255, 255) // White
		}

		for i, cell := range row {
			width := colWidths[i]
			g.pdf.CellFormat(width, 7, cell, "1", 0, "L", true, 0, "")
		}
		g.pdf.Ln(-1)
		fill = !fill
	}

	g.pdf.Ln(3)
}

// AddMetricBox adds a metric box with label and value
func (g *Generator) AddMetricBox(label, value string, x, y, width, height float64) {
	g.pdf.SetXY(x, y)

	// Draw box with border
	g.pdf.SetFillColor(236, 240, 241) // Light gray background
	g.pdf.Rect(x, y, width, height, "FD")

	// Add label
	g.pdf.SetXY(x+5, y+5)
	g.pdf.SetFont("Arial", "", 10)
	g.pdf.SetTextColor(127, 140, 141) // Gray
	g.pdf.Cell(width-10, 5, label)

	// Add value
	g.pdf.SetXY(x+5, y+12)
	g.pdf.SetFont("Arial", "B", 16)
	g.pdf.SetTextColor(41, 128, 185) // Blue
	g.pdf.Cell(width-10, 8, value)

	// Reset colors
	g.pdf.SetTextColor(0, 0, 0)
}

// AddMetricsRow adds a row of metric boxes
func (g *Generator) AddMetricsRow(metrics []Metric) {
	numMetrics := len(metrics)
	if numMetrics == 0 {
		return
	}

	pageWidth := 210.0 // A4 width in mm
	margin := 20.0
	spacing := 5.0
	availableWidth := pageWidth - (2 * margin)
	boxWidth := (availableWidth - (float64(numMetrics-1) * spacing)) / float64(numMetrics)
	boxHeight := 25.0

	_, currentY := g.pdf.GetXY()

	for i, metric := range metrics {
		x := margin + (float64(i) * (boxWidth + spacing))
		g.AddMetricBox(metric.Label, metric.Value, x, currentY, boxWidth, boxHeight)
	}

	g.pdf.SetY(currentY + boxHeight + 8)
}

// Metric represents a metric with label and value
type Metric struct {
	Label string
	Value string
}

// AddSimpleBarChart adds a simple horizontal bar chart
func (g *Generator) AddSimpleBarChart(title string, data []ChartData, maxValue float64) {
	g.AddSectionTitle(title)

	if len(data) == 0 {
		g.AddText("No data available")
		return
	}

	barHeight := 8.0
	maxBarWidth := 120.0
	labelWidth := 60.0
	startX := 20.0

	_, currentY := g.pdf.GetXY()

	for i, item := range data {
		y := currentY + (float64(i) * (barHeight + 3))

		// Label
		g.pdf.SetXY(startX, y)
		g.pdf.SetFont("Arial", "", 10)
		g.pdf.Cell(labelWidth, barHeight, item.Label)

		// Bar
		barWidth := (item.Value / maxValue) * maxBarWidth
		if barWidth > 0 {
			g.pdf.SetFillColor(41, 128, 185) // Blue
			g.pdf.Rect(startX+labelWidth, y, barWidth, barHeight, "F")
		}

		// Value
		g.pdf.SetXY(startX+labelWidth+barWidth+2, y)
		g.pdf.Cell(20, barHeight, fmt.Sprintf("%.1f", item.Value))
	}

	g.pdf.SetY(currentY + (float64(len(data)) * (barHeight + 3)) + 5)
}

// ChartData represents data for a chart
type ChartData struct {
	Label string
	Value float64
}

// AddFooter adds a footer with generation date and page numbers
func (g *Generator) AddFooter(generatedDate time.Time) {
	g.pdf.SetFooterFunc(func() {
		g.pdf.SetY(-15)
		g.pdf.SetFont("Arial", "I", 8)
		g.pdf.SetTextColor(127, 140, 141)

		// Date on left
		g.pdf.Cell(0, 10, fmt.Sprintf("Generated on %s", generatedDate.Format("January 2, 2006")))

		// Page number on right
		g.pdf.SetX(-20)
		g.pdf.Cell(0, 10, fmt.Sprintf("Page %d", g.pdf.PageNo()))
	})
}

// AddDivider adds a horizontal line divider
func (g *Generator) AddDivider() {
	g.pdf.Ln(3)
	g.pdf.SetDrawColor(189, 195, 199) // Light gray
	g.pdf.Line(20, g.pdf.GetY(), 190, g.pdf.GetY())
	g.pdf.Ln(5)
	g.pdf.SetDrawColor(0, 0, 0) // Reset to black
}

// Output returns the PDF as a byte slice
func (g *Generator) Output() ([]byte, error) {
	var buf []byte
	writer := &byteWriter{buf: &buf}
	err := g.pdf.Output(writer)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return buf, nil
}

// GetPDF returns the underlying gofpdf instance for advanced operations
func (g *Generator) GetPDF() *gofpdf.Fpdf {
	return g.pdf
}

// byteWriter implements io.Writer for capturing PDF output
type byteWriter struct {
	buf *[]byte
}

func (w *byteWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}
