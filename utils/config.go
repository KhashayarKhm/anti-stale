package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type Config struct {
	UserAgent string `json:"userAgent"`
	Token     string `json:"token"`
	Owners    Owners `json:"owners"`
}

type Owners map[string]Repositories

type Repositories map[string]Entities

type Entities struct {
	Issues []int `json:"issues"`
	PRs    []int `json:"prs"`
}

func (c *Config) MustLoad(logger *Log, cfgPath string) {
	if cfgPath == "" {
		cwd, err := os.Getwd()
		MustOk(err)
		cfgPath = path.Join(cwd, "anti-stale.json")
		logger.Warning(fmt.Sprintf("config path is empty. use %s", cfgPath))
	}

	data, err := os.ReadFile(cfgPath)
	MustOk(err)

	err = json.Unmarshal(data, &c)
	MustOk(err)
}
