package detector

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/letusgogo/quick/logger"
)

func NewDefaultOcrDetector(stages []*Stage) StageChecker {
	stageMap := make(map[int]*Stage)
	for _, stage := range stages {
		stageMap[stage.Number] = stage
	}
	return &DefaultOcrDetector{
		stageMap: stageMap,
	}
}

type DefaultOcrDetector struct {
	stageMap map[int]*Stage
}

func (d *DefaultOcrDetector) Detect(ctx context.Context, game string, currentStageNum int, imgBase64 string) (match bool, evidence string, err error) {
	stage, ok := d.stageMap[currentStageNum]
	if !ok {
		return false, "", fmt.Errorf("stage %d not found", currentStageNum)
	}

	// Remove data URL prefix if present (e.g., "data:image/png;base64,")
	base64Data := imgBase64
	if strings.HasPrefix(imgBase64, "data:") {
		// Find the comma that separates the metadata from the base64 data
		commaIndex := strings.Index(imgBase64, ",")
		if commaIndex != -1 {
			base64Data = imgBase64[commaIndex+1:]
		}
	}

	// Decode base64 image
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		logger.Errorf("Error decoding base64 image: %v", err)
		return false, "", fmt.Errorf("failed to decode base64 image: %w", err)
	}

	debugMode := true

	var tempImagePath string

	// Only create and write image file if debug mode is enabled
	if debugMode {
		// Create image file for logging
		logDir := "logging/game_stage_imgs"
		timestamp := time.Now().Unix()
		tempImagePath = filepath.Join(logDir, fmt.Sprintf("cropped_screenshot_%s_%d_%s.png", game, timestamp))

		// Ensure log directory exists
		err = os.MkdirAll(logDir, 0755)
		if err != nil {
			logger.Errorf("Error creating log directory: %v", err)
			return false, "", fmt.Errorf("failed to create log directory: %w", err)
		}

		// Write image data to log file
		err = os.WriteFile(tempImagePath, imageData, 0644)
		if err != nil {
			log.Printf("Error writing image to log file: %v", err)
			return false, "", fmt.Errorf("failed to write image to log file: %w", err)
		}
	} else {
		// In non-debug mode, create a temporary file for OCR processing only
		tempFile, err := os.CreateTemp("", "ocr_temp_*.png")
		if err != nil {
			log.Printf("Error creating temporary file: %v", err)
			return false, "", fmt.Errorf("failed to create temporary file: %w", err)
		}
		defer os.Remove(tempFile.Name()) // Clean up temp file
		tempImagePath = tempFile.Name()

		// Write image data to temporary file
		err = os.WriteFile(tempImagePath, imageData, 0644)
		if err != nil {
			log.Printf("Error writing image to temporary file: %v", err)
			return false, "", fmt.Errorf("failed to write image to temporary file: %w", err)
		}
		tempFile.Close()
	}

	ocrResult, err := runTesseractOCR(tempImagePath, "eng", 6)
	if err != nil {
		return false, "", fmt.Errorf("failed to run tesseract ocr: %w", err)
	}
	if ocrResult == "" {
		return false, "", fmt.Errorf("ocr result is empty")
	}

	match, _, matchedKeyword := analyzeTextForKeywordWithExactMatch(ocrResult, stage.Reco.Matchs)
	if !match {
		return false, "", nil
	}

	return true, matchedKeyword, nil
}

// runTesseractOCR executes Tesseract OCR on the image file
func runTesseractOCR(imagePath string, lang string, psm int) (string, error) {
	// Check if Tesseract is installed
	if !isTesseractInstalled() {
		return "", fmt.Errorf("Tesseract OCR is not installed. Please install tesseract-ocr package")
	}

	// Run Tesseract command
	cmd := exec.Command("tesseract", imagePath, "stdout", "-l", lang, "--psm", fmt.Sprint(psm))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Printf("Tesseract command failed - Error: %v, Stderr: %s", err, stderr.String())
		return "", fmt.Errorf("tesseract command failed: %w, stderr: %s", err, stderr.String())
	}

	result := strings.TrimSpace(stdout.String())

	return result, nil
}

// isTesseractInstalled checks if Tesseract is available in the system
func isTesseractInstalled() bool {
	cmd := exec.Command("tesseract", "--version")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("Tesseract installation check failed - Error: %v, Stderr: %s", err, stderr.String())
		return false
	}
	// Note: This log is kept outside debug mode as it's important for troubleshooting OCR issues
	log.Printf("Tesseract installation check passed")
	return true
}

// analyzeTextForKeywordWithExactMatch analyzes the extracted text for a specific target keyword with exact matching
func analyzeTextForKeywordWithExactMatch(identifiedOCRText string, appKeywords []string) (bool, float64, string) {
	if identifiedOCRText == "" || len(appKeywords) == 0 {
		return false, 0.0, ""
	}

	// Convert text to lowercase for case-insensitive matching
	loweridentifiedOCRText := strings.ToLower(identifiedOCRText)
	loweridentifiedOCRText = strings.ReplaceAll(loweridentifiedOCRText, " ", "")

	// Check for exact match only - the identifiedOCRText must be exactly the same as any keyword
	for _, keyword := range appKeywords {
		lowerKeyword := strings.ToLower(keyword)
		lowerKeyword = strings.ReplaceAll(lowerKeyword, " ", "")

		// Only return true if the texts are exactly the same
		if loweridentifiedOCRText == lowerKeyword {
			matchedKeyword := "keyword_1" // Always return keyword_1 regardless of which keyword matched
			confidence := 1.0             // Maximum confidence for exact match

			return true, confidence, matchedKeyword
		}
	}

	// No exact match found
	return false, 0.0, ""
}
