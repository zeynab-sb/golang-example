package cmd

import (
	"github.com/spf13/cobra"
)

var (
	host string
	port string
	db   string
	user string
)

var databaseCMD = &cobra.Command{
	Use:   "database",
	Short: "Database related commands",
}

func init() {
	databaseCMD.AddCommand(migrateDatabaseCMD)
	databaseCMD.AddCommand(seedDatabaseCMD)
}
