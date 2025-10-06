package config

import (
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Path        string
	Tag         string
	CommitTypes struct {
		Accepted []string
		Show     []string
	}
	Template struct {
		File string
	}
}

func LoadConfig() Config {

	pflag.String("path", ".", "Path to the git repository")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	path := viper.GetString("path")

	// Locate file
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.changelistener")

	viper.SetConfigName("changelister")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("CHANGELISTER_")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // this is useful e.g. want to use . in Get() calls, but environmental variables to use _ delimiters (e.g. app.port -> APP_PORT)

	viper.SetDefault("tag", "")
	viper.SetDefault("template.file", "changelog.tmpl")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		panic(err)
	}

	config.Path = path

	return config
}
