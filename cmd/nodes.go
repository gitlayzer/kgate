package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/gitlayzer/kgate/internal/config"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Manage nodes within a cluster",
	Long:  `Provides commands to list, add, or remove nodes for a specific cluster.`,
}

var nodesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all nodes for a specific cluster",
	Run: func(cmd *cobra.Command, args []string) {
		clusterName, _ := cmd.Flags().GetString("cluster")
		if clusterName == "" {
			fmt.Fprintln(os.Stderr, "Error: --cluster flag is required")
			return
		}
		var targetCluster *config.Cluster
		for i := range cfg.Clusters {
			if cfg.Clusters[i].Name == clusterName {
				targetCluster = &cfg.Clusters[i]
				break
			}
		}

		if targetCluster == nil {
			fmt.Fprintf(os.Stderr, "Error: Cluster '%s' not found.\n", clusterName)
			return
		}

		if len(targetCluster.Nodes) == 0 {
			fmt.Printf("No nodes configured for cluster '%s'.\n", clusterName)
			return
		}
		fmt.Printf("Nodes in cluster '%s':\n", clusterName)
		for _, node := range targetCluster.Nodes {
			fmt.Printf("- Alias: %s (Target: %s@%s)\n", node.Alias, node.User, node.IP)
		}
	},
}

var nodesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new node to a specific cluster",
	Run: func(cmd *cobra.Command, args []string) {
		clusterName, _ := cmd.Flags().GetString("cluster")
		if clusterName == "" {
			fmt.Fprintln(os.Stderr, "Error: --cluster flag is required")
			return
		}

		var targetCluster *config.Cluster
		var clusterIndex int = -1

		for i := range cfg.Clusters {
			if cfg.Clusters[i].Name == clusterName {
				targetCluster = &cfg.Clusters[i]
				clusterIndex = i
				break
			}
		}

		if targetCluster == nil {
			fmt.Fprintf(os.Stderr, "Error: Cluster '%s' not found.\n", clusterName)
			return
		}

		prompt := promptui.Prompt{Label: "Node Alias"}
		alias, err := prompt.Run()
		if err != nil {
			return
		}
		prompt = promptui.Prompt{Label: "Node Internal IP"}
		ip, err := prompt.Run()
		if err != nil {
			return
		}
		prompt = promptui.Prompt{Label: "Node User"}
		user, err := prompt.Run()
		if err != nil {
			return
		}

		newNode := config.Node{Alias: alias, IP: ip, User: user}
		cfg.Clusters[clusterIndex].Nodes = append(cfg.Clusters[clusterIndex].Nodes, newNode)
		if err := cfg.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
			return
		}
		fmt.Printf("✅ Node '%s' added to cluster '%s' successfully.\n", alias, clusterName)
	},
}

// nodes remove
var nodesRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a node from a specific cluster",
	Run: func(cmd *cobra.Command, args []string) {
		clusterName, _ := cmd.Flags().GetString("cluster")
		if clusterName == "" {
			fmt.Fprintln(os.Stderr, "Error: --cluster flag is required")
			return
		}

		var targetCluster *config.Cluster
		var clusterIndex int = -1
		for i := range cfg.Clusters {
			if cfg.Clusters[i].Name == clusterName {
				targetCluster = &cfg.Clusters[i]
				clusterIndex = i
				break
			}
		}
		if targetCluster == nil {
			fmt.Fprintf(os.Stderr, "Error: Cluster '%s' not found.\n", clusterName)
			return
		}
		if len(targetCluster.Nodes) == 0 {
			fmt.Printf("No nodes to remove in cluster '%s'.\n", clusterName)
			return
		}

		var nodeAliases []string
		for _, n := range targetCluster.Nodes {
			nodeAliases = append(nodeAliases, n.Alias)
		}

		prompt := promptui.Select{
			Label: "Select node to remove",
			Items: nodeAliases,
		}
		_, result, err := prompt.Run()
		if err != nil {
			if errors.Is(err, promptui.ErrInterrupt) {
				os.Exit(-1)
			}
			return
		}

		var newNodes []config.Node
		for _, n := range targetCluster.Nodes {
			if n.Alias != result {
				newNodes = append(newNodes, n)
			}
		}
		cfg.Clusters[clusterIndex].Nodes = newNodes
		if err := cfg.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
			return
		}
		fmt.Printf("✅ Node '%s' removed from cluster '%s' successfully.\n", result, clusterName)
	},
}

func init() {
	// 为所有 nodes 子命令添加必需的 --cluster 标志
	nodesListCmd.Flags().String("cluster", "", "The name of the cluster")
	nodesAddCmd.Flags().String("cluster", "", "The name of the cluster")
	nodesRemoveCmd.Flags().String("cluster", "", "The name of the cluster")
	nodesListCmd.MarkFlagRequired("cluster")
	nodesAddCmd.MarkFlagRequired("cluster")
	nodesRemoveCmd.MarkFlagRequired("cluster")

	nodesCmd.AddCommand(nodesListCmd)
	nodesCmd.AddCommand(nodesAddCmd)
	nodesCmd.AddCommand(discoverCmd)
	nodesCmd.AddCommand(nodesRemoveCmd)
}
