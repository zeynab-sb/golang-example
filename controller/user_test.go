package controller

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"golang-example/config"
	"golang-example/database"
	"golang-example/utils"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
)

type SignupTestSuite struct {
	suite.Suite
	e        *echo.Echo
	endpoint string
	ctx      context.Context
	sqlMock  sqlmock.Sqlmock
	patch    *gomonkey.Patches
	user     User
}

func (suite *SignupTestSuite) SetupSuite() {
	sqlMock, db := database.NewMySQLDBGormMock()
	suite.sqlMock = sqlMock

	mockCtrl := gomock.NewController(suite.T())
	defer mockCtrl.Finish()

	suite.e = echo.New()
	suite.endpoint = "/signup"
	suite.ctx = context.Background()
	suite.user = User{DB: db}
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
	suite.patch = gomonkey.NewPatches()
}

func (suite *SignupTestSuite) TearDownSuit() {
	suite.patch.Reset()

	sqlDB, _ := suite.user.DB.DB()
	_ = sqlDB.Close()
}

func (suite *SignupTestSuite) TearDownTest() {
	suite.patch.Reset()
}

func (suite *SignupTestSuite) CallHandler(requestBody string) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodPost, suite.endpoint, strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.e.NewContext(req, rec)
	err := suite.user.Signup(c)

	return rec, err
}

func (suite *SignupTestSuite) TestSignup_Signup_Binding_Failure() {
	require := suite.Require()
	expectedError := "code=400, message=error in parse request data"

	requestBody := `{"user_name:"use"}`
	_, err := suite.CallHandler(requestBody)

	require.EqualError(err, expectedError)
}

func (suite *SignupTestSuite) TestSignup_Signup_InvalidUserName_InvalidLength_Failure() {
	require := suite.Require()
	expectedError := "code=400, message=username is invalid"

	requestBody := `{"user_name":"use"}`
	_, err := suite.CallHandler(requestBody)

	require.EqualError(err, expectedError)
}

func (suite *SignupTestSuite) TestSignup_Signup_InvalidUserName_InvalidPattern_Failure() {
	require := suite.Require()
	expectedError := "code=400, message=username is invalid"

	requestBody := `{"user_name":"ddddddduseق"}`
	_, err := suite.CallHandler(requestBody)

	require.EqualError(err, expectedError)
}

func (suite *SignupTestSuite) TestSignup_Signup_InvalidPassword_Failure() {
	require := suite.Require()
	expectedError := "code=400, message=password isn't strong enough"

	suite.patch.ApplyFunc(utils.ValidatePasswordPattern, func(password string) error {
		return errors.New("password should contain at least one special character")
	})

	requestBody := `{"user_name":"username","password":"za12"}`
	_, err := suite.CallHandler(requestBody)

	require.EqualError(err, expectedError)
}

func (suite *SignupTestSuite) TestSignup_Signup_FindUserNameDBError_Failure() {
	require := suite.Require()
	expectedError := errors.New("database error")

	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`user_name` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs("username").
		WillReturnError(errors.New("database error"))

	suite.patch.ApplyFunc(utils.ValidatePasswordPattern, func(password string) error {
		return nil
	})

	requestBody := `{"user_name":"username","password":"Aaaaaaaa768!"}`
	_, err := suite.CallHandler(requestBody)

	require.EqualError(err, expectedError.Error())
}

func (suite *SignupTestSuite) TestSignup_Signup_FindUserNameDB_RecordFoundErr_Failure() {
	require := suite.Require()
	expectedError := "code=400, message=username is already taken"

	rows := sqlmock.NewRows([]string{"id", "user_name"}).
		AddRow(1, "username")
	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`user_name` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs("username").
		WillReturnRows(rows)

	suite.patch.ApplyFunc(utils.ValidatePasswordPattern, func(password string) error {
		return nil
	})

	requestBody := `{"user_name":"username","password":"Aaaaaaaa768!"}`
	_, err := suite.CallHandler(requestBody)

	require.EqualError(err, expectedError)
}

func (suite *SignupTestSuite) TestSignup_Signup_HashPassword_Failure() {
	require := suite.Require()
	expectedError := "code=500, message=Internal Server Error"

	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`user_name` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs("username").
		WillReturnError(gorm.ErrRecordNotFound)

	suite.patch.ApplyFunc(utils.ValidatePasswordPattern, func(password string) error {
		return nil
	})

	suite.patch.ApplyFunc(utils.HashPassword, func(password string) (string, error) {
		return "", errors.New("error")
	})

	requestBody := `{"user_name":"username","password":"Aaaaaaaa768!"}`
	_, err := suite.CallHandler(requestBody)

	require.EqualError(err, expectedError)
}

func (suite *SignupTestSuite) TestSignup_Signup_CreateUserDBErr_Failure() {
	require := suite.Require()
	expectedError := "code=500, message=Internal Server Error"

	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`user_name` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs("username").
		WillReturnError(gorm.ErrRecordNotFound)

	suite.patch.ApplyFunc(utils.ValidatePasswordPattern, func(password string) error {
		return nil
	})

	suite.patch.ApplyFunc(utils.HashPassword, func(password string) (string, error) {
		return "$2a$10$wBDhXmJfiZ9nskiXAijWre1PB8htQBEPhkxRgFPHkK0dQUm65nBIu", nil
	})

	suite.sqlMock.ExpectBegin()
	syntax = "^INSERT INTO `users`"
	suite.sqlMock.ExpectExec(syntax).
		WillReturnError(errors.New("database error"))
	suite.sqlMock.ExpectRollback()

	requestBody := `{"user_name":"username","password":"Aaaaaaaa768!"}`
	_, err := suite.CallHandler(requestBody)

	require.EqualError(err, expectedError)
}

func (suite *SignupTestSuite) TestSignup_Signup_GenerateToken_Failure() {
	require := suite.Require()
	expectedError := "code=500, message=Internal Server Error"

	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`user_name` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs("username").
		WillReturnError(gorm.ErrRecordNotFound)

	suite.patch.ApplyFunc(utils.ValidatePasswordPattern, func(password string) error {
		return nil
	})

	suite.patch.ApplyFunc(utils.HashPassword, func(password string) (string, error) {
		return "$2a$10$wBDhXmJfiZ9nskiXAijWre1PB8htQBEPhkxRgFPHkK0dQUm65nBIu", nil
	})

	suite.patch.ApplyFunc(utils.GenerateToken, func(id uint) (string, error) {
		return "", errors.New("error")
	})

	suite.sqlMock.ExpectBegin()
	syntax = "^INSERT INTO `users`"
	suite.sqlMock.ExpectExec(syntax).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.sqlMock.ExpectCommit()

	requestBody := `{"user_name":"username","password":"Aaaaaaaa768!"}`
	_, err := suite.CallHandler(requestBody)

	require.EqualError(err, expectedError)
}

func (suite *SignupTestSuite) TestSignup_Signup_Success() {
	require := suite.Require()
	expectedMsg := `{"status":"success", "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1Nzc4MzY4NjAsImlhdCI6MTU3NzgzNjgwMCwibmJmIjoxNTc3ODM2ODAwLCJzdWIiOjF9.NFfVAWrBMyWHMvpmnKR7hKLngPTtf7NObp9J9kn78G8"}`

	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`user_name` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs("username").
		WillReturnError(gorm.ErrRecordNotFound)

	suite.patch.ApplyFunc(utils.ValidatePasswordPattern, func(password string) error {
		return nil
	})

	suite.patch.ApplyFunc(utils.HashPassword, func(password string) (string, error) {
		return "$2a$10$wBDhXmJfiZ9nskiXAijWre1PB8htQBEPhkxRgFPHkK0dQUm65nBIu", nil
	})

	suite.patch.ApplyFunc(utils.GenerateToken, func(id uint) (string, error) {
		return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1Nzc4MzY4NjAsImlhdCI6MTU3NzgzNjgwMCwibmJmIjoxNTc3ODM2ODAwLCJzdWIiOjF9.NFfVAWrBMyWHMvpmnKR7hKLngPTtf7NObp9J9kn78G8", nil
	})

	suite.sqlMock.ExpectBegin()
	syntax = "^INSERT INTO `users`"
	suite.sqlMock.ExpectExec(syntax).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.sqlMock.ExpectCommit()

	requestBody := `{"user_name":"username","password":"Aaaaaaaa768!"}`
	response, err := suite.CallHandler(requestBody)

	require.NoError(err)
	require.Equal(response.Code, http.StatusCreated)
	require.JSONEq(expectedMsg, response.Body.String())
}

type LoginTestSuite struct {
	suite.Suite
	e        *echo.Echo
	endpoint string
	ctx      context.Context
	sqlMock  sqlmock.Sqlmock
	patch    *gomonkey.Patches
	user     User
}

func (suite *LoginTestSuite) SetupSuite() {
	sqlMock, db := database.NewMySQLDBGormMock()
	suite.sqlMock = sqlMock

	mockCtrl := gomock.NewController(suite.T())
	defer mockCtrl.Finish()

	suite.e = echo.New()
	suite.endpoint = "/login"
	suite.ctx = context.Background()
	suite.user = User{DB: db}
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

	suite.patch = gomonkey.NewPatches()
}

func (suite *LoginTestSuite) TearDownSuit() {
	suite.patch.Reset()

	sqlDB, _ := suite.user.DB.DB()
	_ = sqlDB.Close()
}

func (suite *LoginTestSuite) CallHandler(requestBody string) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodPost, suite.endpoint, strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.e.NewContext(req, rec)
	err := suite.user.Login(c)

	return rec, err
}

func (suite *LoginTestSuite) TestLogin_Login_Binding_Failure() {
	require := suite.Require()
	expectedError := "code=400, message=error in parse request data"

	requestBody := `{"user_name:"use"}`
	_, err := suite.CallHandler(requestBody)

	require.EqualError(err, expectedError)
}

func (suite *LoginTestSuite) TestLogin_Login_FindUserNameDBError_Failure() {
	require := suite.Require()
	expectedError := errors.New("database error")

	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`user_name` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs("username").
		WillReturnError(errors.New("database error"))

	requestBody := `{"user_name":"username","password":"Aaaaaaaa768!"}`
	_, err := suite.CallHandler(requestBody)

	require.EqualError(err, expectedError.Error())
}

func (suite *LoginTestSuite) TestLogin_Login_FindUserNameDB_RecordNotFoundErr_Failure() {
	require := suite.Require()
	expectedError := "code=400, message=invalid username or password"

	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`user_name` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs("username").
		WillReturnError(gorm.ErrRecordNotFound)

	requestBody := `{"user_name":"username","password":"Aaaaaaaa768!"}`
	_, err := suite.CallHandler(requestBody)

	require.EqualError(err, expectedError)
}

func (suite *LoginTestSuite) TestLogin_Login_VerifyPassword_Failure() {
	require := suite.Require()
	expectedError := "code=400, message=invalid username or password"

	rows := sqlmock.NewRows([]string{"id", "user_name", "password"}).
		AddRow(1, "username", "$2a$10$wBDhXmJfiZ9nskiXAijWre1PB8htQBEPhkxRgPHkK0dQUm65nBIu")
	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`user_name` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs("username").
		WillReturnRows(rows)

	suite.patch.ApplyFunc(utils.VerifyPassword, func(hashedPassword string, candidatePassword string) error {
		return errors.New("error")
	})

	requestBody := `{"user_name":"username","password":"Aaaaaaaa768!"}`
	_, err := suite.CallHandler(requestBody)

	require.EqualError(err, expectedError)
}

func (suite *LoginTestSuite) TestLogin_Login_GenerateToken_Failure() {
	require := suite.Require()
	expectedError := "code=500, message=Internal Server Error"

	rows := sqlmock.NewRows([]string{"id", "user_name", "password"}).
		AddRow(1, "username", "$2a$10$wBDhXmJfiZ9nskiXAijWre1PB8htQBEPhkxRgFPHkK0dQUm65nBIu")
	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`user_name` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs("username").
		WillReturnRows(rows)

	suite.patch.ApplyFunc(utils.VerifyPassword, func(hashedPassword string, candidatePassword string) error {
		return nil
	})

	suite.patch.ApplyFunc(utils.GenerateToken, func(id uint) (string, error) {
		return "", errors.New("error")
	})

	requestBody := `{"user_name":"username","password":"Aaaaaaaa768!"}`
	_, err := suite.CallHandler(requestBody)

	require.EqualError(err, expectedError)
}

func (suite *LoginTestSuite) TestLogin_Login_Success() {
	require := suite.Require()
	expectedMsg := `{"status":"success","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1Nzc4MzY4NjAsImlhdCI6MTU3NzgzNjgwMCwibmJmIjoxNTc3ODM2ODAwLCJzdWIiOjF9.NFfVAWrBMyWHMvpmnKR7hKLngPTtf7NObp9J9kn78G8"}`

	rows := sqlmock.NewRows([]string{"id", "user_name", "password"}).
		AddRow(1, "username", "$2a$10$wBDhXmJfiZ9nskiXAijWre1PB8htQBEPhkxRgFPHkK0dQUm65nBIu")
	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`user_name` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs("username").
		WillReturnRows(rows)

	suite.patch.ApplyFunc(utils.VerifyPassword, func(hashedPassword string, candidatePassword string) error {
		return nil
	})

	suite.patch.ApplyFunc(utils.GenerateToken, func(id uint) (string, error) {
		return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1Nzc4MzY4NjAsImlhdCI6MTU3NzgzNjgwMCwibmJmIjoxNTc3ODM2ODAwLCJzdWIiOjF9.NFfVAWrBMyWHMvpmnKR7hKLngPTtf7NObp9J9kn78G8", nil
	})

	requestBody := `{"user_name":"username","password":"Aaaaaaaa768!"}`
	response, err := suite.CallHandler(requestBody)

	require.NoError(err)
	require.Equal(response.Code, http.StatusOK)
	require.JSONEq(expectedMsg, response.Body.String())
}

func TestSignup(t *testing.T) {
	suite.Run(t, new(SignupTestSuite))
}

func TestLogin(t *testing.T) {
	suite.Run(t, new(LoginTestSuite))
}
