// Package testcontainers provides reusable test container configurations.
package testcontainers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer wraps a postgres testcontainer with pgvector support.
type PostgresContainer struct {
	Container *postgres.PostgresContainer
	ConnStr   string
}

// NewPostgres starts a PostgreSQL container with pgvector extension.
func NewPostgres(ctx context.Context) (*PostgresContainer, error) {
	container, err := postgres.Run(ctx,
		"pgvector/pgvector:pg16",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	return &PostgresContainer{
		Container: container,
		ConnStr:   connStr,
	}, nil
}

// Terminate stops and removes the container.
func (p *PostgresContainer) Terminate(ctx context.Context) error {
	return p.Container.Terminate(ctx)
}

// OllamaContainer wraps an Ollama testcontainer.
type OllamaContainer struct {
	Container testcontainers.Container
	URL       string
}

// NewOllama starts an Ollama container with model caching.
func NewOllama(ctx context.Context) (*OllamaContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "ollama/ollama:latest",
		ExposedPorts: []string{"11434/tcp"},
		WaitingFor: wait.ForHTTP("/").
			WithPort("11434").
			WithStartupTimeout(60 * time.Second),
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericVolumeMountSource{
					Name: "bobby-ollama-models",
				},
				Target: "/root/.ollama",
			},
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start ollama container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	port, err := container.MappedPort(ctx, "11434")
	if err != nil {
		return nil, fmt.Errorf("failed to get port: %w", err)
	}

	url := fmt.Sprintf("http://%s:%s", host, port.Port())

	return &OllamaContainer{
		Container: container,
		URL:       url,
	}, nil
}

// Terminate stops and removes the container.
func (o *OllamaContainer) Terminate(ctx context.Context) error {
	return o.Container.Terminate(ctx)
}

// PullModel ensures a model is available in the Ollama instance.
func (o *OllamaContainer) PullModel(ctx context.Context, model string) error {
	// Check if model exists
	listResp, err := http.Get(o.URL + "/api/tags")
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}
	defer listResp.Body.Close()

	body, _ := io.ReadAll(listResp.Body)
	if strings.Contains(string(body), model) {
		return nil // Already cached
	}

	// Pull the model
	pullReq := strings.NewReader(fmt.Sprintf(`{"name": "%s"}`, model))
	resp, err := http.Post(o.URL+"/api/pull", "application/json", pullReq)
	if err != nil {
		return fmt.Errorf("failed to pull model: %w", err)
	}
	defer resp.Body.Close()

	// Wait for pull to complete
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 && strings.Contains(string(buf[:n]), "success") {
			return nil
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to read pull response: %w", err)
		}
	}
}
