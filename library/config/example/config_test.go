package example

import (
	"testing"

	"gopkg.in/go-playground/assert.v1"

	fconfig "github.com/lzw5399/go-common-public/library/config"
)

func TestConfigInit(t *testing.T) {
	cfg := &CustomConfig{}
	fconfig.Init(cfg, ".")

	assert.Equal(t, cfg.ServerName, "uatServerName")
	assert.Equal(t, cfg.MysqlURL, "baseMysqlUrl")
}

type CustomConfig struct {
	*fconfig.Config
}

func (c *CustomConfig) SetBaseConfig(config *fconfig.Config) {
	c.Config = config
}
