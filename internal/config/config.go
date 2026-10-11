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
    ServerPort    string `mapstructure:"SERVER_PORT"`
    JWTSecret     string `mapstructure:"JWT_SECRET"`
    DBSSLMode     string `mapstructure:"DB_SSL_MODE"`
	EncryptionKey string `mapstructure:"ENCRYPTION_KEY"`
	RedisAddr     string `mapstructure:"REDIS_ADDR"`
}

func LoadConfig() (*Config, error) {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()

    // AutomaticEnv lets GetString read environment variables, but Viper's
    // Unmarshal does not discover them by itself. Bind each application key
    // so values passed by Docker Compose populate Config as well.
    for _, key := range []string{
        "DB_HOST",
        "DB_PORT",
        "DB_USER",
        "DB_PASSWORD",
        "DB_NAME",
        "SERVER_PORT",
        "JWT_SECRET",
        "DB_SSL_MODE",
        "ENCRYPTION_KEY",
        "REDIS_ADDR",
    } {
        if err := viper.BindEnv(key); err != nil {
            return nil, err
        }
    }

    if err := viper.ReadInConfig(); err != nil {
        // It's okay if .env doesn't exist, we might be using system env vars
    }
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    if config.ServerPort == "" {
        config.ServerPort = "8080"
    }

    if config.DBSSLMode == "" {
        config.DBSSLMode = "disable"
    }

    return &config, nil
}
