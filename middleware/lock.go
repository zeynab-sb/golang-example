package middleware

import (
	"fmt"
	goredis "github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"golang-example/config"
	"net/http"
)

func Lock(redis *goredis.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			keys := ctx.QueryParams()
			userID := ctx.Get(userIDContextField)
			var redisKeys []string
			for k := range keys {
				redisKey := fmt.Sprintf("%s:%v", k, userID)
				redisKeys = append(redisKeys, redisKey)

				_, err := redis.Get(ctx.Request().Context(), redisKey).Result()
				if err == nil {
					return ctx.NoContent(http.StatusTooManyRequests)
				}

				if err != goredis.Nil {
					return err
				}

				err = redis.Set(ctx.Request().Context(), redisKey, userID, config.C.LockTTL).Err()
				if err != nil {
					return err
				}
			}

			nerr := next(ctx)

			for _, key := range redisKeys {
				err := redis.Del(ctx.Request().Context(), key).Err()
				if err != nil {
					log.Errorf("redis delete failed key [%s] : %s", key, err)
				}
			}

			return nerr
		}
	}
}
