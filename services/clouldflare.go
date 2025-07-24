// rpg-api/internal/services/cloudflare.go
package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type CloudflareImagesService struct {
	AccountID    string
	APIToken     string
	BaseURL      string
	DeliveryHash string // <-- Make sure this line is there
}

type CloudflareImageResponse struct {
	Result struct {
		ID       string            `json:"id"`
		Filename string            `json:"filename"`
		Variants []string          `json:"variants"`
		Meta     map[string]string `json:"meta"`
	} `json:"result"`
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func NewCloudflareImagesService() *CloudflareImagesService {
	return &CloudflareImagesService{
		AccountID:    os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
		APIToken:     os.Getenv("CLOUDFLARE_API_TOKEN"),
		BaseURL:      "https://api.cloudflare.com/client/v4",
		DeliveryHash: os.Getenv("CLOUDFLARE_DELIVERY_HASH"), // <-- And this line
	}
}

func (s *CloudflareImagesService) UploadImage(file multipart.File, filename string, metadata map[string]string) (*CloudflareImageResponse, error) {
	// Create multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add file
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(fw, file)
	if err != nil {
		return nil, err
	}
	// Add metadata if provided
	if metadata != nil {
		for key, value := range metadata {
			w.WriteField(key, value)
		}
	}
	w.Close()
	// Create request
	url := fmt.Sprintf("%s/accounts/%s/images/v1", s.BaseURL, s.AccountID)
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.APIToken)
	req.Header.Set("Content-Type", w.FormDataContentType())
	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Parse response
	var cfResp CloudflareImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return nil, err
	}
	if !cfResp.Success {
		if len(cfResp.Errors) > 0 {
			return nil, fmt.Errorf("cloudflare error: %s", cfResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("unknown cloudflare error")
	}
	return &cfResp, nil
}

func (s *CloudflareImagesService) GetImageURL(imageID string, variant string) string {
	if variant == "" {
		variant = "public"
	}
	return fmt.Sprintf("https://imagedelivery.net/%s/%s/%s", s.DeliveryHash, imageID, variant)
}

func (s *CloudflareImagesService) DeleteImage(imageID string) error {
	url := fmt.Sprintf("%s/accounts/%s/images/v1/%s", s.BaseURL, s.AccountID, imageID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.APIToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete image: status %d", resp.StatusCode)
	}

	return nil
}
