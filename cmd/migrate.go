package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang-example/config"
	"golang-example/database"
	"golang.org/x/term"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	libmigrate "github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/spf13/cobra"
)

var (
	migrationsPath  string
	migrationsTable string
)

var migrateDatabaseCMD = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		migrateDB()
	},
}

func init() {
	migrateDatabaseCMD.Flags().StringVarP(&migrationsPath, "migrations-path", "m", "", "path to migrations directory")
	migrateDatabaseCMD.Flags().StringVarP(&migrationsTable, "migrations-table", "t", "schema_migrations", "database table holding migrations")
}

func migrateDB() {
	if migrationsPath == "" {
		log.Fatal("migrations path is required")
	}

	if !(strings.HasPrefix(migrationsPath, "/")) {
		path, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		migrationsPath, err = filepath.Abs(filepath.Join(path, migrationsPath))
		if err != nil {
			log.Fatal("cannot resolve full migration path")
		}
	}
	log.Infof("migrations path: %s", migrationsPath)

	if !(host == "" && port == "" && db == "" && user == "") {
		fmt.Print("Password: ")
		terminalOutput, err := term.ReadPassword(0)
		if err != nil {
			log.Fatalf("there is problem on reading password from terminal: %s", err)
		}
		config.C.Database.Password = string(terminalOutput)
	}

	appDB, err := database.InitDatabase().DB()
	if err != nil {
		log.Fatal(err)
	}

	driver, err := mysql.WithInstance(appDB, &mysql.Config{})
	if err != nil {
		log.Fatal(err)
	}
	m, err := libmigrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		config.C.Database.DB,
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	err = m.Up()

	if err != nil {
		if err.Error() == "no change" {
			log.Info(err)
		} else {
			log.Fatal(err)
		}
	}
}
