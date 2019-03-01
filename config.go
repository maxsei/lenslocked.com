package main

import "fmt"

// DefaultPostgresConfig returns a PostgresConfig that has all of the default
// connection info to connect to the db in development
func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "$m0kycat",
		Name:     "lenslocked_dev",
	}
}

// PostgresConfig is a type that turns a json configuration into a
// go struct to be used to configure the postgres database connection
type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// Dialect returns the postgres dialect as a string
func (c PostgresConfig) Dialect() string {
	return "postgres"
}

// ConnectionInfo returns the connections info required by gorm to connect to the
// posgres database
func (c PostgresConfig) ConnectionInfo() string {
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
func DefaultConfig() Config {
	return Config{
		Port:    8080,
		Env:     "dev",
		Pepper:  "nubis",
		HMACKey: "secret-hmac-key",
	}
}

// Config is responsible for configuring aspects of the applications environment
type Config struct {
	Port    int    `json:"port"`
	Env     string `json:"env"`
	Pepper  string `json:"pepper"`
	HMACKey string `json:"hmac_key"`
}

// IsProd looks at the config's Env and if it equals "prod"
func (c Config) InProd() bool {
	return c.Env == "prod"
}

//
// inProd := false
//
// port 8080
// http.ListenAndServe(":8080", csrfMw(userMw.Apply(r)))
//
// # models/users.go
// const userPwPepper = "nubis"
// const hmacSecretKey = "secret-hmac-key"
//
// # models/services.go
// db, err := gorm.Open("postgres", conInfo)
// if err != nil {
//     return nil, err
// }
// db.LogMode(true)
