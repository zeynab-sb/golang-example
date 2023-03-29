package controller

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"golang-example/model"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

const userIDContextField = "user_id"

type UserMeta struct {
	DB *gorm.DB
}

func (um *UserMeta) Update(ctx echo.Context) error {
	id := ctx.Get(userIDContextField).(uint)
	err := um.DB.Where(model.User{ID: id}).First(&model.User{}).Error
	if err == gorm.ErrRecordNotFound {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	params := ctx.QueryParams()
	var userMetas []model.UserMeta
	if v, ok := params["age"]; ok {
		age, err := strconv.Atoi(v[0])
		if err != nil || age <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, errors.New("invalid age"))
		}

		userMetas = append(userMetas, model.UserMeta{
			MetaKey:   model.UMKAge,
			MetaValue: fmt.Sprint(age),
			UserID:    id,
		})
	}

	if v, ok := params["gender"]; ok {
		gender := v[0]
		if _, ok := model.GendersMap[gender]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, errors.New("invalid gender"))
		}

		userMetas = append(userMetas, model.UserMeta{
			MetaKey:   model.UMKGender,
			MetaValue: gender,
			UserID:    id,
		})
	}

	if len(userMetas) != 0 {
		for _, userMeta := range userMetas {
			result := um.DB.Model(&model.UserMeta{}).
				Where("user_id = ?", userMeta.UserID).
				Where("meta_key = ?", userMeta.MetaKey).Update("meta_value", userMeta.MetaValue)

			if result.Error != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
			}

			if result.RowsAffected == 0 {
				err = um.DB.Create(&userMeta).Error
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
				}
			}
		}
	}

	return ctx.NoContent(http.StatusNoContent)
}

type getReq struct {
	Key string `query:"key"`
}

func (req *getReq) validate() error {
	if req.Key != "" {
		if _, ok := model.KeysMap[model.UserMetaKey(req.Key)]; !ok {
			return errors.New("invalid key")
		}
	}

	return nil
}

type getRes struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (um *UserMeta) Get(ctx echo.Context) error {
	var req getReq
	err := ctx.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "error in parse request data")
	}

	if err = req.validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	id := ctx.Get(userIDContextField).(uint)
	err = um.DB.Where(model.User{ID: id}).First(&model.User{}).Error
	if err == gorm.ErrRecordNotFound {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	query := um.DB.Where("user_id = ?", id)
	if req.Key != "" {
		query = query.Where("meta_key = ?", req.Key)
	}

	var userMetas []model.UserMeta
	if err = query.Find(&userMetas).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	var response []getRes
	for _, userMeta := range userMetas {
		response = append(response, getRes{
			Key:   string(userMeta.MetaKey),
			Value: userMeta.MetaValue,
		})
	}

	return ctx.JSON(http.StatusOK, response)
}
