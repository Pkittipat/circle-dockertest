package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
)

var db *sql.DB

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "11",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=user_name",
			"POSTGRES_DB=dbname",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://user_name:secret@%s/dbname?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseUrl)

	resource.Expire(120) // Tell docker to hard kill the container in 120 seconds

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		db, err = sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// Execute SQL commands from schema file
	// _, err = db.Exec(migrations.SchemaSQL)
	// if err != nil {
	// 	log.Fatalf("Failed to execute schema.sql: %s", err)
	// }

	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}()

	// run tests
	m.Run()
}

func TestCartProducts(t *testing.T) {
	// all tests
	// dbx := sqlx.NewDb(db, "postgres")
	// repo := NewRepository(dbx, nil)
	// t.Run("GetCartProducts", func(t *testing.T) {
	// 	cartProeduct, err := repo.GetCartProducts(context.Background(), model.GetCartProductParams{
	// 		UserUUID:   "70533a15-52ca-3134-b95f-836d1ba27c42",
	// 		CartStatus: model.WaitingForCheckingOut,
	// 		Limit:      10,
	// 	})
	// 	assert.NoError(t, err)
	// 	fmt.Printf("cartProduct: %+v", cartProeduct)
	// })
}
