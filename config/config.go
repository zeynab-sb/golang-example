package config

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var (
	// C represents config of the project loaded into a struct
	C Config

	V *viper.Viper
)

var builtinConfig = []byte(`address: 0.0.0.0:8080
database:
  driver: mysql
  host: localhost
  port: 3306
  db: server
  user: server
  password: server
  max_conn: 10
  idle_conn: 5
  timeout: 10s
  dial_retry: 12
  dial_timeout: 5s
redis:
  dial_retry: 12
  max_conn: 10
  idle_conn: 5
  address: 'localhost:6379'
  password: ''
  db: 0
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s
token:
  expires_in: 5m
  secret: secret
loc_ttl: 30s
`)

type Config struct {
	Address  string        `yaml:"address"`
	Database SQLDatabase   `yaml:"database"`
	Redis    Redis         `yaml:"redis"`
	Token    Token         `yaml:"token"`
	LockTTL  time.Duration `yaml:"loc_ttl"`
}

type Token struct {
	ExpiresIn time.Duration `yaml:"expires_in"`
	Secret    string        `yaml:"secret"`
}

func initViper(path string, c *Config) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if err := v.ReadConfig(bytes.NewReader(builtinConfig)); err != nil {
		return nil, fmt.Errorf("loading builtin config failed: %s", err)
	}

	if path != "" {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("opening config [%s] failed: %s", path, err)
		}

		configFiles := make([]string, 0)
		if info.IsDir() {
			configFiles, err = filepath.Glob(filepath.Join(path, "*config.yml"))
			if err != nil {
				return nil, fmt.Errorf("loading config failed: %s", err)
			}
		} else {
			configFiles = append(configFiles, path)
		}

		for _, f := range configFiles {
			v.SetConfigFile(f)
			if err := v.MergeInConfig(); err != nil {
				return nil, fmt.Errorf("opening config file [%s] failed: %s", f, err)
			} else {
				log.Infof("config file [%s] opened successfully", f)
			}
		}
	}

	err := v.Unmarshal(c, func(config *mapstructure.DecoderConfig) {
		config.TagName = "yaml"

		config.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)
	})
	if err != nil {
		return nil, fmt.Errorf("failed on config unmarshal: %s", err)
	}

	return v, nil
}

func Init(path string) *Config {
	var err error
	c := Config{}
	V, err = initViper(path, &c)
	if err != nil {
		log.Fatal("Failed on config initialization: %v", err)
	}

	C = c
	return &c
}
