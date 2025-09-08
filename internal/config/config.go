package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Clusters []Cluster `yaml:"clusters"`
}

type Cluster struct {
	Name    string  `yaml:"name"`
	Bastion Bastion `yaml:"bastion"`
	Nodes   []Node  `yaml:"nodes,omitempty"`
}

type Bastion struct {
	Host         string `yaml:"host"`
	User         string `yaml:"user"`
	Port         int    `yaml:"port,omitempty"`
	IdentityFile string `yaml:"identityFile,omitempty"`
}

type Node struct {
	Alias string `yaml:"alias"`
	IP    string `yaml:"ip"`
	User  string `yaml:"user"`
}

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", ".kgate", "config.yaml"), nil
}

func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// 如果目录或文件不存在，则创建
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return nil, err
		}
		return &Config{Clusters: []Cluster{}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}

func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c *Config) FindNode(alias string) (*Node, *Cluster, error) {
	for i, cluster := range c.Clusters {
		for j, node := range cluster.Nodes {
			if node.Alias == alias {
				return &c.Clusters[i].Nodes[j], &c.Clusters[i], nil
			}
		}
	}
	return nil, nil, fmt.Errorf("node with alias '%s' not found", alias)
}

func (c *Config) FindCluster(name string) (*Cluster, error) {
	for i, cluster := range c.Clusters {
		if cluster.Name == name {
			return &c.Clusters[i], nil
		}
	}
	return nil, fmt.Errorf("cluster with name '%s' not found", name)
}

func (c *Cluster) AddNode(alias, ip, user string) {
	newNode := Node{
		Alias: alias,
		IP:    ip,
		User:  user,
	}
	c.Nodes = append(c.Nodes, newNode)
}
