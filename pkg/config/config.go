package config

import (
	"io"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

var (
	once     sync.Once
	instance *Config
)

// LogLevel represents logging verbosity levels
type LogLevel int

const (
	LogTrace LogLevel = iota
	LogDebug
	LogInfo
	LogWarn
	LogError
	LogCritical
)

// Config holds all application configuration
type Config struct {
	Database DatabaseConfig
	Logger   LoggerConfig
	Auth     AuthConfig
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
}

// DSN constructs a PostgreSQL connection string
func (d *DatabaseConfig) DSN() string {
	return "host=" + d.Host +
		" user=" + d.Username +
		" password=" + d.Password +
		" dbname=" + d.Name +
		" port=" + strconv.Itoa(d.Port) +
		" sslmode=" + d.SSLMode +
		" TimeZone=" + d.TimeZone
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level       LogLevel
	Output      io.Writer
	ErrorOutput io.Writer
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	SecretKey string
}

// Get returns the singleton config instance
func Get() *Config {
	once.Do(func() {
		instance = &Config{}
		initialize()
	})
	return instance
}

func initialize() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	instance.Database.Host = os.Getenv("DB_HOST")
	instance.Database.Port = mustAtoi(os.Getenv("DB_PORT"))
	instance.Database.Username = os.Getenv("DB_USERNAME")
	instance.Database.Password = os.Getenv("DB_PASSWORD")
	instance.Database.Name = os.Getenv("DB_NAME")
	instance.Database.SSLMode = getEnvOrDefault("DB_SSLMODE", "disable")
	instance.Database.TimeZone = getEnvOrDefault("DB_TIMEZONE", "UTC")

	instance.Logger.Level = parseLogLevel(os.Getenv("LOGGER_LEVEL"))
	instance.Logger.Output = parseOutput(getEnvOrDefault("LOGGER_OUTPUT", "stdout"))
	instance.Logger.ErrorOutput = parseOutput(getEnvOrDefault("LOGGER_ERROR_OUTPUT", "stderr"))

	instance.Auth.SecretKey = os.Getenv("AUTH_SECRET")
}

func mustAtoi(s string) int {
	if s == "" {
		return 0
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Error converting string to int: %v", err)
	}
	return i
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseLogLevel(level string) LogLevel {
	switch level {
	case "trace":
		return LogTrace
	case "debug":
		return LogDebug
	case "info":
		return LogInfo
	case "warn":
		return LogWarn
	case "error":
		return LogError
	case "critical":
		return LogCritical
	default:
		return LogInfo
	}
}

func parseOutput(output string) io.Writer {
	switch output {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	default:
		file, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Error opening log file: %v", err)
		}
		return file
	}
}
