package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	// ConfigFilename is the path to the location of the config file which should be at the root directory
	ConfigFilename = ".config.json"
)

// DefaultPostgresConfig returns a PostgresConfig that has all of the default
// connection info to connect to the db in development
func DefaultPostgresConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "$m0kycat",
		Name:     "lenslocked_dev",
	}
}

// DatabaseConfig is a type that turns a json configuration into a
// go struct to be used to configure the postgres database connection
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// Dialect returns the postgres dialect as a string
func (c DatabaseConfig) Dialect() string {
	return "postgres"
}

// ConnectionInfo returns the connections info required by gorm to connect to the
// posgres database
func (c DatabaseConfig) ConnectionInfo() string {
	if c.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
			c.Host, c.Port, c.User, c.Name,
		)
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name,
	)
}

// DefaultConfig returns the default configuration which is the
// host port on 8080 and the environment of the application in development
func DefaultConfig() *Config {
	return &Config{
		Port:     8080,
		Env:      "dev",
		Pepper:   "nubis",
		HMACKey:  "secret-hmac-key",
		Database: DefaultPostgresConfig(),
	}
}

// LoadConfig attempts to load the configuration based of a .config json
// if a configuration is deemed to be required. If the configuration is not required
// or the function fails to load a config file the default configuration is used instead
func LoadConfig(cfgReq bool) (*Config, error) {
	if !cfgReq {
		fmt.Println("using default configuration")
		return DefaultConfig(), nil
	}
	f, err := os.Open(ConfigFilename)
	if err != nil {
		if cfgReq {
			return nil, err
		}
		fmt.Println("no config found using default config")
		return DefaultConfig(), nil
	}
	defer f.Close()

	var cfg Config
	dec := json.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	fmt.Println("successfully loaded config")
	return &cfg, nil
}

// Config is responsible for configuring aspects of the applications environment
type Config struct {
	Port     int            `json:"port"`
	Env      string         `json:"env"`
	Pepper   string         `json:"pepper"`
	HMACKey  string         `json:"hmac_key"`
	Database DatabaseConfig `json:"database"`
}

// InProd looks at the config's Env and if it equals "prod"
func (c Config) InProd() bool {
	return c.Env == "prod"
}
