package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var connectCmd = &cobra.Command{
	Use:   "connect [node-alias]",
	Short: "Connect to a node via its bastion using an interactive SSH session",
	Args:  cobra.ExactArgs(1),
	Run:   runConnect,
}

func runConnect(cmd *cobra.Command, args []string) {
	nodeAlias := args[0]

	node, cluster, err := cfg.FindNode(nodeAlias)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	bastion := cluster.Bastion
	bastionAddr := fmt.Sprintf("%s@%s", bastion.User, bastion.Host)
	nodeAddr := fmt.Sprintf("%s@%s", node.User, node.IP)

	fmt.Printf("--> Connecting to %s (%s) via bastion %s (%s) using bastion's key\n", node.Alias, node.IP, cluster.Name, bastion.Host)

	innerSshArgs := []string{}
	if term.IsTerminal(int(os.Stdin.Fd())) {
		innerSshArgs = append(innerSshArgs, "-t")
	}
	innerSshArgs = append(innerSshArgs, nodeAddr, "/bin/bash", "-l")
	remoteCommand := "ssh " + strings.Join(innerSshArgs, " ")

	var outerSshArgs []string

	outerSshArgs = append(outerSshArgs, "-t")

	if bastion.IdentityFile != "" {
		outerSshArgs = append(outerSshArgs, "-i", bastion.IdentityFile)
	}
	outerSshArgs = append(outerSshArgs, bastionAddr, remoteCommand)

	sshCmd := exec.Command("ssh", outerSshArgs...)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	if err := sshCmd.Run(); err != nil {
		os.Exit(1)
	}
}
