package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
)

type Judge0Client struct {
	client *http.Client
	apiKey string
	host   string
}

type RedisClientInterface interface {
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) (string, error)
	Incr(ctx context.Context, key string)
}

type RedisClient struct {
	client *redis.Client
}

var (
	J0Client  *Judge0Client
	RdsClient *RedisClient
	once      sync.Once
)

func init() {
	judge0apikey := os.Getenv("X_JUDGE0_KEY")
	judge0Host := os.Getenv("X_JUDGE0_HOST")
	rdsHost := os.Getenv("RDS_HOST")
	rdsPort := os.Getenv("RDS_PORT")

	log.Printf("%s, %s", rdsHost, rdsPort)

	once.Do(func() {
		RdsClient = &RedisClient{
			client: redis.NewClient(&redis.Options{
				// Addr: fmt.Sprintf("%s:%s", rdsHost, rdsPort),
				Addr: "localhost:6379",
			}),
		}
		J0Client = &Judge0Client{
			client: &http.Client{},
			apiKey: judge0apikey,
			host:   judge0Host,
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

func CacheOperation(redisClientInterface RedisClientInterface, key string, value string) error {
	if err := redisClientInterface.Set(context.Background(), key, value); err != nil {
		log.Printf("failed to set value: %v", err)
		return err
	}
	return nil
}

// func (c *RedisClient) generateGameRoomID() (int64, error) {
// 	ctx := context.Background()
// 	return c.client.Incr(ctx, "game:room:id").Result()
// }

func (c *RedisClient) Get(ctx context.Context, key string) (*string, error) {
	result, err := c.client.Get(ctx, key).Result()
	log.Printf("result => %s", result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *RedisClient) Set(ctx context.Context, key string, value interface{}) error {
	err := c.client.Set(ctx, key, value, 0).Err() // 0 means no expiration
	if err != nil {
		return err
	}
	return nil
}
