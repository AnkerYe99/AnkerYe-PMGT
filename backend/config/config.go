package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port    int    `yaml:"port"`
		Host    string `yaml:"host"`
	} `yaml:"server"`
	DB struct {
		Path string `yaml:"path"`
	} `yaml:"db"`
	JWT struct {
		Secret string `yaml:"secret"`
		Expire int    `yaml:"expire"` // hours
	} `yaml:"jwt"`
	Upload struct {
		Dir string `yaml:"dir"`
	} `yaml:"upload"`
	Backup struct {
		Key string `yaml:"key"`
	} `yaml:"backup"`
}

var Cfg Config

func Load(path string) error {
	setDefaults()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // use defaults
		}
		return err
	}
	return yaml.Unmarshal(data, &Cfg)
}

func setDefaults() {
	Cfg.Server.Port = 9001
	Cfg.Server.Host = "0.0.0.0"
	Cfg.DB.Path = "/opt/ankerye-pmgt/data.db"
	Cfg.JWT.Secret = "ankerye-pmgt-default-secret-change-me"
	Cfg.JWT.Expire = 72
	Cfg.Upload.Dir = "/opt/ankerye-pmgt/uploads"
	Cfg.Backup.Key = "ankerye-pmgt-backup-aes-key-32byte"
}
