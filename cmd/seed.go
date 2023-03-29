package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang-example/config"
	"golang-example/database"
	"golang-example/model"
	"golang.org/x/crypto/bcrypt"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var seedDatabaseCMD = &cobra.Command{
	Use:   "seed",
	Short: "seed database with data",
	Run: func(cmd *cobra.Command, args []string) {
		seedDB()
	},
}

func seedDB() {
	if !(host == "" && port == "" && db == "" && user == "") {
		fmt.Print("Password: ")
		terminalOutput, err := term.ReadPassword(0)
		if err != nil {
			log.Fatalf("there is problem on reading password from terminal: %s", err)
		}

		config.C.Database.Password = string(terminalOutput)
	}

	db := database.InitDatabase()

	log.Info("Truncating `users`")
	if err := db.Exec("TRUNCATE TABLE users;").Error; err != nil {
		log.Fatalf("error in truncating `users`: `%s`", err)
	}

	log.Info("Truncating `user_meta``")
	if err := db.Exec("TRUNCATE TABLE user_meta;").Error; err != nil {
		log.Fatalf("error in truncating `user_meta`: `%s`", err)
	}

	log.Info("Creating users")
	users := createMockUsers(3)
	err := db.CreateInBatches(users, 6).Error
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Creating user_meta")
	userMetas := createMockUserMeta(3)
	err = db.CreateInBatches(userMetas, 6).Error
	if err != nil {
		log.Fatal(err)
	}
}

func createMockUsers(n int) []*model.User {
	users := make([]*model.User, 0, n)

	for i := 1; i < n+1; i++ {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte(fmt.Sprintf("password%03d", i)), bcrypt.DefaultCost)
		u := &model.User{
			UserName:  fmt.Sprintf("user%03d", i),
			Password:  string(hashedPass),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		users = append(users, u)
	}

	return users
}

func createMockUserMeta(n int) []*model.UserMeta {
	userMetas := make([]*model.UserMeta, 0, n)

	for i := 1; i < n+1; i++ {
		um := &model.UserMeta{
			MetaKey:   model.UMKAge,
			MetaValue: "22",
			UserID:    uint(i),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		userMetas = append(userMetas, um)
	}

	return userMetas
}
