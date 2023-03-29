package config

import (
	"fmt"
	"net/url"
	"time"
)

type SQLDatabase struct {
	Driver      string        `yaml:"driver"`
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	DB          string        `yaml:"db"`
	User        string        `yaml:"user"`
	Password    string        `yaml:"password"`
	MaxConn     int           `yaml:"max_conn"`
	IdleConn    int           `yaml:"idle_conn"`
	Timeout     time.Duration `yaml:"timeout"`
	DialRetry   int           `yaml:"dial_retry"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
}

type Redis struct {
	DialRetry    int           `yaml:"dial_retry"`
	MaxConn      int           `yaml:"max_conn"`
	IdleConn     int           `yaml:"idle_conn"`
	Address      string        `yaml:"address"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

func (d *SQLDatabase) String() string {
	switch d.Driver {
	case "mysql":
		return d.mysqlDSN()
	case "postgres":
		return d.postgresqlDSN()
	}

	panic("SQLDatabase driver is not supported")
}

func (d *SQLDatabase) mysqlDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true&collation=utf8mb4_general_ci&loc=Asia%%2FTehran", d.User, d.Password, d.Host, d.Port, d.DB)
}

func (d *SQLDatabase) postgresqlDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", d.User, url.QueryEscape(d.Password), d.Host, d.Port, d.DB)
}
