package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "phforge",
	Short: "PHForge CLI Tool",
	Long:  `PHForge CLI tool allows you to run Composer with specific PHP versions managed by PHForge Server.`,
}

func execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	execute()
}
