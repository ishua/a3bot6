package config

type Config struct {
    Xray   XrayConfig   `yaml:"xray"`
    Auth   AuthConfig   `yaml:"auth"`
    Server ServerConfig `yaml:"server"`
}

type XrayConfig struct {
    ConfigDir    string `yaml:"config_dir"`
    XrayGRPCAddr string `yaml:"xray_grpc_addr"`
    HTTPProxy    string `yaml:"http_proxy"`
    BalancerTag  string `yaml:"balancer_tag"`
}

type AuthConfig struct {
    Secrets []string `yaml:"secrets"`
}

type ServerConfig struct {
    Addr string `yaml:"addr"`
}