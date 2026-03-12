package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all the configuration for the application.
type Config struct {
	Environment            string `mapstructure:"ENVIRONMENT"`
	OAuthServerType        string `mapstructure:"OAUTH_SERVER_TYPE"`
	OAuthDomain            string `mapstructure:"OAUTH_DOMAIN"`
	OAuthClientID          string `mapstructure:"OAUTH_CLIENT_ID"`
	OAuthClientSecret      string `mapstructure:"OAUTH_CLIENT_SECRET"`
	OAuthMTLSDomain        string `mapstructure:"OAUTH_MTLS_DOMAIN"`
	OAuthPrivateKeyPath    string `mapstructure:"OAUTH_PRIVATE_KEY_PATH"`
	A2AServerName          string `mapstructure:"A2A_SERVER_NAME"`
	AgentObservabilityLevel int    `mapstructure:"AGENT_OBSERVABILITY_LEVEL"`
}

var globalConfig *Config

// Load initializes the configuration from .env and environment variables.
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv() 

	// Set Defaults
	viper.SetDefault("A2A_SERVER_NAME", "localhost")

	// Allow for OAUTH_ prefixes to be correctly mapped
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: No .env file found, using environment variables")
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	globalConfig = config
	return config, nil
}

// Get returns the loaded configuration.
func Get() *Config {
	if globalConfig == nil {
		_, _ = Load()
	}
	return globalConfig
}
