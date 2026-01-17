package repository

import (
	"context"
	"fmt"
	"kiwi-user/config"
	"kiwi-user/internal/infrastructure/repository/ent"

	"github.com/futurxlab/golanggraph/logger"
)

type Client struct {
	logger logger.ILogger
	*ent.Client
}

func NewClient(logger logger.ILogger, config *config.Config) (*Client, error) {
	client, err := ent.Open("postgres", config.Postgresql.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed opening connection to postgres, %w", err)
	}

	return &Client{
		logger: logger,
		Client: client,
	}, nil
}

func (c *Client) Close() {
	err := c.Client.Close()
	if err != nil {
		c.logger.Errorf(context.Background(), "failed closing connection to postgres: %w", err)
	}
}
