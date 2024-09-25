package env

import (
	coreEnv "github.com/caarlos0/env/v9"
)

type Config struct {
	DBName                     string `env:"POSTGRES_DB" envDefault:"orders"`
	DBPassword                 string `env:"POSTGRES_PASSWORD" envDefault:"S3cret"`
	DBUser                     string `env:"POSTGRES_USER" envDefault:"citizix_user"`
	PortServer                 string `env:"PORT_SERVER" envDefault:":8872"`
	CourierGrpcPort            string `env:"COURIER_GRPC_PORT" envDefault:":9671"`
	KafkaAddress               string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
	KafkaSchemaRegistryAddress string `env:"KAFKA_SCHEMA_REGISTRY_ADDRESS" envDefault:"http://localhost:8085"`
	Assignor                   string `env:"KAFKA_CONSUMER_ASSIGNOR" envDefault:"range"`
	Oldest                     bool   `env:"KAFKA_CONSUMER_OLDEST" envDefault:"true"`
	Verbose                    bool   `env:"KAFKA_CONSUMER_VERBOSE" envDefault:"false"`
}

func GetConfig() (config Config, err error) {
	cfg := Config{}
	err = coreEnv.Parse(&cfg)

	return cfg, err
}
