package database

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

func NewMySQLDBGormMock() (sqlmock.Sqlmock, *gorm.DB) {
	mockDB, sqlMock, err := sqlmock.New()
	if err != nil {
		log.Fatal("error in new connection", zap.Error(err))
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	}))
	if err != nil {
		log.Fatal("error in open connection", zap.Error(err))
	}

	return sqlMock, db
}

func NewRedisMock() (*miniredis.Miniredis, *redis.Client) {
	server, err := miniredis.Run()
	if err != nil {
		log.Fatal(err)
	}

	client := redis.NewClient(&redis.Options{Addr: server.Addr()})

	return server, client
}
