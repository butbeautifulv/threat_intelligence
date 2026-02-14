package config

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	MongoConfig MongoConfig
}

type MongoConfig struct {
	URI            string `mapstructure:"uri"`             // mongodb://user:pass@host:port
	Host           string `mapstructure:"host"`            // localhost
	Port           int    `mapstructure:"port"`            // 27017
	Username       string `mapstructure:"username"`        // admin
	Password       string `mapstructure:"password"`        // secret
	Database       string `mapstructure:"database"`        // mydb
	AuthSource     string `mapstructure:"auth_source"`     // admin
	MaxPoolSize    uint64 `mapstructure:"max_pool_size"`   // 10–100
	MinPoolSize    uint64 `mapstructure:"min_pool_size"`   // 1–5
	ConnectTimeout int    `mapstructure:"connect_timeout"` // ms
}

func LoadConfig() (*Config, error) {
	// Minimal config loader for local usage. Replace with YAML/env loader as needed.
	return &Config{
		Env: "local",
		MongoConfig: MongoConfig{
			Host:     "localhost",
			Port:     27017,
			Database: "vuln",
		},
	}, nil
}
