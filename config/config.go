package config

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const DefaultConfigName = "config"
const DefaultTemplateFile = "changelog.tmpl"

type Config struct {
	Path        string
	Tag         string
	DefaultTag  string
	CommitTypes struct {
		Show map[string]string
	}
	Template struct {
		File string
	}
}

func fileExists(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err == nil {
		return !info.IsDir(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func locateTemplateFile(filename string) (string, error) {

	filePath := filename
	exists, _ := fileExists(filePath)
	if exists {
		return filePath, nil
	}

	// Check XDG data dirs
	if file, ok := os.LookupEnv("XDG_DATA_DIR"); ok {
		filePath = file + "/changelister/" + filename
	} else if home, ok := os.LookupEnv("HOME"); ok {
		filePath = home + "/.local/share/changelister/" + filename
	}

	if exists, _ := fileExists(filePath); exists {
		return filePath, nil
	}

	// Not found
	return "", errors.New("template not found")

}

func defineFlags() {
	pflag.String("path", ".", "Path to the git repository")
	pflag.String("tag", "", "Tag to start from (latest if empty)")
	pflag.String("templatefile", DefaultTemplateFile, "Template to use for generating the changelog")
}

func addDefaultConfigDirs() {
	if file, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		viper.AddConfigPath(file)
	} else if home, ok := os.LookupEnv("HOME"); ok {
		viper.AddConfigPath(home + "/.config/changelister")
	}
	viper.AddConfigPath(".")
}

func addEnvironmentConfiguration() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CHANGELISTER_")
	// this is useful e.g. want to use . in Get() calls, but environmental
	// variables to use _ delimiters (e.g. app.port -> APP_PORT)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func LoadConfig() (Config, error) {

	templateFile := DefaultTemplateFile
	configurationName := DefaultConfigName

	defineFlags()
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return Config{}, err
	}
	path := viper.GetString("path")
	tag := viper.GetString("tag")
	templateFile = viper.GetString("templatefile")

	addDefaultConfigDirs()
	viper.SetConfigName(configurationName)
	viper.SetConfigType("yaml")

	addEnvironmentConfiguration()

	viper.SetDefault("tag", "")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		panic(err)
	}

	config.Path = path
	config.Tag = tag

	dataDir, err := locateTemplateFile(templateFile)
	if err != nil {
		return Config{}, err
	}
	config.Template.File = dataDir

	return config, nil
}
