package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec [node-alias] [command...]",
	Short: "Execute a non-interactive command on a remote node",
	Long: `Executes a command on a specified node via its bastion.
This command is non-interactive. It's useful for running scripts or getting quick outputs.`,
	Args: cobra.MinimumNArgs(2),
	Run:  runExec,
}

func runExec(cmd *cobra.Command, args []string) {
	nodeAlias := args[0]
	// 将用户输入的所有命令部分连接成一个字符串
	commandToRun := strings.Join(args[1:], " ")

	node, cluster, err := cfg.FindNode(nodeAlias)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	bastion := cluster.Bastion
	bastionAddr := fmt.Sprintf("%s@%s", bastion.User, bastion.Host)
	nodeAddr := fmt.Sprintf("%s@%s", node.User, node.IP)

	fmt.Printf("--> Executing on %s via bastion %s: [%s]\n", node.Alias, cluster.Name, commandToRun)

	// 构建将在跳板机上执行的远程命令
	// 注意：这里没有 -t 参数，因为我们不需要交互式终端
	// 我们将用户的命令用双引号包裹，以确保它被作为一个整体在目标节点上执行
	remoteCommand := fmt.Sprintf("ssh %s \"%s\"", nodeAddr, commandToRun)

	// 构建最终的 ssh 命令参数
	var sshArgs []string

	if bastion.IdentityFile != "" {
		sshArgs = append(sshArgs, "-i", bastion.IdentityFile)
	}

	sshArgs = append(sshArgs, bastionAddr, remoteCommand)

	// 创建 exec.Command
	sshCmd := exec.Command("ssh", sshArgs...)

	// 对于 exec，我们只需要获取其输出，所以连接 stdout 和 stderr
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	// 注意：我们不连接 stdin，因为这是非交互式的

	// 执行命令
	if err := sshCmd.Run(); err != nil {
		// SSH 客户端会自己打印详细的错误信息
		os.Exit(1)
	}
}
