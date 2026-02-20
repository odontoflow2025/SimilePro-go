package config

import (
    "github.com/spf13/viper"
)

type Config struct {
    DBHost     string `mapstructure:"DB_HOST"`
    DBPort     string `mapstructure:"DB_PORT"`
    DBUser     string `mapstructure:"DB_USER"`
    DBPassword string `mapstructure:"DB_PASSWORD"`
    DBName     string `mapstructure:"DB_NAME"`
    ServerPort string `mapstructure:"SERVER_PORT"`
    JWTSecret  string `mapstructure:"JWT_SECRET"`
}

func LoadConfig() (*Config, error) {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        // It's okay if .env doesn't exist, we might be using system env vars
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    // Set defaults
    if config.ServerPort == "" {
        config.ServerPort = "8080"
    }

    return &config, nil
}
