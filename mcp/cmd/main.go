package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mcp/internal/components"
	"mcp/internal/config"
)

func main() {
	var (
		env = flag.String("env", envOr("MCP_ENV", "local"), "local|dev|prod")

		neo4jURI  = flag.String("neo4j-uri", envOr("NEO4J_URI", "neo4j://localhost:7687"), "neo4j uri")
		neo4jUser = flag.String("neo4j-user", envOr("NEO4J_USER", "neo4j"), "neo4j username")
		neo4jPass = flag.String("neo4j-pass", envOr("NEO4J_PASS", "neo4jpassword"), "neo4j password")
		neo4jDB   = flag.String("neo4j-db", envOr("NEO4J_DB", "neo4j"), "neo4j database")
	)
	flag.Parse()

	logger := components.SetupLogger(*env)

	cfg := &config.Config{
		Neo4j: config.Neo4jConfig{
			URI:      *neo4jURI,
			Username: *neo4jUser,
			Password: *neo4jPass,
			Database: *neo4jDB,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigQuit
		cancel()
	}()

	c, err := components.InitComponents(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Shutdown()

	// MCP server runs over stdio. On context cancel, exit.
	go func() {
		<-ctx.Done()
		os.Exit(0)
	}()

	if err := c.MCPServer.Run(ctx, os.Stdin, os.Stdout); err != nil {
		logger.Error("mcp server stopped", slog.Any("err", err))
		time.Sleep(200 * time.Millisecond)
		os.Exit(1)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

