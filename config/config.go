package config

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	yaml "gopkg.in/yaml.v3"

	"github.com/forta-network/forta-node/utils"

	log "github.com/sirupsen/logrus"
)

const EnvConfig = "FORTA_CONFIG"
const EnvJsonRpcHost = "JSON_RPC_HOST"
const EnvJsonRpcPort = "JSON_RPC_PORT"
const EnvAgentGrpcPort = "AGENT_GRPC_PORT"
const ContainerNamePrefix = "forta"

const (
	//TODO: figure out protocol for this other than direct communication
	EnvQueryNode = "QUERY_NODE"
	// NATS node name env var
	EnvNatsHost = "NATS_HOST"
)

// Docker container names
var (
	DockerNatsContainerName         = fmt.Sprintf("%s-nats", ContainerNamePrefix)
	DockerScannerContainerName      = fmt.Sprintf("%s-scanner", ContainerNamePrefix)
	DockerJSONRPCProxyContainerName = fmt.Sprintf("%s-json-rpc", ContainerNamePrefix)
	DockerQueryContainerName        = fmt.Sprintf("%s-query", ContainerNamePrefix)

	DockerNetworkName = DockerScannerContainerName
)

// Global constant values
var (
	DefaultNatsPort    = "4222"
	DefaultIPFSGateway = "https://cloudflare-ipfs.com"
)

type AgentConfig struct {
	ID         string  `yaml:"id" json:"id"`
	Image      string  `yaml:"image" json:"image"`
	StartBlock *uint64 `yaml:"startBlock" json:"startBlock,omitempty"`
	StopBlock  *uint64 `yaml:"stopBlock" json:"stopBlock,omitempty"`
}

func (ac AgentConfig) ImageHash() string {
	_, digest := utils.SplitImageRef(ac.Image)
	return digest
}

func (ac AgentConfig) ContainerName() string {
	_, digest := utils.SplitImageRef(ac.Image)
	return fmt.Sprintf("%s-agent-%s-%s", ContainerNamePrefix, utils.ShortenString(ac.ID, 8), utils.ShortenString(digest, 4))
}

func (ac AgentConfig) GrpcPort() string {
	return "50051"
}

type DBConfig struct {
	Path string `yaml:"path" json:"path" validate:"dir"`
}

type EthereumConfig struct {
	JsonRpcUrl   string            `yaml:"jsonRpcUrl" json:"jsonRpcUrl" validate:"omitempty,url"`
	WebsocketUrl string            `yaml:"websocketUrl" json:"websocketUrl" validate:"required_without=jsonRpcUrl,url"`
	Headers      map[string]string `yaml:"headers" json:"headers"`
}

type QueryConfig struct {
	QueryImage string          `yaml:"queryImage" json:"queryImage"`
	Port       int             `yaml:"port" json:"port" validate:"number,max=65535"`
	DB         DBConfig        `yaml:"db" json:"db"`
	PublishTo  PublisherConfig `yaml:"publishTo" json:"publishTo"`
}
type ScannerConfig struct {
	ChainID      int            `yaml:"chainId" json:"chainId" validate:"number"`
	ScannerImage string         `yaml:"scannerImage" json:"scannerImage"`
	StartBlock   int            `yaml:"startBlock" json:"startBlock" validate:"omitempty,number"`
	EndBlock     int            `yaml:"endBlock" json:"endBlock" validate:"omitempty,number"`
	Ethereum     EthereumConfig `yaml:"ethereum" json:"ethereum"`
}

type TraceConfig struct {
	Ethereum EthereumConfig `yaml:"ethereum" json:"ethereum"`
	Enabled  bool           `yaml:"enabled" json:"enabled" validate:"omitempty,boolean"`
}

type JsonRpcProxyConfig struct {
	JsonRpcImage string         `yaml:"jsonRpcImage" json:"jsonRpcImage"`
	Ethereum     EthereumConfig `yaml:"ethereum" json:"ethereum"`
}

type LogConfig struct {
	Level       string `yaml:"level" json:"level"`
	MaxLogSize  string `yaml:"maxLogSize" json:"maxLogSize"`
	MaxLogFiles int    `yaml:"maxLogFiles" json:"maxLogFiles" validate:"omitempty,number"`
}

type RegistryConfig struct {
	Ethereum          EthereumConfig `yaml:"ethereum" json:"ethereum"`
	IPFSGateway       *string        `yaml:"ipfsGateway" json:"ipfs,omitempty" validate:"omitempty,url"`
	ContractAddress   string         `yaml:"contractAddress" json:"contractAddress" validate:"eth_addr"`
	ContainerRegistry string         `yaml:"containerRegistry" json:"containerRegistry" validate:"hostname"`
	Username          string         `yaml:"username" json:"username"`
	Password          string         `yaml:"password" json:"password"`
}

type PublisherConfig struct {
	SkipPublish     bool           `yaml:"skipPublish" json:"skipPublish"`
	Ethereum        EthereumConfig `yaml:"ethereum" json:"ethereum"  validate:"dive,required_unless=skipPublish"`
	ContractAddress string         `yaml:"contractAddress" json:"contractAddress" validate:"eth_addr"`
	IPFS            struct {
		GatewayURL string `yaml:"gatewayUrl" json:"gatewayUrl" validate:"required_unless=skipPublish,url"`
		Username   string `yaml:"username" json:"username"`
		Password   string `yaml:"password" json:"password"`
	} `yaml:"ipfs" json:"ipfs"`
	Batch BatchConfig `yaml:"batch" json:"batch"`
}

type BatchConfig struct {
	SkipEmpty       bool `yaml:"skipEmpty" json:"skipEmpty" validate:"omitempty,boolean"`
	IntervalSeconds *int `yaml:"intervalSeconds" json:"intervalSeconds" validate:"omitempty,number"`
	MaxAlerts       *int `yaml:"maxAlerts" json:"maxAlerts" validate:"omitempty,number"`
}

type Config struct {
	Production bool   `yaml:"-" json:"-"`
	FortaDir   string `yaml:"-" json:"-"`
	ConfigPath string `yaml:"-" json:"-"`
	KeyDirPath string `yaml:"-" json:"-"`
	Passphrase string `yaml:"-" json:"-"`

	Registry     RegistryConfig     `yaml:"registry" json:"registry"`
	Scanner      ScannerConfig      `yaml:"scanner" json:"scanner"`
	Query        QueryConfig        `yaml:"query" json:"query"`
	Trace        TraceConfig        `yaml:"trace" json:"trace"`
	JsonRpcProxy JsonRpcProxyConfig `yaml:"json-rpc-proxy" json:"jsonRpcProxy"`
	Log          LogConfig          `yaml:"log" json:"log"`
}

func GetCfgDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.forta", home), nil
}

func GetKeyStorePath() (string, error) {
	cfgDir, err := GetCfgDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.keys", cfgDir), nil
}

func ParseBigInt(num int) *big.Int {
	var val *big.Int
	if num != 0 {
		val = big.NewInt(int64(num))
	}
	return val
}

func InitLogLevel(cfg Config) error {
	if cfg.Log.Level != "" {
		lvl, err := log.ParseLevel(cfg.Log.Level)
		if err != nil {
			return err
		}
		log.SetLevel(lvl)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	return nil
}

func GetConfigFromEnv() (Config, error) {
	cfgJson := os.Getenv(EnvConfig)
	if cfgJson == "" {
		return Config{}, fmt.Errorf("%s is required", EnvConfig)
	}
	var cfg Config
	if err := json.Unmarshal([]byte(cfgJson), &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func GetConfig(filename string) (Config, error) {
	var cfg Config
	if err := readFile(filename, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func readFile(filename string, cfg *Config) error {
	f, err := os.Open(filename)
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		return err
	}

	decoder := yaml.NewDecoder(f)
	return decoder.Decode(cfg)
}

// EnvDefaults contain default values for one env.
type EnvDefaults struct {
	DefaultRegistryContractAddress string
	DefaultAlertContractAddress    string
	DiscoSubdomain                 string
}

// GetEnvDefaults returns the default values for an env.
func GetEnvDefaults(production bool) EnvDefaults {
	if production {
		return EnvDefaults{
			DefaultRegistryContractAddress: "0x38C1e080BeEb26eeA91932178E62987598230271",
			DefaultAlertContractAddress:    "0xf4746faBc1D5E751248Ea3AC87ceB13C432F0C1A",
			DiscoSubdomain:                 "disco-dev",
		}
	}
	return EnvDefaults{
		DefaultRegistryContractAddress: "0x1Ad235EF22Dd15d291ecD4b44a5739aD4F61b3A5",
		DefaultAlertContractAddress:    "0x8c06716460e4A6E8Ca6a0bfe7190b1a6A059eA2F",
		DiscoSubdomain:                 "disco",
	}
}
