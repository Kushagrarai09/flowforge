package config

import "os"

type Config struct{ DatabaseURL, RedisURL, AMQPURL, Port string }

func Load() Config {
	return Config{env("DATABASE_URL", "postgres://flowforge:flowforge@localhost:5433/flowforge?sslmode=disable"), env("REDIS_URL", "localhost:6379"), env("AMQP_URL", "amqp://guest:guest@localhost:5672/"), env("PORT", "8080")}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
