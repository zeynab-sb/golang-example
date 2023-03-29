package middleware

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"golang-example/database"
	"net/http"
	"net/http/httptest"
	"testing"

	goredis "github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

func lockNewEchoContext() (echo.Context, *httptest.ResponseRecorder) {
	request := httptest.NewRequest(http.MethodPut, "/metas?gender=male", nil)
	response := httptest.NewRecorder()
	e := echo.New()
	ctx := e.NewContext(request, response)
	ctx.Set(userIDContextField, 1)

	return ctx, response
}

type LockTestSuite struct {
	suite.Suite
	redisServer *miniredis.Miniredis
	redisClient *goredis.Client
	handler     echo.HandlerFunc
}

func (suite *LockTestSuite) SetupSuite() {
	server, client := database.NewRedisMock()

	suite.redisServer = server
	suite.redisClient = client

	suite.handler = func(ctx echo.Context) error {
		return ctx.NoContent(http.StatusOK)
	}
}

func (suite *LockTestSuite) SetupTest() {
	suite.redisClient.FlushAll(context.Background())
}

func (suite *LockTestSuite) TearDownSuite() {
	suite.redisServer.Close()
}

func (suite *LockTestSuite) TestSuccess() {
	require := suite.Require()
	expectedErrorMessage := "ERR no such key"

	ctx, resp := lockNewEchoContext()

	err := Lock(suite.redisClient)(suite.handler)(ctx)
	require.NoError(err)
	require.Equal(resp.Code, http.StatusOK)

	_, err = suite.redisServer.Get("gender:1")
	require.EqualError(err, expectedErrorMessage)
}

func (suite *LockTestSuite) TestTooManyRequests() {
	require := suite.Require()

	ctx, resp := lockNewEchoContext()

	err := suite.redisServer.Set("gender:1", "1")
	require.NoError(err)

	err = Lock(suite.redisClient)(suite.handler)(ctx)
	require.NoError(err)
	require.Equal(http.StatusTooManyRequests, resp.Code)
}

func TestLock(t *testing.T) {
	suite.Run(t, new(LockTestSuite))
}
