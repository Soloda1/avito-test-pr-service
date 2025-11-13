package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Env        string
	HTTPServer HTTPServer
	Database   Database
}

type HTTPServer struct {
	Address string
	Port    int
}

type Database struct {
	Username       string
	Password       string
	Host           string
	Port           string
	DbName         string
	MigrationsPath string
}

func MustLoad() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	viper.SetDefault("env", "dev")

	viper.SetDefault("http_server.address", "0.0.0.0")
	viper.SetDefault("http_server.port", 8080)

	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "admin")
	viper.SetDefault("database.host", "pr-db")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.db_name", "prservice")
	viper.SetDefault("database.migrations_path", "migrations")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %s", err)
		os.Exit(1)
	}

	config := &Config{
		Env: viper.GetString("env"),
		HTTPServer: HTTPServer{
			Address: viper.GetString("http_server.address"),
			Port:    viper.GetInt("http_server.port"),
		},
		Database: Database{
			Username:       viper.GetString("database.username"),
			Password:       viper.GetString("database.password"),
			Host:           viper.GetString("database.host"),
			Port:           viper.GetString("database.port"),
			DbName:         viper.GetString("database.db_name"),
			MigrationsPath: viper.GetString("database.migrations_path"),
		},
	}

	return config
}
