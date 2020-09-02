package config

import "github.com/caarlos0/env/v6"

type Config struct {
	API       API
	Kafka     Kafka
	Generator Generator
}

type API struct {
	Port string `env:"API_PORT" envDefault:"1001"`
}

type Kafka struct {
	URL string `env:"KAFKA_URL" envDefault:"localhost:9092"`
}

type Generator struct {
	PercentageChangeMin float64 `env:"GENERATOR_PERCENTAGE_CHANGE_MIN" envDefault:"2"`
	PercentageChangeMax float64 `env:"GENERATOR_PERCENTAGE_CHANGE_MAX" envDefault:"5"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
