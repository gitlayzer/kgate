package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gitlayzer/kgate/internal/config"
	"github.com/spf13/cobra"
)

var recursive bool

var scpCmd = &cobra.Command{
	Use:   "scp [-r] [source] [destination]",
	Short: "Copy files between a local machine and a remote node",
	Long: `Securely copies files or directories using the bastion's own keys.
One path must be local, and the other remote.
A remote path is specified with the syntax: [node-alias]:/path/to/file`,
	Args: cobra.ExactArgs(2),
	Run:  runScp,
}

// parseScpArg 解析 scp 参数，判断其是否为远程路径
// 返回: 别名, 路径, 是否为远程路径
func parseScpArg(arg string) (alias, path string, isRemote bool) {
	parts := strings.SplitN(arg, ":", 2)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], true
	}
	return arg, "", false
}

func runScp(cmd *cobra.Command, args []string) {
	source := args[0]
	destination := args[1]

	srcAlias, srcPath, srcIsRemote := parseScpArg(source)
	destAlias, destPath, destIsRemote := parseScpArg(destination)

	if srcIsRemote == destIsRemote {
		fmt.Fprintln(os.Stderr, "Error: One path must be local and one must be remote (e.g., 'node-alias:/path').")
		os.Exit(1)
	}

	// 查找节点和跳板机信息
	var nodeAlias string
	if srcIsRemote {
		nodeAlias = srcAlias
	} else {
		nodeAlias = destAlias
	}
	node, cluster, err := cfg.FindNode(nodeAlias)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	fmt.Printf("--> Transferring files via bastion %s using bastion's key...\n", cluster.Name)

	if !srcIsRemote { // --- UPLOAD ---
		localPath := source
		remotePath := destPath
		upload(&cluster.Bastion, node, localPath, remotePath)
	} else { // --- DOWNLOAD ---
		remotePath := srcPath
		localPath := destination
		download(&cluster.Bastion, node, remotePath, localPath)
	}

	fmt.Println("✅ Transfer complete.")
}

// upload handles file uploads using 'tar' over a double SSH pipe.
func upload(bastion *config.Bastion, node *config.Node, localPath, remotePath string) {
	localDir := filepath.Dir(localPath)
	localFile := filepath.Base(localPath)

	remoteNodeCmd := fmt.Sprintf("ssh %s@%s \"mkdir -p %s && tar xf - -C %s\"", node.User, node.IP, remotePath, remotePath)

	bastionAddr := fmt.Sprintf("%s@%s", bastion.User, bastion.Host)
	sshArgs := []string{}
	if bastion.IdentityFile != "" {
		sshArgs = append(sshArgs, "-i", bastion.IdentityFile)
	}
	sshArgs = append(sshArgs, bastionAddr, remoteNodeCmd)
	bastionCmd := exec.Command("ssh", sshArgs...)

	// --- 核心修改点 ---
	// 智能判断是否需要添加 --no-xattr 标志来消除 macOS 上的警告
	tarArgs := []string{"cf", "-", "-C", localDir, localFile}
	if runtime.GOOS == "darwin" {
		// 如果在 macOS 上运行，添加此标志以避免 xattr 警告
		tarArgs = append([]string{"--no-xattr"}, tarArgs...)
	}
	localCmd := exec.Command("tar", tarArgs...)

	var err error
	bastionCmd.Stdin, err = localCmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating stdout pipe for local tar: %v\n", err)
		os.Exit(1)
	}
	bastionCmd.Stdout = os.Stdout
	bastionCmd.Stderr = os.Stderr

	if err := localCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting local tar command: %v\n", err)
		os.Exit(1)
	}
	if err := bastionCmd.Run(); err != nil {
		os.Exit(1)
	}
	if err := localCmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Error waiting for local tar command: %v\n", err)
		os.Exit(1)
	}
}

// download handles file downloads using 'tar' over a double SSH pipe.
func download(bastion *config.Bastion, node *config.Node, remotePath, localPath string) {
	remoteDir := filepath.Dir(remotePath)
	remoteFile := filepath.Base(remotePath)

	remoteNodeCmd := fmt.Sprintf("ssh %s@%s \"tar cf - -C %s %s\"", node.User, node.IP, remoteDir, remoteFile)

	bastionAddr := fmt.Sprintf("%s@%s", bastion.User, bastion.Host)
	sshArgs := []string{}
	if bastion.IdentityFile != "" {
		sshArgs = append(sshArgs, "-i", bastion.IdentityFile)
	}
	sshArgs = append(sshArgs, bastionAddr, remoteNodeCmd)
	bastionCmd := exec.Command("ssh", sshArgs...)

	// 智能判断本地解压目录
	destDir := "." // 默认解压到当前目录
	info, err := os.Stat(localPath)
	if err == nil && info.IsDir() {
		// 如果用户提供的路径是一个已存在的目录，则使用该目录
		destDir = localPath
	}

	// 本地命令：从 stdin 解压 tar 包到正确的目录
	localCmd := exec.Command("tar", "xf", "-", "-C", destDir)

	// --- 管道连接逻辑保持不变 ---
	localCmd.Stdin, err = bastionCmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating stdout pipe for bastion ssh: %v\n", err)
		os.Exit(1)
	}
	localCmd.Stdout = os.Stdout
	localCmd.Stderr = os.Stderr

	if err := bastionCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting bastion ssh command: %v\n", err)
		os.Exit(1)
	}
	if err := localCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during local tar execution: %v\n", err)
		os.Exit(1)
	}
	if err := bastionCmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Error waiting for bastion ssh command: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	scpCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively copy entire directories (default for tar)")
}
