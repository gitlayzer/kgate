package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"golang.org/x/net/proxy"
)

var (
	discoverRange       string
	discoverPort        string = "22"
	discoverWorkers     int    = 100
	discoverDefaultUser string
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover nodes within the network through springboard machines",
	Long: `Scan the host with the SSH port open in the specified IP segment.
This operation protects the performance of the springboard by running a scanner on your local machine and routes its traffic through the SOCKS agent established with the springboard.`,
	Run: runDiscover,
}

func init() {
	discoverCmd.Flags().StringVarP(&discoverRange, "range", "r", "", "IP range of CIDR format to scan (for example, 192.168.1.0/24)")
	discoverCmd.Flags().StringVarP(&discoverPort, "port", "p", "22", "SSH port to scan")
	discoverCmd.Flags().IntVarP(&discoverWorkers, "workers", "w", 100, "Number of worker threads for concurrent scans")
	discoverCmd.Flags().StringVarP(&discoverDefaultUser, "default-user", "u", "", "Set a default username for all newly discovered hosts to skip interactive prompts")
	discoverCmd.MarkFlagRequired("range")
}

func ipListFromCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); func(ip net.IP) {
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}(ip) {
		ips = append(ips, ip.String())
	}
	// Remove network and broadcast addresses
	if len(ips) > 2 {
		return ips[1 : len(ips)-1], nil
	}
	return ips, nil
}

func runDiscover(cmd *cobra.Command, args []string) {
	clusterName, _ := cmd.Flags().GetString("cluster")
	if clusterName == "" {
		fmt.Fprintln(os.Stderr, "错误: discover 命令需要 --cluster 标志。")
		os.Exit(1)
	}

	// 1. 查找集群和跳板机信息
	cluster, err := cfg.FindCluster(clusterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
	bastion := &cluster.Bastion

	// 2. 启动后台 SSH SOCKS 代理
	fmt.Println("--> 正在通过跳板机启动 SSH SOCKS 代理...")
	proxyPort := "1080" // Standard SOCKS port
	proxyAddr := "127.0.0.1:" + proxyPort

	sshArgs := []string{"-N", "-D", proxyAddr}
	if bastion.IdentityFile != "" {
		sshArgs = append(sshArgs, "-i", bastion.IdentityFile)
	}
	sshArgs = append(sshArgs, fmt.Sprintf("%s@%s", bastion.User, bastion.Host))

	proxyCmd := exec.Command("ssh", sshArgs...)
	proxyCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Crucial for killing the process group

	if err := proxyCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "启动 SSH 代理时出错: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("--> SSH SOCKS 代理已在端口 %s 上成功启动 (PID: %d)\n", proxyPort, proxyCmd.Process.Pid)

	// 使用 defer 确保扫描结束后代理进程被终止
	defer func() {
		fmt.Println("\n--> 正在关闭 SSH SOCKS 代理...")
		if err := syscall.Kill(-proxyCmd.Process.Pid, syscall.SIGKILL); err != nil {
			fmt.Fprintf(os.Stderr, "关闭代理进程组时出错: %v\n", err)
		} else {
			fmt.Println("--> 代理已关闭。")
		}
	}()

	time.Sleep(2 * time.Second)

	// 3. 解析 IP 范围并准备扫描
	ipsToScan, err := ipListFromCIDR(discoverRange)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析 IP 范围时出错: %v\n", err)
		os.Exit(1)
	}

	// 4. 设置 SOCKS5 拨号器
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建 SOCKS5 拨号器时出错: %v\n", err)
		os.Exit(1)
	}
	contextDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		fmt.Fprintln(os.Stderr, "错误: 拨号器不支持 context。")
		os.Exit(1)
	}

	// 5. 并发扫描
	var wg sync.WaitGroup
	ipsChan := make(chan string, discoverWorkers)
	openHosts := make(chan string, len(ipsToScan))

	fmt.Printf("--> 正在扫描 %s 中的 %d 个主机，查找开放的端口 %s...\n", discoverRange, len(ipsToScan), discoverPort)

	for i := 0; i < discoverWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range ipsChan {
				target := fmt.Sprintf("%s:%s", ip, discoverPort)
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

				conn, err := contextDialer.DialContext(ctx, "tcp", target)
				cancel()

				if err == nil {
					conn.Close()
					openHosts <- ip
				}
			}
		}()
	}

	for _, ip := range ipsToScan {
		ipsChan <- ip
	}
	close(ipsChan)
	wg.Wait()
	close(openHosts)

	// 6. 处理扫描结果
	var newHosts []string
	existingIPs := make(map[string]bool)
	for _, node := range cluster.Nodes {
		existingIPs[node.IP] = true
	}

	for host := range openHosts {
		if !existingIPs[host] {
			newHosts = append(newHosts, host)
		}
	}

	if len(newHosts) == 0 {
		fmt.Println("\n--> 未发现新主机。")
		return
	}

	fmt.Printf("\n✅ 扫描完成。发现 %d 个新的潜在主机:\n", len(newHosts))
	for _, host := range newHosts {
		fmt.Println("  - " + host)
	}

	// 7. 交互式添加新节点
	prompt := promptui.Prompt{
		Label:     "您想将这些主机添加到配置中吗?",
		IsConfirm: true,
	}
	if _, err := prompt.Run(); err != nil {
		fmt.Println("操作已中止。")
		return
	}

	// --- 核心修改点 ---
	// 优化用户交互：一次性获取通用配置
	userForAll := discoverDefaultUser
	if userForAll == "" { // 如果用户没有通过 -u 标志提供默认用户
		userPrompt := promptui.Prompt{
			Label:   "为所有新主机输入一个通用的用户名 (可稍后逐个修改)",
			Default: "root",
		}
		userForAll, _ = userPrompt.Run()
	}

	for _, host := range newHosts {
		fmt.Printf("\n--- 正在添加主机 %s ---\n", host)
		aliasPrompt := promptui.Prompt{
			Label:   fmt.Sprintf("为此主机 '%s' 输入别名", host),
			Default: fmt.Sprintf("node-%s", host),
		}
		alias, _ := aliasPrompt.Run()

		// 直接使用之前获取的通用用户名
		cluster.AddNode(alias, host, userForAll)
		fmt.Printf("已添加节点: %s (%s)，用户名为 %s\n", alias, host, userForAll)
	}

	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "\n保存配置时出错: %v\n", err)
	} else {
		fmt.Println("\n✅ 配置已成功保存！")
	}
}
