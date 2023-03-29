package cmd

import (
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"golang-example/config"
	"golang-example/controller"
	"golang-example/database"
	"golang-example/middleware"
)

var serveCMD = &cobra.Command{
	Use:   "serve",
	Short: "serve API",
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func serve() {
	db := database.InitDatabase()

	redis := database.InitRedis()
	defer database.CloseRedis(redis)

	e := echo.New()

	userController := controller.User{DB: db}
	userMetaController := controller.UserMeta{DB: db}

	e.POST("/signup", userController.Signup)
	e.POST("/login", userController.Login)

	e.PUT("/metas", userMetaController.Update, middleware.UserAuthorized(), middleware.Lock(redis))
	e.GET("/metas", userMetaController.Get, middleware.UserAuthorized())

	// Start server
	e.Logger.Fatal(e.Start(config.C.Address))
}
