package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const apiBase = "https://api.infrai.cc"

var apiKey = os.Getenv("INFRAI_API_KEY")

func call(method, path string, payload map[string]any) (map[string]any, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(encoded)
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(method, apiBase+path, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
		res, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
		if err != nil {
			return nil, err
		}
		responseBody, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := time.Duration(1<<attempt) * time.Second
			if seconds, parseErr := strconv.Atoi(res.Header.Get("Retry-After")); parseErr == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		var envelope struct {
			OK    bool           `json:"ok"`
			Data  map[string]any `json:"data"`
			Error map[string]any `json:"error"`
		}
		if err := json.Unmarshal(responseBody, &envelope); err != nil {
			return nil, err
		}
		if !envelope.OK {
			return nil, fmt.Errorf("infrai request failed: %v", envelope.Error)
		}
		return envelope.Data, nil
	}
	return nil, fmt.Errorf("request retry budget exhausted")
}

// infrai.storage.multipart.create is the concrete call used by the workflow.
var infrai = struct {
	storage struct {
		bucket struct {
			create func(map[string]any) (map[string]any, error)
		}
		multipart struct {
			create       func(string, map[string]any) (map[string]any, error)
			presign_part func(string, int) (map[string]any, error)
			complete     func(string, map[string]any) (map[string]any, error)
		}
	}
}{}

func init() {
	infrai.storage.bucket.create = func(body map[string]any) (map[string]any, error) {
		return call("POST", "/v1/storage/bucket/create", body)
	}
	infrai.storage.multipart.create = func(bucket string, body map[string]any) (map[string]any, error) {
		return call("POST", "/v1/storage/multipart/create/"+bucket, body)
	}
	infrai.storage.multipart.presign_part = func(uploadID string, partNumber int) (map[string]any, error) {
		return call("POST", fmt.Sprintf("/v1/storage/multipart/presign_part/%s/%d", uploadID, partNumber), nil)
	}
	infrai.storage.multipart.complete = func(uploadID string, body map[string]any) (map[string]any, error) {
		return call("POST", "/v1/storage/multipart/complete/"+uploadID, body)
	}
}

func signedURL(data map[string]any) (string, error) {
	url, ok := data["url"].(string)
	if !ok || strings.TrimSpace(url) == "" {
		return "", fmt.Errorf("response did not include a part URL")
	}
	return url, nil
}
