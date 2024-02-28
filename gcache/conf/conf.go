package conf

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// 全局配置
type GlobalConfig struct {
	ApiAddr      string   `json:"api_addr,omitempty"`
	Prefix       string   `json:"prefix,omitempty"`
	DNS          string   `json:"dns,omitempty"`
	Replicas     int      `json:"replicas,omitempty"`
	HttpBasePath string   `json:"http_base_path,omitempty"`
	MaxBytes     int      `json:"max_bytes,omitempty"`
	Endpoints    []string `json:"endpoints,omitempty"`
	DialTimeout  int      `json:"dial_timeout,omitempty"`
	LeaseTTL     int      `json:"lease_ttl,omitempty"`
}

// 全局配置变量
var GConfig *GlobalConfig

func Init(path string) {
	GConfig = &GlobalConfig{}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("ReadFile", "err", err)
	}
	err = json.Unmarshal(data, GConfig)
	if err != nil {
		log.Fatal("Unmarshal", "err", err)
	}
	fmt.Println(GConfig)
}
