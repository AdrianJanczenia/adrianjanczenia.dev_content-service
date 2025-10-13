package registry

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type ExchangeConfig struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Durable bool   `yaml:"durable"`
}

type QueueConfig struct {
	Name    string `yaml:"name"`
	Durable bool   `yaml:"durable"`
	DLQ     string `yaml:"dlq"`
}

type BindingConfig struct {
	Exchange   string `yaml:"exchange"`
	QueueKey   string `yaml:"queue_key"`
	RoutingKey string `yaml:"routing_key"`
}

type RabbitMQTopologyConfig struct {
	Exchanges []ExchangeConfig       `yaml:"exchanges"`
	Queues    map[string]QueueConfig `yaml:"queues"`
	Bindings  []BindingConfig        `yaml:"bindings"`
}

type Config struct {
	Server struct {
		GRPCPort string
		HTTPPort string
	}
	Redis struct {
		URL string
	}
	RabbitMQ struct {
		URL       string
		Consumers struct {
			DefaultCount int `yaml:"defaultCount"`
		}
		Topology RabbitMQTopologyConfig
	}
	Content struct {
		DefaultLang string
		Files       map[string]string
	}
	Cv struct {
		Password string
		TokenTTL time.Duration
		Files    map[string]string `yaml:"files"`
	}
}

var Cfg *Config

func LoadConfig() (*Config, error) {
	type yamlConfig struct {
		Server struct {
			GRPCPort string `yaml:"grpcPort"`
			HTTPPort string `yaml:"httpPort"`
		} `yaml:"server"`
		Redis struct {
			URL string `yaml:"url"`
		} `yaml:"redis"`
		RabbitMQ struct {
			URL       string `yaml:"url"`
			Consumers struct {
				DefaultCount int `yaml:"defaultCount"`
			} `yaml:"consumers"`
			Topology RabbitMQTopologyConfig `yaml:"topology"`
		} `yaml:"rabbitmq"`
		Content struct {
			DefaultLang string            `yaml:"defaultLang"`
			Files       map[string]string `yaml:"files"`
		} `yaml:"content"`
		Cv struct {
			Password string            `yaml:"password"`
			TokenTTL int               `yaml:"tokenTTLSeconds"`
			Files    map[string]string `yaml:"files"`
		} `yaml:"cv"`
	}

	env := os.Getenv("APP_ENV")
	if env != "production" {
		env = "local"
	}
	configPath := filepath.Join("config", env, "config.yml")
	log.Printf("INFO: loading configuration from %s", configPath)

	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var yc yamlConfig
	if err := yaml.NewDecoder(f).Decode(&yc); err != nil {
		return nil, err
	}

	cfg := &Config{}
	cfg.Server.GRPCPort = yc.Server.GRPCPort
	cfg.Server.HTTPPort = yc.Server.HTTPPort
	cfg.Redis.URL = yc.Redis.URL
	cfg.RabbitMQ.URL = yc.RabbitMQ.URL
	cfg.RabbitMQ.Consumers = yc.RabbitMQ.Consumers
	cfg.RabbitMQ.Topology = yc.RabbitMQ.Topology
	cfg.Content.DefaultLang = yc.Content.DefaultLang
	cfg.Content.Files = yc.Content.Files
	cfg.Cv.Password = yc.Cv.Password
	cfg.Cv.TokenTTL = time.Duration(yc.Cv.TokenTTL) * time.Second
	cfg.Cv.Files = yc.Cv.Files

	overrideFromEnv("CV_PASSWORD", &cfg.Cv.Password)
	overrideFromEnv("REDIS_URL", &cfg.Redis.URL)
	overrideFromEnv("RABBITMQ_URL", &cfg.RabbitMQ.URL)

	return cfg, nil
}

func overrideFromEnv(envKey string, configValue *string) {
	if value, exists := os.LookupEnv(envKey); exists && value != "" {
		*configValue = value
	}
}
