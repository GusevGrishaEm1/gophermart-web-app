package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	DBName = "gophermart"
	DBUser = "user"
	DBPass = "user"
)

type TestDatabase struct {
	DBInstance *pgxpool.Pool
	DBAddress  string
	container  testcontainers.Container
}

func SetupTestDatabase() *TestDatabase {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	container, dbInstance, dbAddr, err := createContainer(ctx)
	if err != nil {
		log.Fatal("failed to setup test", err)
	}
	return &TestDatabase{
		container:  container,
		DBInstance: dbInstance,
		DBAddress:  dbAddr,
	}
}

func (tdb *TestDatabase) TearDown() {
	tdb.DBInstance.Close()
	_ = tdb.container.Terminate(context.Background())
}

func createContainer(ctx context.Context) (testcontainers.Container, *pgxpool.Pool, string, error) {

	var env = map[string]string{
		"POSTGRES_PASSWORD": DBPass,
		"POSTGRES_USER":     DBUser,
		"POSTGRES_DB":       DBName,
	}
	var port = "5432/tcp"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:14-alpine",
			ExposedPorts: []string{port},
			Env:          env,
			WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return container, nil, "", fmt.Errorf("failed to start container: %v", err)
	}

	p, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return container, nil, "", fmt.Errorf("failed to get container external port: %v", err)
	}

	log.Println("postgres container ready and running at port: ", p.Port())

	time.Sleep(time.Second)

	dbAddr := fmt.Sprintf("localhost:%s", p.Port())
	db, err := pgxpool.New(ctx, fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", DBUser, DBPass, dbAddr, DBName))
	if err != nil {
		return container, db, dbAddr, fmt.Errorf("failed to establish database connection: %v", err)
	}
	return container, db, dbAddr, nil
}
