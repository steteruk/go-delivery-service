package env

import (
	coreEnv "github.com/caarlos0/env/v9"
)

type Config struct {
	DbName     string `env:"POSTGRES_DB" envDefault:"courier"`
	DbPassword string `env:"POSTGRES_PASSWORD" envDefault:"S3cret"`
	DbUser     string `env:"POSTGRES_USER" envDefault:"citizix_user"`
	PortServer string `env:"PORT_SERVER" envDefault:":8888"`
}

func GetConfig() (config Config, err error) {
	cfg := Config{}
	err = coreEnv.Parse(&cfg)

	return cfg, err
}
