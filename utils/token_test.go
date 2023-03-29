package utils

import (
	"errors"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/suite"
	"golang-example/config"
	"testing"
	"time"
)

type TokenTestSuite struct {
	suite.Suite
	Token *jwt.Token
	patch *gomonkey.Patches
}

func (suite *TokenTestSuite) SetupSuite() {
	suite.patch = gomonkey.NewPatches()
	config.C = config.Config{
		Address:  "",
		Database: config.SQLDatabase{},
		Redis:    config.Redis{},
		Token: config.Token{
			ExpiresIn: time.Minute,
			Secret:    "secret",
		},
		LockTTL: 0,
	}
}

func (suite *TokenTestSuite) TestToken_GenerateToken_Failure() {
	require := suite.Require()
	expectedErr := "generating JWT Token failed: error"

	suite.patch.ApplyMethodReturn(suite.Token, "SignedString", "", errors.New("error"))

	_, err := GenerateToken(1)
	require.Equal(expectedErr, err.Error())
}

func (suite *TokenTestSuite) TestToken_GenerateToken_Success() {
	require := suite.Require()
	expectedToken := "bevyb4v7346vb74bvycbc6734g674bc"

	suite.patch.ApplyMethodReturn(suite.Token, "SignedString", "bevyb4v7346vb74bvycbc6734g674bc", nil)

	token, err := GenerateToken(1)
	require.NoError(err)
	require.Equal(expectedToken, token)
}

func (suite *TokenTestSuite) TestToken_ValidateToken_Failure() {
	require := suite.Require()
	expectedErr := "token contains an invalid number of segments"

	suite.patch.ApplyFunc(jwt.ParseWithClaims, func(tokenString string, claims jwt.Claims, keyFunc jwt.Keyfunc) (*jwt.Token, error) {
		return nil, errors.New("token contains an invalid number of segments")
	})

	_, err := ValidateToken("bevyb4v7346vb74bvycbc6734g674bc")
	require.Equal(expectedErr, err.Error())
}

func (suite *TokenTestSuite) TestToken_ValidateToken_Success() {
	require := suite.Require()

	token, err := GenerateToken(1)

	id, err := ValidateToken(token)
	require.NoError(err)
	require.Equal(uint(1), id)
}

func TestToken(t *testing.T) {
	suite.Run(t, new(TokenTestSuite))
}
