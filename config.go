// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

import (
	"fmt"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Listen   string `mapstructure:"listen"`
	Address  string `mapstructure:"address"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`

	TTL time.Duration `mapstructure:"ttl"`

	TLS *struct {
		CACert   string `mapstructure:"cacert"`
		Cert     string `mapstructure:"cert"`
		Key      string `mapstructure:"key"`
		Insecure bool   `mapstructure:"insecure"`
	} `mapstructure:"tls"`

	ACL []AccessControlEntry `mapstructure:"acl"`
}

func ParseConfig(flags *flag.FlagSet) (*Config, error) {
	v := viper.NewWithOptions()

	v.SetDefault("username", "admin")
	v.SetDefault("password", "admin")
	v.SetDefault("ttl", DefaultTTL)
	v.SetDefault("listen", ":8080")
	v.SetDefault("tls.insecure", false)

	v.SetConfigType("yaml")

	if f := flags.Lookup("config"); f != nil && f.Value.String() != "" {
		cfgFile := f.Value.String()
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("config")

		v.AddConfigPath("/etc/pdud/")
		v.AddConfigPath("$HOME/.pdud")
		v.AddConfigPath(".")
	}

	v.SetEnvPrefix("pdud")
	v.AutomaticEnv()

	for _, key := range []string{
		"listen",
		"address",
		"username",
		"password",
		"ttl",
		"tls.cacert",
		"tls.cert",
		"tls.key",
		"tls.insecure",
	} {
		flag := strings.ReplaceAll(key, ".", "-")
		v.BindPFlag(key, flags.Lookup(flag))
	}

	if err := v.ReadInConfig(); err != nil { // Handle errors reading the config file
		return nil, fmt.Errorf("fatal error config file: %w", err)
	}

	c := &Config{}

	if err := v.Unmarshal(c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return c, nil
}
