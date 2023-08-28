package env

import (
	coreEnv "github.com/caarlos0/env/v9"
)

type Config struct {
	AddrKafka     string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
	AddrRedis     string `env:"REDIS_ADDRESS" envDefault:"localhost:6379"`
	PortServer    string `env:"PORT_SERVER" envDefault:":8888"`
	NumberDbRedis int    `env:"NUMBER_DB_REDIS" envDefault:"0"`
	Assignor      string `env:"KAFKA_CONSUMER_ASSIGNOR" envDefault:"range"`
	Oldest        bool   `env:"KAFKA_CONSUMER_OLDEST" envDefault:"true"`
	Verbose       bool   `env:"KAFKA_CONSUMER_VERBOSE" envDefault:"false"`
	DbName        string `env:"POSTGRES_DB" envDefault:"courier_location"`
	DbPassword    string `env:"POSTGRES_PASSWORD" envDefault:"S3cret"`
	DbUser        string `env:"POSTGRES_USER" envDefault:"citizix_user"`
}

func GetConfig() (config Config, err error) {
	cfg := Config{}
	err = coreEnv.Parse(&cfg)

	return cfg, err
}
