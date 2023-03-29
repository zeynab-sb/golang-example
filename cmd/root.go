package cmd

import (
	"github.com/spf13/cobra"
	"golang-example/config"
	"log"
)

var rootCMD = &cobra.Command{
	Use: "Server",
}

var configPath string

func init() {
	cobra.OnInitialize(func() {
		config.Init(configPath)
	})

	rootCMD.PersistentFlags().StringVar(&configPath, "config", "", "config path (directory or file)")
	rootCMD.PersistentFlags().StringVar(&host, "db-host", "", "database server host")
	rootCMD.PersistentFlags().StringVar(&port, "db-port", "", "database server port")
	rootCMD.PersistentFlags().StringVar(&db, "db-name", "", "database name")
	rootCMD.PersistentFlags().StringVar(&user, "db-user", "", "database user")

	rootCMD.AddCommand(serveCMD)
	rootCMD.AddCommand(databaseCMD)
}

func Execute() {
	if err := rootCMD.Execute(); err != nil {
		log.Fatal(err)
	}
}
