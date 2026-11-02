package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env       string
	LogLevel  string
	LogFormat string

	HTTPAddr           string
	ReadHeaderTimeout  time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	ShutdownTimeout    time.Duration

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	S3Endpoint  string
	S3Region    string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3UseSSL    bool

	JWTSecret          string
	JWTAccessTTL       time.Duration
	JWTRefreshTTL      time.Duration
	SessionIdleTimeout time.Duration
}

func Load() *Config {
	return &Config{
		Env:       getEnv("APP_ENV", "dev"),
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		HTTPAddr:          getEnv("HTTP_ADDR", ":8080"),
		ReadHeaderTimeout: getDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
		ReadTimeout:       getDuration("HTTP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:      getDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:       getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout:   getDuration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "portal"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		S3Endpoint:  getEnv("S3_ENDPOINT", "localhost:9000"),
		S3Region:    getEnv("S3_REGION", "us-east-1"),
		S3AccessKey: getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey: getEnv("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:    getEnv("S3_BUCKET", "materials"),
		S3UseSSL:    getBool("S3_USE_SSL", false),

		JWTSecret:          getEnv("JWT_SECRET", "change-me-in-production"),
		JWTAccessTTL:       getDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:      getDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
		SessionIdleTimeout: getDuration("SESSION_IDLE_TIMEOUT", 30*time.Minute),
	}
}

func (c *Config) Validate() error {
	if c.Env == "prod" && c.JWTSecret == "change-me-in-production" {
		return errors.New("JWT_SECRET must be set in production")
	}
	if c.HTTPAddr == "" {
		return errors.New("HTTP_ADDR is empty")
	}
	return nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func getBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	b, err := strconv.ParseBool(strings.ToLower(v))
	if err != nil {
		return def
	}
	return b
}