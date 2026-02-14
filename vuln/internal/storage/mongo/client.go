package mongo

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vuln/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClientWrapper struct {
	Client *mongo.Client
	logger *slog.Logger
}

func NewMongoClient(cfg config.MongoConfig, logger *slog.Logger) (*ClientWrapper, error) {
	logger = logger.With(slog.String("component", "mongo"))

	// Если URI указан — используем его
	var uri string
	if cfg.URI != "" {
		uri = cfg.URI
	} else {
		// Build URI depending on whether credentials/authSource are provided
		if cfg.Username != "" && cfg.Password != "" {
			if cfg.AuthSource != "" {
				uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/?authSource=%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.AuthSource)
			} else {
				uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/", cfg.Username, cfg.Password, cfg.Host, cfg.Port)
			}
		} else {
			// No auth
			uri = fmt.Sprintf("mongodb://%s:%d/", cfg.Host, cfg.Port)
		}
	}

	clientOpts := options.Client().ApplyURI(uri)

	if cfg.MaxPoolSize > 0 {
		clientOpts.SetMaxPoolSize(cfg.MaxPoolSize)
	}
	if cfg.MinPoolSize > 0 {
		clientOpts.SetMinPoolSize(cfg.MinPoolSize)
	}

	// Таймаут подключения
	timeout := time.Duration(cfg.ConnectTimeout) * time.Millisecond
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		logger.Error("failed to connect to MongoDB", slog.String("error", err.Error()))
		return nil, err
	}

	// Проверяем соединение
	if err := client.Ping(ctx, nil); err != nil {
		logger.Error("failed to ping MongoDB", slog.String("error", err.Error()))
		return nil, err
	}

	logger.Info("connected to MongoDB")

	return &ClientWrapper{
		Client: client,
		logger: logger,
	}, nil
}

func (c *ClientWrapper) Database(name string) *mongo.Database {
	return c.Client.Database(name)
}

func (c *ClientWrapper) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Client.Disconnect(ctx); err != nil {
		c.logger.Error("failed to disconnect MongoDB", slog.String("error", err.Error()))
		return
	}

	c.logger.Info("MongoDB connection closed")
}

func NewMongo(cfg config.MongoConfig, logger *slog.Logger) (*mongo.Client, error) {
	var uri string
	if cfg.URI != "" {
		uri = cfg.URI
	} else {
		if cfg.Username != "" && cfg.Password != "" {
			if cfg.AuthSource != "" {
				uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/?authSource=%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.AuthSource)
			} else {
				uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/", cfg.Username, cfg.Password, cfg.Host, cfg.Port)
			}
		} else {
			uri = fmt.Sprintf("mongodb://%s:%d/", cfg.Host, cfg.Port)
		}
	}

	opts := options.Client().ApplyURI(uri)
	if cfg.MaxPoolSize > 0 {
		opts.SetMaxPoolSize(cfg.MaxPoolSize)
	}
	if cfg.MinPoolSize > 0 {
		opts.SetMinPoolSize(cfg.MinPoolSize)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnectTimeout)*time.Millisecond)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	return client, nil
}
