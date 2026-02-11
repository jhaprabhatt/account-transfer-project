package config

import (
	"fmt"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func LoadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     GetEnv("DB_HOST", "localhost"),
		Port:     GetEnv("DB_PORT", "5432"),
		User:     GetEnv("DB_USER", "user"),
		Password: GetEnv("DB_PASSWORD", "password"),
		Name:     GetEnv("DB_NAME", "account_transfer_db"),
	}
}

func (dbConfig DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name)
}
