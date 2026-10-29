package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var phpVersion string
var projectPath string

var composerCmd = &cobra.Command{
	Use:   "composer [commands...]",
	Short: "Run composer with a specific PHP version",
	Long:  `Run composer with a specific PHP version managed by PHForge. Example: phforge composer install --php=8.2`,
	Run: func(cmd *cobra.Command, args []string) {
		if phpVersion == "" {
			fmt.Println("Error: --php flag is required (e.g. --php=8.2)")
			os.Exit(1)
		}

		exePath, err := os.Executable()
		if err != nil {
			fmt.Printf("Failed to get executable path: %v\n", err)
			os.Exit(1)
		}
		baseDir := filepath.Join(filepath.Dir(exePath), "dist")
		phpExePath := filepath.Join(baseDir, "php-versions", "php-"+phpVersion, "php.exe")
		
		// Fallback for testing on developer machine
		if _, err := os.Stat(phpExePath); os.IsNotExist(err) {
			phpExePath = "php" // Assume it's in PATH for testing if not found in PHForge folder
		}

		composerPharPath := filepath.Join(baseDir, "composer", "composer.phar")
		
		// Fallback for testing on developer machine
		if _, err := os.Stat(composerPharPath); os.IsNotExist(err) {
			composerPharPath = "composer" // Assume composer is in PATH
		}

		// Prepare command arguments: php.exe composer.phar [args...]
		cmdArgs := []string{}
		if composerPharPath != "composer" {
			cmdArgs = append(cmdArgs, composerPharPath)
		}
		cmdArgs = append(cmdArgs, args...)

		var execCmd *exec.Cmd
		if composerPharPath == "composer" && phpExePath == "php" {
			// Just run composer directly if both are in PATH (fallback mode)
			execCmd = exec.Command("composer", args...)
		} else {
			execCmd = exec.Command(phpExePath, cmdArgs...)
		}

		// Set working directory if provided
		if projectPath != "" {
			execCmd.Dir = projectPath
		}

		// Forward Stdin, Stdout, Stderr so it feels native
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		err = execCmd.Run()
		if err != nil {
			fmt.Printf("Composer command failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(composerCmd)

	composerCmd.Flags().StringVar(&phpVersion, "php", "", "PHP version to use (e.g. 8.2)")
	composerCmd.Flags().StringVar(&projectPath, "path", "", "Project directory path (optional)")
}
