# kgate - 现代化的 SSH 跳板机管理工具

***kgate*** 是一款使用 ***Go*** 语言开发的现代化、高效率的命令行工具，旨在解决在多跳板机（Bastion Host）环境中管理和访问内网服务器的复杂性。

本工具专注于 ***跳板机密钥认证*** 模式：用户仅需拥有跳板机的访问权限，即可通过 ***kgate*** 无缝、安全地访问所有已授权的后端节点，而无需在每个节点上配置个人密钥。

## ✨ 功能特性
- 交互式连接 (connect): 快速登录到任意后端节点，并获得一个功能完整的 bash 登录会话。
- 远程命令执行 (exec): 在后端节点上执行非交互式命令，轻松实现自动化脚本。
- 安全文件传输 (scp): 在本地和后端节点之间安全地上传和下载文件/目录，无需暴露本地密钥。
- 交互式配置管理 (config, nodes): 通过友好的命令行提示，安全、轻松地管理集群和节点，无需手动编辑 YAML 文件。
- 跨平台构建 (Makefile): 一键构建适用于 Linux (amd64) 和 macOS (arm64) 的发行版本。

## 🛠️ 功能状态总览

| 功能模块 | 子命令        | 状态 | 备注                          |
|------|------------|--|-----------------------------|
| 核心连接 | connect    | ✅ 已完成 | 提供功能完整的交互式 Bash Login Shell |
| 命令执行 | exec       | ✅ 已完成 | 在远程节点上执行非交互式命令              |
| 文件传输 | scp        | ✅ 已完成 | 在本地与远程节点间安全传输文件             |
| 配置管理 | config     | ✅ 已完成 | 用于管理集群/跳板机配置                |
| 节点管理 | nodes      | ✅ 已完成 | 用于手动管理集群下的节点信息              |
| 节点扫描 | discover   | ✅ 已完成 | 自动化扫描节点信息并添加到配置             |
| 构建系统 | Makefile   | ✅ 已完成 | 支持 Linux/macOS 交叉编译，注入版本信息  |
| 版本控制 | .gitignore | ✅ 已完成 | 已配置标准的 Go 项目忽略规则            |

## 🚀 安装与构建
### 前提条件
- Go (版本 1.24+)
- make
- git
### 构建流程
项目使用 Makefile 进行标准化构建。

#### 1: 构建所有目标平台 (Linux amd64, macOS arm64):
make all

#### 2: 清理构建产物:
make clean

#### 构建后的二进制文件将位于项目根目录的 bin/ 文件夹下。

## ⚙️ 配置文件
***kgate*** 的所有行为都由一个位于 ***~/.config/.kgate/config.yaml*** 的 YAML 文件驱动。您可以使用 ***kgate config*** 和 ***kgate nodes*** 命令来管理此文件。

### 配置示例
```shell
clusters:
  - name: jump-server
    bastion:
      host: 1.1.1.1
      user: ubuntu
      identityFile: ~/.ssh/id_rsa
    nodes:
    - alias: dev-01
      ip: 2.2.2.2
      user: ubuntu
```

## 📚 使用指南 (命令参考)
***kgate connect [node-alias]*** \
与指定的后端节点建立一个功能完整的、交互式的 SSH 会话。

***kgate exec [node-alias] [command...]*** \
在指定的后端节点上执行一条或多条非交互式命令。

***kgate scp [-r] [source] [destination]*** \
在本地和指定的后端节点之间安全地传输文件或目录。
- 远程路径格式: [node-alias]:/path/to/file
- 目录复制: 本工具默认支持目录传输。为保持与传统 scp 兼容，也提供了 -r 标志，但并非必需。

#### 示例:
```shell
# 上传文件
./bin/kgate scp ./local.log dev-01:/tmp/

# 下载文件
./bin/kgate scp dev-01:/var/log/app.log ./

# 上传整个目录
./bin/kgate scp ./my-app dev-01:/opt/
```
***kgate config 和 kgate nodes*** \
提供 list, add, remove 子命令，用于通过命令行交互式地管理集群和节点配置。

## 🔮 未来计划 (Planned Features)
### kgate nodes discover - 节点自动发现
- **状态: ✅ 已完成**
- **功能: 自动扫描跳板机所在的内网环境，发现新的可访问节点，并引导用户将其添加到配置文件中。**

## 📄 许可证 (License)
本项目采用 MIT 许可证。详情请参阅 [LICENSE](./LICENSE) 文件。