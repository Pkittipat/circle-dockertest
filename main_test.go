package main

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPostgresWithTestContainer(t *testing.T) {
	ctx := context.Background()

	// Request a PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	postgresContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer postgresContainer.Terminate(ctx)

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatal(err)
	}

	dsn := fmt.Sprintf(
		"postgres://testuser:testpassword@%s:%s/testdb?sslmode=disable",
		host,
		port.Port(),
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Ensure database connection is alive
	err = db.Ping()
	if err != nil {
		t.Fatal(err)
	}

	// Run your tests here, for example:
	_, err = db.Exec("CREATE TABLE test (id SERIAL PRIMARY KEY, name TEXT NOT NULL)")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO test (name) VALUES ($1)", "John Doe")
	if err != nil {
		t.Fatal(err)
	}

	var name string
	err = db.QueryRow("SELECT name FROM test WHERE id = $1", 1).Scan(&name)
	if err != nil {
		t.Fatal(err)
	}

	if name != "John Doe" {
		t.Fatalf("expected name to be 'John Doe' but got '%s'", name)
	}
}
