package cmd

import (
	"fmt"
	"os"

	"github.com/gitlayzer/kgate/internal/config"
	"github.com/spf13/cobra"
)

var (
	version   string
	buildDate string
)

// cfg 将在所有子命令中共享
var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "kgate",
	Short: "kgate is a smart SSH gateway tool for multi-bastion environments",
	Long: `A flexible and easy-to-use CLI tool to manage SSH connections
through multiple bastion hosts without complex ssh_config files.`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// cobra.OnInitialize 会在任何命令执行前调用 initConfig 函数
	cobra.OnInitialize(initConfig)
	// 在这里添加所有子命令
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(nodesCmd)
	rootCmd.AddCommand(scpCmd)
	rootCmd.SetVersionTemplate(fmt.Sprintf("{{.Use}} version %s (built on %s)\n", version, buildDate))
}

func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading config:", err)
		os.Exit(1)
	}
}
