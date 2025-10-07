package registry

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

var Cfg *Config

type Config struct {
	Server struct {
		GRPCPort string
		HTTPPort string
	}
	Redis struct {
		Addr string
	}
	RabbitMQ struct {
		URL            string
		CVRequestQueue string
	}
	Content struct {
		Path string
	}
	Cv struct {
		FilePath string
		Password string
		TokenTTL time.Duration
	}
}

func LoadConfig() (*Config, error) {
	type yamlConfig struct {
		Server struct {
			GRPCPort string `yaml:"grpcPort"`
			HTTPPort string `yaml:"httpPort"`
		} `yaml:"server"`
		Redis struct {
			Addr string `yaml:"addr"`
		} `yaml:"redis"`
		RabbitMQ struct {
			URL            string `yaml:"url"`
			CVRequestQueue string `yaml:"cvRequestQueue"`
		} `yaml:"rabbitmq"`
		Content struct {
			Path string `yaml:"path"`
		} `yaml:"content"`
		Cv struct {
			FilePath string `yaml:"filePath"`
			Password string `yaml:"password"`
			TokenTTL int    `yaml:"tokenTTLSeconds"`
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
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&yc); err != nil {
		return nil, err
	}

	cfg := &Config{}
	cfg.Server.GRPCPort = yc.Server.GRPCPort
	cfg.Server.HTTPPort = yc.Server.HTTPPort
	cfg.Redis.Addr = yc.Redis.Addr
	cfg.RabbitMQ.URL = yc.RabbitMQ.URL
	cfg.RabbitMQ.CVRequestQueue = yc.RabbitMQ.CVRequestQueue
	cfg.Content.Path = yc.Content.Path
	cfg.Cv.FilePath = yc.Cv.FilePath
	cfg.Cv.Password = yc.Cv.Password
	cfg.Cv.TokenTTL = time.Duration(yc.Cv.TokenTTL) * time.Second

	overrideFromEnv("CV_PASSWORD", &cfg.Cv.Password)
	overrideFromEnv("REDIS_ADDR", &cfg.Redis.Addr)
	overrideFromEnv("RABBITMQ_URL", &cfg.RabbitMQ.URL)

	return cfg, nil
}

func overrideFromEnv(envKey string, configValue *string) {
	if value, exists := os.LookupEnv(envKey); exists && value != "" {
		*configValue = value
	}
}
