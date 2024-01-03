package client

import (
	"net/http"
	"time"
)

type Client struct {
	HTTPClient *http.Client
}

func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout: time.Second * 30, // Example timeout
		},
	}
}
