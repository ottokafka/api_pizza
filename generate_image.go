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

	"image"       // Basic image interface
	"image/jpeg"  // For compressing to JPG
	_ "image/png" // REQUIRED: Registers PNG decoder so image.Decode knows how to read the input
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

// GenerateAndSaveImage calls the API, compresses the result, saves to disk, and returns the path
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

	// 4. Decode Base64 to Bytes
	imgBytes, err := base64.StdEncoding.DecodeString(genResp.Data[0].B64JSON)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 image: %v", err)
	}

	// --- NEW COMPRESSION LOGIC STARTS HERE ---

	// A. Decode bytes into an Image Object
	// We need 'import _ "image/png"' above for this to recognize the format automatically.
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return "", fmt.Errorf("failed to decode image format: %v", err)
	}

	// B. Setup Directory
	uploadDir := "./images"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	// C. Create Filename (Changed extension to .jpg)
	safePrompt := strings.ReplaceAll(prompt, " ", "_")
	if len(safePrompt) > 20 {
		safePrompt = safePrompt[:20]
	}
	filename := fmt.Sprintf("ai_%d_%s.jpg", time.Now().Unix(), safePrompt)
	filePath := filepath.Join(uploadDir, filename)

	// D. Create the file on disk
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// E. Compress and Save as JPEG
	// Quality ranges from 1 to 100.
	// 60-70 is usually the "Sweet Spot" for web (tiny size, decent look).
	// 10 = lowest 12kb quality 100 = highest quaility
	jpegOptions := &jpeg.Options{Quality: 20}

	if err := jpeg.Encode(dst, img, jpegOptions); err != nil {
		return "", fmt.Errorf("failed to encode/compress jpeg: %v", err)
	}

	// --- NEW COMPRESSION LOGIC ENDS HERE ---

	return "./images/" + filename, nil
}
