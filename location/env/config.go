package env

import (
	coreEnv "github.com/caarlos0/env/v9"
)

type Config struct {
	AddrKafka     string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
	AddrRedis     string `env:"REDIS_ADDRESS" envDefault:"localhost:6379"`
	PortServer    string `env:"PORT_SERVER" envDefault:":8888"`
	NumberDbRedis int    `env:"NUMBER_DB_REDIS" envDefault:"0"`
}

func GetConfig() (config Config, err error) {
	cfg := Config{}
	err = coreEnv.Parse(&cfg)

	return cfg, err
}
