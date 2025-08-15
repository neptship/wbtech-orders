package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Postgres PostgresConfig
	Kafka    KafkaConfig
	HTTP     HTTPConfig
	Cache    CacheConfig
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

type HTTPConfig struct {
	Port string
}

type CacheConfig struct {
	Size int
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() Config {
	_ = godotenv.Load()

	pg := PostgresConfig{
		Host:     getenv("POSTGRES_HOST", "localhost"),
		Port:     parseInt("POSTGRES_PORT", 5432, 1),
		User:     getenv("POSTGRES_USER", "postgres"),
		Password: getenv("POSTGRES_PASSWORD", "postgres"),
		DBName:   getenv("POSTGRES_DB", "orders_data"),
	}

	kafkaCfg := KafkaConfig{
		Brokers: strings.Split(getenv("KAFKA_BROKERS", "localhost:9092"), ","),
		Topic:   getenv("KAFKA_TOPIC", "orders"),
		GroupID: getenv("KAFKA_GROUP_ID", "order-consumer-group"),
	}

	httpCfg := HTTPConfig{
		Port: getenv("HTTP_PORT", "8081"),
	}

	cacheCfg := CacheConfig{
		Size: parseInt("CACHE_SIZE", 1000, 1),
	}

	cfg := Config{Postgres: pg, Kafka: kafkaCfg, HTTP: httpCfg, Cache: cacheCfg}
	validate(&cfg)
	return cfg
}

func parseInt(key string, def, min int) int {
	v := getenv(key, strconv.Itoa(def))
	n, err := strconv.Atoi(v)
	if err != nil || n < min {
		n = def
	}
	return n
}

func validate(cfg *Config) {
	if len(cfg.Kafka.Brokers) == 0 || cfg.Kafka.Brokers[0] == "" {
		log.Println("warning: no Kafka brokers configured")
	}
}

func GetDSN(p PostgresConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", p.User, p.Password, p.Host, p.Port, p.DBName)
}
