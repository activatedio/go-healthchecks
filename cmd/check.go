/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/activatedio/go-healthchecks/checks"
	"github.com/activatedio/go-healthchecks/config"
	"github.com/activatedio/go-healthchecks/driver"
	"os"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Runs health check",
	Long:  `Runs health check`,
	Run: func(cmd *cobra.Command, args []string) {

		l := func(msg string) {
			fmt.Println(msg)
		}

		ctx := context.Background()

		c, err := config.NewConfig()
		CheckExit(err)

		d := driver.NewDriver(driver.DriverParams{
			Registry: checks.NewRegistry(),
		})

		s, err := d.Run(ctx, c)

		CheckExit(err)

		switch s {
		case checks.StatusHealthy:
			l("all checks are healthy")
			os.Exit(0)
		case checks.StatusUnhealthy:
			l("not all checks are healthy")
			os.Exit(1)
		default:
			CheckExit(errors.New("unrecognized status"))
		}

	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
