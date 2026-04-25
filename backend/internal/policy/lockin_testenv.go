package policy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupTestEnv spins up Postgres and Kafka containers for integration tests
func SetupTestEnv(t *testing.T) (db *gorm.DB, kafkaAddr string, cleanup func()) {
	ctx := context.Background()

	// Start Postgres container
	pgC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_PASSWORD": "testpass",
				"POSTGRES_DB":       "testdb",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)

	pgHost, _ := pgC.Host(ctx)
	pgPort, _ := pgC.MappedPort(ctx, "5432")
	pgDSN := "host=" + pgHost + " port=" + pgPort.Port() + " user=postgres password=testpass dbname=testdb sslmode=disable"
	db, err = gorm.Open(postgres.Open(pgDSN), &gorm.Config{})
	require.NoError(t, err)

	// Start Kafka container
	kafkaC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "confluentinc/cp-kafka:7.2.1",
			ExposedPorts: []string{"9092/tcp"},
			Env: map[string]string{
				"KAFKA_BROKER_ID":                        "1",
				"KAFKA_ZOOKEEPER_CONNECT":                "zookeeper:2181",
				"KAFKA_LISTENERS":                        "PLAINTEXT://0.0.0.0:9092",
				"KAFKA_ADVERTISED_LISTENERS":             "PLAINTEXT://localhost:9092",
				"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
			},
			WaitingFor: wait.ForListeningPort("9092/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)

	kafkaHost, _ := kafkaC.Host(ctx)
	kafkaPort, _ := kafkaC.MappedPort(ctx, "9092")
	kafkaAddr = kafkaHost + ":" + kafkaPort.Port()

	cleanup = func() {
		_ = pgC.Terminate(ctx)
		_ = kafkaC.Terminate(ctx)
	}

	return db, kafkaAddr, cleanup
}
