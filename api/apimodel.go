package api

// API config
type Config struct {
	APIHost  string `mapstructure:"ApiHost"`
	NodeID   int	`mapstructure:"NodeID"`
	Key      string	`mapstructure:"ApiKey"`
	NodeType string	`mapstructure:"NodeType"`
}

// Node status
type NodeStatus struct {
	CPU    float64
	Mem    float64
	Disk   float64
	Uptime int
}

type NodeInfo struct {
	NodeType          string // Must be V2ray, Trojan, and Shadowsocks
	NodeID            int
	Port              int
	SpeedLimit        uint64
	AlterID           int
	TransportProtocol string
	Host              string
	Path              string
	EnableTLS         bool
	TLSType           string
	EnableVless       bool
}

type UserInfo struct {
	UID           int
	Email         string
	Passwd        string
	Port          int
	Method        string
	SpeedLimit    uint64
	Protocol      string
	ProtocolParam string
	Obfs          string
	ObfsParam     string
	UUID          string
}

type OnlineUser struct {
	UID int
	IP  string
}

type UserTraffic struct {
	UID      int
	Email    string
	Upload   int64
	Download int64
}

type ClientInfo struct {
	APIHost  string
	NodeID   int
	Key      string
	NodeType string
}
