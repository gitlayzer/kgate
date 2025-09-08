package cmd

import (
	"fmt"
	"os"

	"github.com/gitlayzer/kgate/internal/config"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage cluster configurations (bastions)",
	Long:  `Provides commands to list, add, or remove cluster configurations.`,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured clusters",
	Run: func(cmd *cobra.Command, args []string) {
		if len(cfg.Clusters) == 0 {
			fmt.Println("No clusters configured. Use 'kgate config add' to create one.")
			return
		}
		fmt.Println("Configured Clusters:")
		for _, cluster := range cfg.Clusters {
			fmt.Printf("- %s (Bastion: %s@%s)\n", cluster.Name, cluster.Bastion.User, cluster.Bastion.Host)
		}
	},
}

var configAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new cluster configuration",
	Run: func(cmd *cobra.Command, args []string) {
		prompt := promptui.Prompt{Label: "Cluster Name"}
		name, err := prompt.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prompt failed %v\n", err)
			return
		}

		prompt = promptui.Prompt{Label: "Bastion Host (IP or FQDN)"}
		host, err := prompt.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prompt failed %v\n", err)
			return
		}

		prompt = promptui.Prompt{Label: "Bastion User"}
		user, err := prompt.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prompt failed %v\n", err)
			return
		}

		prompt = promptui.Prompt{Label: "Path to Bastion Identity File (e.g., ~/.ssh/id_rsa)"}
		identityFile, err := prompt.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prompt failed %v\n", err)
			return
		}

		newCluster := config.Cluster{
			Name: name,
			Bastion: config.Bastion{
				Host:         host,
				User:         user,
				IdentityFile: identityFile,
			},
		}

		cfg.Clusters = append(cfg.Clusters, newCluster)
		if err := cfg.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
			return
		}
		fmt.Printf("✅ Cluster '%s' added successfully.\n", name)
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a cluster configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if len(cfg.Clusters) == 0 {
			fmt.Println("No clusters to remove.")
			return
		}

		var clusterNames []string
		for _, c := range cfg.Clusters {
			clusterNames = append(clusterNames, c.Name)
		}

		prompt := promptui.Select{
			Label: "Select cluster to remove",
			Items: clusterNames,
		}
		_, result, err := prompt.Run()
		if err != nil {
			// User pressed Ctrl+C, it's not an error
			if err == promptui.ErrInterrupt {
				os.Exit(-1)
			}
			fmt.Fprintf(os.Stderr, "Prompt failed %v\n", err)
			return
		}

		var newClusters []config.Cluster
		for _, c := range cfg.Clusters {
			if c.Name != result {
				newClusters = append(newClusters, c)
			}
		}
		cfg.Clusters = newClusters
		if err := cfg.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
			return
		}
		fmt.Printf("✅ Cluster '%s' removed successfully.\n", result)
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configRemoveCmd)
}
