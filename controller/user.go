package controller

import (
	"errors"
	"github.com/labstack/echo/v4"
	"golang-example/model"
	"golang-example/utils"
	"gorm.io/gorm"
	"net/http"
	"regexp"
)

var userNamePattern *regexp.Regexp

func init() {
	userNamePattern = regexp.MustCompile("^[a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*$")
}

type User struct {
	DB *gorm.DB
}

type signupReq struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

func (req *signupReq) validate() error {
	if !(len(req.UserName) < 40 && len(req.UserName) > 7) || !userNamePattern.MatchString(req.UserName) {
		return errors.New("username is invalid")
	}

	if err := utils.ValidatePasswordPattern(req.Password); err != nil {
		return errors.New("password isn't strong enough")
	}

	return nil
}

type signupRes struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

func (u *User) Signup(ctx echo.Context) error {
	var req signupReq
	err := ctx.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "error in parse request data")
	}

	if err = req.validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var user model.User
	err = u.DB.Where(model.User{UserName: req.UserName}).First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if user.ID != 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "username is already taken")
	}

	hashedPass, err := utils.HashPassword(req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	user.UserName = req.UserName
	user.Password = hashedPass

	err = u.DB.Create(&user).Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	return ctx.JSON(http.StatusCreated, signupRes{Status: "success", Token: token})
}

type loginReq struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

func (u *User) Login(ctx echo.Context) error {
	var req loginReq
	err := ctx.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "error in parse request data")
	}

	var user model.User
	err = u.DB.Where(model.User{UserName: req.UserName}).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid username or password")
		}

		return err
	}

	if err = utils.VerifyPassword(user.Password, req.Password); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid username or password")
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	return ctx.JSON(http.StatusOK, signupRes{Status: "success", Token: token})
}
