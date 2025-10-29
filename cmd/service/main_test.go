package main

import (
	"testing"
	"order-service/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestAppInitialization(t *testing.T) {
	cfg := config.Load()
	
	assert.NotNil(t, cfg)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, 5433, cfg.DBPort)
}
