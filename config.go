package dify

import (
	"net/http"
	"time"
)

type ClientConfig struct {
	Host string
	ApiSecretKey string
	Timeout time.Duration
	Transport *http.Transport
}