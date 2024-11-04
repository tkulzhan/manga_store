package databases

import (
	"context"
	"manga_store/internal/helpers"
	"manga_store/internal/logger"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var neo4jDriver neo4j.DriverWithContext

func InitNeo4j() {
	neo4jUri := helpers.GetEnv("NEO4J_URI", "neo4j://localhost:7687")
	neo4jUsername := helpers.GetEnv("NEO4J_USERNAME", "neo4j")
	neo4jPassword := helpers.GetEnv("NEO4J_PASSWORD", "neo4jpassword")

	var err error
	neo4jDriver, err = neo4j.NewDriverWithContext(neo4jUri, neo4j.BasicAuth(neo4jUsername, neo4jPassword, ""))
	if err != nil {
		logger.Error("Failed to connect to Neo4j: " + err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = neo4jDriver.VerifyConnectivity(ctx)
	if err != nil {
		logger.Error("Neo4j connectivity verification failed: " + err.Error())
	} else {
		logger.Info("Connected to Neo4j")
	}
}

func Neo4j(ctx context.Context) neo4j.SessionWithContext {
	return neo4jDriver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
}

func CloseNeo4j(ctx context.Context) {
	if neo4jDriver != nil {
		neo4jDriver.Close(ctx)
	}
}
