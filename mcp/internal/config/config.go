package config

type Config struct {
	Neo4j Neo4jConfig
}

type Neo4jConfig struct {
	URI      string
	Username string
	Password string
	Database string
}

