package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
)

type Judge0Client struct {
	client *http.Client
	apiKey string
	host   string
}

var (
	Client *Judge0Client
	once   sync.Once
)

func init() {

	envPath := filepath.Join(os.Getenv("PWD"), ".env")
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("cannot load .env")
	}
	once.Do(func() {
		Client = &Judge0Client{
			client: &http.Client{},
			apiKey: os.Getenv("X_JUDGE0_KEY"),
			host:   os.Getenv("X_JUDGE0_HOST"),
		}
	})
}

func (c *Judge0Client) GET(endpoint string) (*http.Response, error) {
	return c.doRequest("GET", endpoint, nil)
}

func (c *Judge0Client) POST(endpoint string, body interface{}) (*http.Response, error) {
	return c.doRequest("POST", endpoint, body)
}

func (c *Judge0Client) PUT(endpoint string, body interface{}) (*http.Response, error) {
	return c.doRequest("PUT", endpoint, body)
}

func (c *Judge0Client) DELETE(endpoint string) (*http.Response, error) {
	return c.doRequest("DELETE", endpoint, nil)
}

func (c *Judge0Client) doRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	req, err := c.judge0Request(method, endpoint, body)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func (c *Judge0Client) judge0Request(method, endpoint string, body interface{}) (*http.Request, error) {
	var requestBody []byte
	var err error
	if body != nil {
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	url := fmt.Sprintf("https://%s%s", "judge0-ce.p.rapidapi.com", endpoint)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-rapidapi-key", c.apiKey)
	req.Header.Set("x-rapidapi-host", c.host)

	return req, nil
}
