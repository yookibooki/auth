package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	SMTP   SMTPConfig
}

type ServerConfig struct {
	Port int
	Host string
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvInt("SERVER_PORT", 8080),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "auth_user"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "auth_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnvInt("SMTP_PORT", 587),
			User:     getEnv("SMTP_USER", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@example.com"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.DB.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.SMTP.Host == "" {
		return fmt.Errorf("SMTP_HOST is required")
	}
	if c.SMTP.User == "" {
		return fmt.Errorf("SMTP_USER is required")
	}
	if c.SMTP.Password == "" {
		return fmt.Errorf("SMTP_PASSWORD is required")
	}
	return nil
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
