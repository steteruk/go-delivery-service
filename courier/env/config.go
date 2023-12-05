package env

import (
	coreEnv "github.com/caarlos0/env/v9"
)

type Config struct {
	DBName          string `env:"POSTGRES_DB" envDefault:"courier"`
	DBPassword      string `env:"POSTGRES_PASSWORD" envDefault:"S3cret"`
	DBUser          string `env:"POSTGRES_USER" envDefault:"citizix_user"`
	PortServer      string `env:"PORT_SERVER" envDefault:":8877"`
	CourierGrpcPort string `env:"COURIER_GRPC_PORT" envDefault:":9666"`
}

func GetConfig() (config Config, err error) {
	cfg := Config{}
	err = coreEnv.Parse(&cfg)

	return cfg, err
}
