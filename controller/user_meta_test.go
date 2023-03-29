package controller

import (
	"context"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"golang-example/config"
	"golang-example/database"
	"golang-example/model"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type UpdateTestSuite struct {
	suite.Suite
	e        *echo.Echo
	endpoint string
	ctx      context.Context
	sqlMock  sqlmock.Sqlmock
	patch    *gomonkey.Patches
	userMeta UserMeta
	userID   uint
}

func (suite *UpdateTestSuite) SetupSuite() {
	sqlMock, db := database.NewMySQLDBGormMock()
	suite.sqlMock = sqlMock

	mockCtrl := gomock.NewController(suite.T())
	defer mockCtrl.Finish()

	suite.e = echo.New()
	suite.endpoint = "/metas"
	suite.ctx = context.Background()
	suite.userMeta = UserMeta{DB: db}
	suite.userID = 1
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
	suite.patch.ApplyFunc(time.Now, func() time.Time {
		return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	})
}

func (suite *UpdateTestSuite) TearDownSuit() {
	suite.patch.Reset()

	sqlDB, _ := suite.userMeta.DB.DB()
	_ = sqlDB.Close()
}

func (suite *UpdateTestSuite) CallHandler(query string) (*httptest.ResponseRecorder, error) {
	if query != "" {
		suite.endpoint = suite.endpoint + query
	}

	req := httptest.NewRequest(http.MethodPut, suite.endpoint, strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.e.NewContext(req, rec)
	c.Set("user_id", suite.userID)
	err := suite.userMeta.Update(c)

	return rec, err
}

func (suite *UpdateTestSuite) TestUpdate_Update_UserIDNotFound_Failure() {
	require := suite.Require()
	expectedError := "code=404, message=user not found"

	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnError(gorm.ErrRecordNotFound)

	query := `?gender=male`
	_, err := suite.CallHandler(query)

	require.EqualError(err, expectedError)
}

func (suite *UpdateTestSuite) TestUpdate_Update_UserIDDBErr_Failure() {
	require := suite.Require()
	expectedError := "code=500, message=Internal Server Error"

	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnError(errors.New("database err"))

	query := `?gender=male`
	_, err := suite.CallHandler(query)

	require.EqualError(err, expectedError)
}

func (suite *UpdateTestSuite) TestUpdate_Update_InvalidGender_Failure() {
	require := suite.Require()
	expectedError := "code=400, message=invalid gender"

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1)
	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnRows(rows)

	query := `?gender=mal`
	_, err := suite.CallHandler(query)

	require.EqualError(err, expectedError)
}

func (suite *UpdateTestSuite) TestUpdate_Update_EmptyQuery() {
	require := suite.Require()

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1)
	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnRows(rows)

	response, err := suite.CallHandler("")

	require.NoError(err)
	require.Equal(http.StatusNoContent, response.Code)
}

func (suite *UpdateTestSuite) TestUpdate_Update_UpdateMetaDBErr_Failure() {
	require := suite.Require()
	expectedError := "code=500, message=Internal Server Error"

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1)
	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnRows(rows)

	syntax = "^UPDATE `user_meta` SET `meta_value`=.+,`updated_at`=.+ WHERE user_id = .+ AND meta_key = .+"
	suite.sqlMock.ExpectBegin()
	suite.sqlMock.ExpectExec(syntax).
		WithArgs("male", time.Now(), suite.userID, model.UMKGender).
		WillReturnError(errors.New("database err"))
	suite.sqlMock.ExpectRollback()

	query := `?gender=male`
	_, err := suite.CallHandler(query)

	require.EqualError(err, expectedError)
}

func (suite *UpdateTestSuite) TestUpdate_Update_CreateMetaDBErr_Failure() {
	require := suite.Require()
	expectedError := "code=500, message=Internal Server Error"

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1)
	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnRows(rows)

	syntax = "^UPDATE `user_meta` SET `meta_value`=.+,`updated_at`=.+ WHERE user_id = .+ AND meta_key = .+"
	suite.sqlMock.ExpectBegin()
	suite.sqlMock.ExpectExec(syntax).
		WithArgs("male", time.Now(), suite.userID, model.UMKGender).
		WillReturnResult(sqlmock.NewResult(0, 0))
	suite.sqlMock.ExpectCommit()

	suite.sqlMock.ExpectBegin()
	syntax = "^INSERT INTO `user_meta`"
	suite.sqlMock.ExpectExec(syntax).
		WillReturnError(errors.New("database err"))
	suite.sqlMock.ExpectRollback()

	query := `?gender=male`
	_, err := suite.CallHandler(query)

	require.EqualError(err, expectedError)
}

func (suite *UpdateTestSuite) TestUpdate_Update_UpdateMeta_Success() {
	require := suite.Require()

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1)
	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnRows(rows)

	syntax = "^UPDATE `user_meta` SET `meta_value`=.+,`updated_at`=.+ WHERE user_id = .+ AND meta_key = .+"
	suite.sqlMock.ExpectBegin()
	suite.sqlMock.ExpectExec(syntax).
		WithArgs("male", time.Now(), suite.userID, model.UMKGender).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.sqlMock.ExpectCommit()

	query := `?gender=male`
	response, err := suite.CallHandler(query)

	require.NoError(err)
	require.Equal(http.StatusNoContent, response.Code)
}

func (suite *UpdateTestSuite) TestUpdate_Update_UpdateMeta_TwoKeysSent_Success() {
	require := suite.Require()

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1)
	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnRows(rows)

	syntax = "^UPDATE `user_meta` SET `meta_value`=.+,`updated_at`=.+ WHERE user_id = .+ AND meta_key = .+"
	suite.sqlMock.ExpectBegin()
	suite.sqlMock.ExpectExec(syntax).
		WithArgs("22", time.Now(), suite.userID, model.UMKAge).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.sqlMock.ExpectCommit()

	syntax = "^UPDATE `user_meta` SET `meta_value`=.+,`updated_at`=.+ WHERE user_id = .+ AND meta_key = .+"
	suite.sqlMock.ExpectBegin()
	suite.sqlMock.ExpectExec(syntax).
		WithArgs("male", time.Now(), suite.userID, model.UMKGender).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.sqlMock.ExpectCommit()

	query := `?gender=male&&age=22`
	response, err := suite.CallHandler(query)

	require.NoError(err)
	require.Equal(http.StatusNoContent, response.Code)
}

func (suite *UpdateTestSuite) TestUpdate_Update_CreateMeta_Success() {
	require := suite.Require()

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1)
	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnRows(rows)

	syntax = "^UPDATE `user_meta` SET `meta_value`=.+,`updated_at`=.+ WHERE user_id = .+ AND meta_key = .+"
	suite.sqlMock.ExpectBegin()
	suite.sqlMock.ExpectExec(syntax).
		WithArgs("male", time.Now(), suite.userID, model.UMKGender).
		WillReturnResult(sqlmock.NewResult(0, 0))
	suite.sqlMock.ExpectCommit()

	suite.sqlMock.ExpectBegin()
	syntax = "^INSERT INTO `user_meta`"
	suite.sqlMock.ExpectExec(syntax).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.sqlMock.ExpectCommit()

	query := `?gender=male`
	response, err := suite.CallHandler(query)

	require.NoError(err)
	require.Equal(http.StatusNoContent, response.Code)
}

type GetTestSuite struct {
	suite.Suite
	e        *echo.Echo
	endpoint string
	ctx      context.Context
	sqlMock  sqlmock.Sqlmock
	patch    *gomonkey.Patches
	userMeta UserMeta
	userID   uint
}

func (suite *GetTestSuite) SetupSuite() {
	sqlMock, db := database.NewMySQLDBGormMock()
	suite.sqlMock = sqlMock

	mockCtrl := gomock.NewController(suite.T())
	defer mockCtrl.Finish()

	suite.e = echo.New()
	suite.endpoint = "/metas"
	suite.ctx = context.Background()
	suite.userMeta = UserMeta{DB: db}
	suite.userID = 1
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
	suite.patch.ApplyFunc(time.Now, func() time.Time {
		return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	})
}

func (suite *GetTestSuite) TearDownSuit() {
	suite.patch.Reset()

	sqlDB, _ := suite.userMeta.DB.DB()
	_ = sqlDB.Close()
}

func (suite *GetTestSuite) CallHandler(key string) (*httptest.ResponseRecorder, error) {
	if key != "" {
		suite.endpoint = fmt.Sprintf("%v?key=%v", suite.endpoint, key)
	}

	req := httptest.NewRequest(http.MethodGet, suite.endpoint, strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.e.NewContext(req, rec)
	c.Set("user_id", suite.userID)
	err := suite.userMeta.Get(c)

	return rec, err
}

func (suite *GetTestSuite) TestGet_Get_InvalidKey_Failure() {
	require := suite.Require()
	expectedError := "code=400, message=invalid key"

	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := suite.CallHandler("locale")

	require.EqualError(err, expectedError)
}

func (suite *GetTestSuite) TestGet_Get_UserIDNotFound_Failure() {
	require := suite.Require()
	expectedError := "code=404, message=user not found"

	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := suite.CallHandler("")

	require.EqualError(err, expectedError)
}

func (suite *GetTestSuite) TestGet_Get_UserIDDBErr_Failure() {
	require := suite.Require()
	expectedError := "code=500, message=Internal Server Error"

	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnError(errors.New("database err"))

	_, err := suite.CallHandler("")

	require.EqualError(err, expectedError)
}

func (suite *GetTestSuite) TestGet_Get_FindMetaDBErr_Failure() {
	require := suite.Require()
	expectedError := "code=500, message=Internal Server Error"

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1)
	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnRows(rows)

	syntax = "^SELECT (.+) FROM `user_meta` WHERE `user_meta`.`user_id` = (.+)"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnError(errors.New("database err"))

	_, err := suite.CallHandler("")

	require.EqualError(err, expectedError)
}

func (suite *GetTestSuite) TestGet_Get_WithoutKey_Success() {
	require := suite.Require()
	expectedMsg := "[{\"key\":\"gender\",\"value\":\"male\"},{\"key\":\"age\",\"value\":\"23\"}]\n"

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1)
	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"meta_key", "meta_value", "user_id"}).
		AddRow(model.UMKGender, "male", 1).AddRow(model.UMKAge, 23, 1)
	syntax = "^SELECT (.+) FROM `user_meta` WHERE user_id = (.+)"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnRows(rows)

	response, err := suite.CallHandler("")

	require.NoError(err)
	require.Equal(expectedMsg, response.Body.String())
	require.Equal(http.StatusOK, response.Code)
}

func (suite *GetTestSuite) TestGet_Get_WithKey_Success() {
	require := suite.Require()
	expectedMsg := "[{\"key\":\"gender\",\"value\":\"male\"}]\n"

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1)
	syntax := "^SELECT (.+) FROM `users` WHERE `users`.`id` = (.+) ORDER BY `users`.`id` LIMIT 1"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID).
		WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"meta_key", "meta_value", "user_id"}).
		AddRow(model.UMKGender, "male", 1)
	syntax = "^SELECT (.+) FROM `user_meta` WHERE user_id = (.+) AND meta_key = (.+)"
	suite.sqlMock.ExpectQuery(syntax).
		WithArgs(suite.userID, model.UMKGender).
		WillReturnRows(rows)

	response, err := suite.CallHandler("gender")

	require.NoError(err)
	require.Equal(expectedMsg, response.Body.String())
	require.Equal(http.StatusOK, response.Code)
}

func TestUpdate(t *testing.T) {
	suite.Run(t, new(UpdateTestSuite))
}

func TestGet(t *testing.T) {
	suite.Run(t, new(GetTestSuite))
}
