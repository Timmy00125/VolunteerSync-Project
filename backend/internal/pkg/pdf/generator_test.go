package pdf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	generator := NewGenerator()
	require.NotNil(t, generator)
	require.NotNil(t, generator.pdf)
}

func TestGenerator_BasicPDFGeneration(t *testing.T) {
	generator := NewGenerator()

	// Set document metadata
	generator.SetTitle("Test Report")
	generator.SetAuthor("Test Author")

	// Add content
	generator.AddHeader("Test Header", "Test Subtitle")
	generator.AddSectionTitle("Test Section")
	generator.AddText("This is test content.")
	generator.AddKeyValue("Key", "Value")

	// Generate PDF
	pdfBytes, err := generator.Output()
	require.NoError(t, err)
	require.NotNil(t, pdfBytes)
	assert.Greater(t, len(pdfBytes), 0, "PDF should have content")

	// PDF files start with %PDF
	assert.Equal(t, []byte("%PDF"), pdfBytes[:4], "Should be a valid PDF file")
}

func TestGenerator_AddMetricsRow(t *testing.T) {
	generator := NewGenerator()

	metrics := []Metric{
		{Label: "Metric 1", Value: "100"},
		{Label: "Metric 2", Value: "200"},
		{Label: "Metric 3", Value: "300"},
	}

	generator.AddMetricsRow(metrics)

	pdfBytes, err := generator.Output()
	require.NoError(t, err)
	assert.Greater(t, len(pdfBytes), 0)
}

func TestGenerator_AddTable(t *testing.T) {
	generator := NewGenerator()

	headers := []string{"Column 1", "Column 2", "Column 3"}
	rows := [][]string{
		{"Row 1 Col 1", "Row 1 Col 2", "Row 1 Col 3"},
		{"Row 2 Col 1", "Row 2 Col 2", "Row 2 Col 3"},
	}
	colWidths := []float64{60, 60, 60}

	generator.AddTable(headers, rows, colWidths)

	pdfBytes, err := generator.Output()
	require.NoError(t, err)
	assert.Greater(t, len(pdfBytes), 0)
}

func TestGenerator_AddSimpleBarChart(t *testing.T) {
	generator := NewGenerator()

	chartData := []ChartData{
		{Label: "Category A", Value: 50},
		{Label: "Category B", Value: 75},
		{Label: "Category C", Value: 100},
	}

	generator.AddSimpleBarChart("Test Chart", chartData, 100)

	pdfBytes, err := generator.Output()
	require.NoError(t, err)
	assert.Greater(t, len(pdfBytes), 0)
}

func TestGenerator_EmptyChart(t *testing.T) {
	generator := NewGenerator()

	// Empty chart should still work
	generator.AddSimpleBarChart("Empty Chart", []ChartData{}, 0)

	pdfBytes, err := generator.Output()
	require.NoError(t, err)
	assert.Greater(t, len(pdfBytes), 0)
}

func TestGenerator_CompletePDFWithAllFeatures(t *testing.T) {
	generator := NewGenerator()

	// Document metadata
	generator.SetTitle("Complete Test Report")
	generator.SetAuthor("Test Suite")

	// Header
	generator.AddHeader("Complete Test Report", "Testing all features")

	// Section 1: Metrics
	generator.AddSectionTitle("Metrics Section")
	metrics := []Metric{
		{Label: "Total", Value: "1000"},
		{Label: "Average", Value: "50"},
	}
	generator.AddMetricsRow(metrics)

	// Divider
	generator.AddDivider()

	// Section 2: Table
	generator.AddSectionTitle("Table Section")
	headers := []string{"Name", "Value"}
	rows := [][]string{
		{"Item 1", "100"},
		{"Item 2", "200"},
	}
	colWidths := []float64{100, 70}
	generator.AddTable(headers, rows, colWidths)

	// Divider
	generator.AddDivider()

	// Section 3: Chart
	generator.AddSectionTitle("Chart Section")
	chartData := []ChartData{
		{Label: "Q1", Value: 25},
		{Label: "Q2", Value: 50},
		{Label: "Q3", Value: 75},
		{Label: "Q4", Value: 100},
	}
	generator.AddSimpleBarChart("Quarterly Data", chartData, 100)

	// Generate PDF
	pdfBytes, err := generator.Output()
	require.NoError(t, err)
	assert.Greater(t, len(pdfBytes), 1000, "Complete PDF should be substantial")
	assert.Equal(t, []byte("%PDF"), pdfBytes[:4])
}

func TestGenerator_MultiPagePDF(t *testing.T) {
	generator := NewGenerator()

	generator.AddHeader("Multi-Page Test", "Testing pagination")

	// Add enough content to trigger multiple pages
	for i := 0; i < 50; i++ {
		generator.AddSectionTitle("Section")
		generator.AddText("This is some text content that should fill up the page.")
		generator.AddKeyValue("Key", "Value")
	}

	pdfBytes, err := generator.Output()
	require.NoError(t, err)
	assert.Greater(t, len(pdfBytes), 5000, "Multi-page PDF should be larger")
}
