package testhelper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

func GetProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	dir := currentDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)

		if parent == dir {
			return "", fmt.Errorf("project root not found, reached filesystem root")
		}

		dir = parent
	}
}

// RenderHTMLWithCSS renders an HTTP response with CSS and saves it as an image
func RenderHTMLWithCSS(t *testing.T, resp *http.Response) {
	t.Helper()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	// Create temporary HTML file
	tempDir, err := os.MkdirTemp("", "visual-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	projectRoot, err := GetProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	// Copy the CSS file to the temp directory to make it accessible to the file server
	cssSourcePath := filepath.Join(projectRoot, "web", "static", "stylesheets", "all.css")

	// Create stylesheets directory in tempDir
	stylesheetsDir := filepath.Join(tempDir, "static", "stylesheets")
	err = os.MkdirAll(stylesheetsDir, 0755)
	assert.NoError(t, err)

	// Copy CSS file to the temp directory structure
	cssDestPath := filepath.Join(stylesheetsDir, "all.css")
	cssContent, err := os.ReadFile(cssSourcePath)
	assert.NoError(t, err)
	err = os.WriteFile(cssDestPath, cssContent, 0644)
	assert.NoError(t, err)

	// Create HTML file with CSS links that are relative to the server root
	htmlContent := []byte("<!DOCTYPE html>\n<html>\n<head>\n")
	htmlContent = append(htmlContent, []byte("  <link rel=\"stylesheet\" href=\"/static/stylesheets/all.css\">\n")...)
	htmlContent = append(htmlContent, []byte("</head>\n<body>\n")...)
	htmlContent = append(htmlContent, body...)
	htmlContent = append(htmlContent, []byte("\n</body>\n</html>")...)

	htmlFile := filepath.Join(tempDir, "index.html")
	err = os.WriteFile(htmlFile, htmlContent, 0644)
	assert.NoError(t, err)

	// Start a local file server
	fileServer := http.FileServer(http.Dir(tempDir))
	server := httptest.NewServer(fileServer)
	defer server.Close()

	// Setup output directory
	outputDir := "test-output"
	err = os.MkdirAll(outputDir, 0755)
	assert.NoError(t, err)

	// Render using chromedp
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var buf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate(server.URL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		// Wait a bit for any CSS transitions or fonts to load
		chromedp.Sleep(2000*time.Millisecond),
		chromedp.FullScreenshot(&buf, 100),
	)
	assert.NoError(t, err)

	// Save screenshot
	outputFile := filepath.Join(outputDir, t.Name()+".png")
	err = os.WriteFile(outputFile, buf, 0644)
	assert.NoError(t, err)
	t.Logf("Visual test output saved to: %s", outputFile)
}
