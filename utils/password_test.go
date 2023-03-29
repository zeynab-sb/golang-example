package utils

import (
	"errors"
	"github.com/agiledragon/gomonkey/v2"
	_ "github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

type PasswordTestSuite struct {
	suite.Suite
	patch *gomonkey.Patches
}

func (suite *PasswordTestSuite) SetupSuite() {
	suite.patch = gomonkey.NewPatches()
}

func (suite *PasswordTestSuite) TestPassword_ValidatePasswordPattern() {
	require := suite.Require()
	testCases := map[string]struct {
		input       string
		expectedErr error
	}{
		"Valid Password": {
			input:       "Password1234!",
			expectedErr: nil,
		},
		"Less than 8 chars": {
			input:       "passwo",
			expectedErr: errors.New("password should be of 8 characters long"),
		},
		"Not contain capital letter": {
			input:       "password1234!",
			expectedErr: errors.New("password should contain at least one upper case character"),
		},
		"Not contain small letter": {
			input:       "PASSWORD1234!",
			expectedErr: errors.New("password should contain at least one lower case character"),
		},
		"Not contain digit": {
			input:       "PASSWORDaaaa!",
			expectedErr: errors.New("password should contain at least one digit"),
		},
		"Not contain special char": {
			input:       "PASSWORDaaaaa2",
			expectedErr: errors.New("password should contain at least one special character"),
		},
	}

	for desc, v := range testCases {
		suite.Run(desc, func() {
			err := ValidatePasswordPattern(v.input)
			require.Equal(v.expectedErr, err)
		})
	}
}

func (suite *PasswordTestSuite) TestPassword_HashPassword_Failure() {
	require := suite.Require()
	expectedErr := "could not hash password error"

	suite.patch.ApplyFunc(bcrypt.GenerateFromPassword, func(password []byte, cost int) ([]byte, error) {
		return nil, errors.New("error")
	})

	_, err := HashPassword("Password1234!")
	require.Equal(expectedErr, err.Error())
}

func (suite *PasswordTestSuite) TestPassword_HashPassword_Success() {
	require := suite.Require()
	expectedHash := "bvuyrbvuyrbvyr"

	suite.patch.ApplyFunc(bcrypt.GenerateFromPassword, func(password []byte, cost int) ([]byte, error) {
		return []byte("bvuyrbvuyrbvyr"), nil
	})

	hashed, err := HashPassword("Password1234!")
	require.NoError(err)
	require.Equal(expectedHash, hashed)
}

func (suite *PasswordTestSuite) TestPassword_VerifyPassword_Failure() {
	require := suite.Require()
	expectedErr := "error"

	suite.patch.ApplyFunc(bcrypt.CompareHashAndPassword, func(hashedPassword, password []byte) error {
		return errors.New("error")
	})

	err := VerifyPassword("bvuyrbvuyrbvyr", "Password1234!")
	require.Equal(expectedErr, err.Error())
}

func (suite *PasswordTestSuite) TestPassword_VerifyPassword_Success() {
	require := suite.Require()

	suite.patch.ApplyFunc(bcrypt.CompareHashAndPassword, func(hashedPassword, password []byte) error {
		return nil
	})

	err := VerifyPassword("bvuyrbvuyrbvyr", "Password1234!")
	require.NoError(err)
}

func TestPassword(t *testing.T) {
	suite.Run(t, new(PasswordTestSuite))
}
