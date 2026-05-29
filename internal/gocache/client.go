package gocache

import (
	"fmt"
	"net/http"
)

const baseURL = "https://api.gocache.com.br/v1"

type Client struct {
	Token      string
	HTTPClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		Token:      token,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) newRequest(method string, path string) (*http.Request, error) {
	req, err := http.NewRequest(method, baseURL+path, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("GoCache-Token", c.Token)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) ValidateToken() error {
	req, err := c.newRequest("GET", "/plan")

	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("invalid GoCache token, status code: %d", resp.StatusCode)
	}

	return nil
}
