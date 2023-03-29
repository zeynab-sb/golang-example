package database

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang-example/config"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newGormMySQLConnection(
	baseDSN string,
	retry int,
	maxOpenConn int,
	maxIdleConn int,
	retryTimeout time.Duration,
	timeout time.Duration) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	counter := 0
	var id int

	db, err = gorm.Open(mysql.Open(baseDSN))
	if err != nil {
		return nil, fmt.Errorf("cannot open database %s: %s", baseDSN, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("cannot get sql database %s: %s", baseDSN, err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConn)
	sqlDB.SetMaxIdleConns(maxIdleConn)
	sqlDB.SetConnMaxLifetime(timeout)

	counter = 0
	for time.Now(); true; <-time.NewTicker(retryTimeout).C {
		counter++
		err := sqlDB.QueryRow("SELECT connection_id()").Scan(&id)
		if err == nil {
			break
		}

		log.Errorf("Cannot connect to database %s: %s", baseDSN, err)
		if counter >= retry {
			return nil, fmt.Errorf("cannot connect to database %s after %d retries: %s", baseDSN, counter, err)
		}
	}

	log.Info("Connected to mysql database: ", baseDSN)

	return db, nil
}

func InitDatabase() *gorm.DB {
	db, err := newGormMySQLConnection(
		config.C.Database.String(),
		config.C.Database.DialRetry,
		config.C.Database.MaxConn,
		config.C.Database.IdleConn,
		config.C.Database.DialTimeout,
		config.C.Database.Timeout,
	)

	if err != nil {
		log.Errorf("InitDatabase: error in new connection %v", err)
	}

	return db
}
