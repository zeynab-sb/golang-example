package database

import (
	"context"
	"golang-example/config"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

func waitForRedis(client *redis.Client, retry int, retryTimeout time.Duration) {
	counter := 0
	for ; ; <-time.NewTicker(retryTimeout).C {
		counter++
		_, err := client.Ping(context.Background()).Result()
		if err == nil {
			break
		}

		log.Errorf("Cannot connect to redis %s: %s", client.Options().Addr, err)
		if counter >= retry {
			log.Errorf("Cannot connect to redis %s after %d retries: %s", client.Options().Addr, counter, err)
			return
		}
	}
}

func newRedisConnection(
	address, password string,
	db, retry, maxOpenConn, minIdleConn int,
	retryTimeout, readTimeout, writeTimeout time.Duration) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         address,
		Password:     password,
		DB:           db,
		PoolSize:     maxOpenConn,
		MinIdleConns: minIdleConn,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	})
	waitForRedis(client, retry, retryTimeout)
	log.Infof("Connected to redis: %s/%d", address, db)

	return client
}

func CloseRedis(redis *redis.Client) {
	err := redis.Close()
	if err != nil {
		log.Error(err)
	}
}

func InitRedis() *redis.Client {
	return newRedisConnection(
		config.C.Redis.Address,
		config.C.Redis.Password,
		config.C.Redis.DB,
		config.C.Redis.DialRetry,
		config.C.Redis.MaxConn,
		config.C.Redis.IdleConn,
		config.C.Redis.DialTimeout,
		config.C.Redis.ReadTimeout,
		config.C.Redis.WriteTimeout,
	)
}
