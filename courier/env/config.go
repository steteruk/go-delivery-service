package env

import (
	coreEnv "github.com/caarlos0/env/v9"
)

type Config struct {
	DBName                string `env:"POSTGRES_DB" envDefault:"courier"`
	DBPassword            string `env:"POSTGRES_PASSWORD" envDefault:"S3cret"`
	DBUser                string `env:"POSTGRES_USER" envDefault:"citizix_user"`
	PortServer            string `env:"PORT_SERVER" envDefault:":8881"`
	CourierGrpcPort       string `env:"COURIER_GRPC_PORT" envDefault:":9666"`
	AssignCourierGrpcPort string `env:"ASSIGN_COURIER_GRPC_PORT" envDefault:":9671"`
	AddrKafka             string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
	Assignor              string `env:"KAFKA_CONSUMER_ASSIGNOR" envDefault:"range"`
	Oldest                bool   `env:"KAFKA_CONSUMER_OLDEST" envDefault:"true"`
	Verbose               bool   `env:"KAFKA_CONSUMER_VERBOSE" envDefault:"false"`
}

func GetConfig() (config Config, err error) {
	cfg := Config{}
	err = coreEnv.Parse(&cfg)

	return cfg, err
}
