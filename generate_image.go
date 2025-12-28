package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Configuration for the Alice / Forest Interactive API
const (
	AliceAPIURL = "https://alice.forest-interactive.com/v1/images/generations"
	AliceAPIKey = "5e59cd1883bdcb8d1afeff7fbc74bfd8f32111c2d3853398a3afb25cf4423376" // <--- REPLACE THIS WITH YOUR ACTUAL KEY
)

type ImageGenRequest struct {
	Prompt string `json:"prompt"`
	Size   string `json:"size"`
}

type ImageGenResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		B64JSON string `json:"b64_json"`
	} `json:"data"`
}

// GenerateAndSaveImage calls the API, decodes the result, saves to disk, and returns the web-accessible path
func GenerateAndSaveImage(prompt string) (string, error) {
	// 1. Prepare Request
	reqBody := ImageGenRequest{
		Prompt: prompt,
		Size:   "512x512",
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", AliceAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+AliceAPIKey)

	// 2. Execute Request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 3. Parse Response
	var genResp ImageGenResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", err
	}

	if len(genResp.Data) == 0 {
		return "", fmt.Errorf("no image data received")
	}

	// 4. Decode Base64
	// The API returns raw base64 data string
	imgBytes, err := base64.StdEncoding.DecodeString(genResp.Data[0].B64JSON)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 image: %v", err)
	}

	// 5. Save to Disk
	uploadDir := "./images"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	// Create a safe filename from prompt or timestamp
	safePrompt := strings.ReplaceAll(prompt, " ", "_")
	if len(safePrompt) > 20 {
		safePrompt = safePrompt[:20]
	}
	filename := fmt.Sprintf("ai_%d_%s.png", time.Now().Unix(), safePrompt)
	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err = dst.Write(imgBytes); err != nil {
		return "", err
	}

	return "./images/" + filename, nil
}
