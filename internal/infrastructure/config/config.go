package config

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env        string
	HTTPServer HTTPServer
	Database   Database
}

type HTTPServer struct {
	Address        string
	Port           int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	RequestTimeout time.Duration
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
	viper.SetDefault("http_server.read_timeout", "10s")
	viper.SetDefault("http_server.write_timeout", "10s")
	viper.SetDefault("http_server.idle_timeout", "60s")
	viper.SetDefault("http_server.request_timeout", "10s")

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
			Address:        viper.GetString("http_server.address"),
			Port:           viper.GetInt("http_server.port"),
			ReadTimeout:    viper.GetDuration("http_server.read_timeout"),
			WriteTimeout:   viper.GetDuration("http_server.write_timeout"),
			IdleTimeout:    viper.GetDuration("http_server.idle_timeout"),
			RequestTimeout: viper.GetDuration("http_server.request_timeout"),
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
