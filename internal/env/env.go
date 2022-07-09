package env

import (
	"github.com/kelseyhightower/envconfig"
)

var Env EnvConfig

type EnvConfig struct {
	HTTPAddress string `default:"0.0.0.0"`
	HTTPPort    uint16 `default:"3000"`
	Root        string `required:"true"`
}

func Process() error {
	return envconfig.Process("filetundra", &Env)
}
