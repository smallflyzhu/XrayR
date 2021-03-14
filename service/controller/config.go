package controller

type Config struct {
	UpdatePeriodic int `mapstructure:"UpdatePeriodic"`
	CertConfig     *CertConfig `mapstructure:"CertConfig"`
}

type CertConfig struct {
	CertMode   string            `mapstructure:"CertMode"` // none, file, http, dns
	CertDomain string            `mapstructure:"CertDomain"`
	CertFile   string            `mapstructure:"CertFile"`
	KeyFile    string            `mapstructure:"KeyFile"`
	Provider   string            `mapstructure:"Provider"` // alidns, cloudflare, gandi, godaddy....
	Email      string            `mapstructure:"Email"`
	DNSEnv     map[string]string `mapstructure:"DNSEnv"`
}
