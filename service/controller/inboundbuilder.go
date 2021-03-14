//Package generate the InbounderConfig used by add inbound
package controller

import (
	"encoding/json"
	"fmt"

	"github.com/RManLuo/XrayR/api"
	"github.com/RManLuo/XrayR/common/legocmd"

	"github.com/xtls/xray-core/common/uuid"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"
)

//InboundBuilder build Inbound config for different protocol
func InboundBuilder(nodeInfo *api.NodeInfo, certConfig *CertConfig) (*core.InboundHandlerConfig, error) {
	inboundDetourConfig := &conf.InboundDetourConfig{}

	// Build Port
	portRange := &conf.PortRange{From: uint32(nodeInfo.Port), To: uint32(nodeInfo.Port)}
	inboundDetourConfig.PortRange = portRange
	// Build Tag
	inboundDetourConfig.Tag = fmt.Sprintf("%s_%d", nodeInfo.NodeType, nodeInfo.Port)
	// SniffingConfig
	sniffingConfig := &conf.SniffingConfig{Enabled: true,
		DestOverride: &conf.StringList{"http", "tls"},
	}
	inboundDetourConfig.SniffingConfig = sniffingConfig

	var (
		protocol      string
		streamSetting *conf.StreamConfig
		setting       json.RawMessage
	)

	var proxySetting interface{}
	// Build Protocol and Protocol setting
	if nodeInfo.NodeType == "V2ray" {
		if nodeInfo.EnableVless {
			protocol = "vless"
			proxySetting = &conf.VLessInboundConfig{
				Decryption: "none",
			}
		} else {
			protocol = "vmess"
			proxySetting = &conf.VMessInboundConfig{}
		}
	} else if nodeInfo.NodeType == "Trojan" {
		protocol = "trojan"
		proxySetting = &conf.TrojanServerConfig{}
	} else if nodeInfo.NodeType == "Shadowsocks" {
		protocol = "shadowsocks"
		proxySetting = &conf.ShadowsocksServerConfig{}
		randomPasswd := uuid.New()
		defaultSSuser := &conf.ShadowsocksUserConfig{
			Cipher:   "aes-128-gcm",
			Password: randomPasswd.String(),
		}
		proxySetting, _ := proxySetting.(*conf.ShadowsocksServerConfig)
		proxySetting.Users = append(proxySetting.Users, defaultSSuser)
	} else {
		return nil, fmt.Errorf("Unsupported node type: %s, Only support: V2ray, Trojan, and Shadowsocks", nodeInfo.NodeType)
	}

	setting, err := json.Marshal(proxySetting)
	if err != nil {
		return nil, fmt.Errorf("Marshal proxy %s config fialed: %s", nodeInfo.NodeType, err)
	}

	// Build streamSettings
	streamSetting = new(conf.StreamConfig)
	transportProtocol := conf.TransportProtocol(nodeInfo.TransportProtocol)
	networkType, err := transportProtocol.Build()
	if err != nil {
		return nil, fmt.Errorf("convert TransportProtocol failed: %s", err)
	}
	if networkType == "websocket" {
		headers := make(map[string]string)
		headers["Host"] = nodeInfo.Host
		wsSettings := &conf.WebSocketConfig{
			Path:    nodeInfo.Path,
			Headers: headers,
		}
		streamSetting.WSSettings = wsSettings
	}

	streamSetting.Network = &transportProtocol
	streamSetting.Security = nodeInfo.TLSType
	// Build TLS and XTLS settings
	if nodeInfo.EnableTLS {
		certFile, keyFile, err := getCertFile(certConfig)
		if err != nil {
			return nil, err
		}
		if nodeInfo.TLSType == "tls" {
			tlsSettings := &conf.TLSConfig{}
			tlsSettings.Certs = append(tlsSettings.Certs, &conf.TLSCertConfig{CertFile: certFile, KeyFile: keyFile, OcspStapling: 3600})

			streamSetting.TLSSettings = tlsSettings
		} else if nodeInfo.TLSType == "xtls" {
			xtlsSettings := &conf.XTLSConfig{}
			xtlsSettings.Certs = append(xtlsSettings.Certs, &conf.XTLSCertConfig{CertFile: certFile, KeyFile: keyFile, OcspStapling: 3600})
			streamSetting.XTLSSettings = xtlsSettings
		}
	}

	inboundDetourConfig.Protocol = protocol
	inboundDetourConfig.StreamSetting = streamSetting
	inboundDetourConfig.Settings = &setting

	return inboundDetourConfig.Build()
}

func getCertFile(certConfig *CertConfig) (certFile string, keyFile string, err error) {
	if certConfig.CertMode == "file" {
		if certConfig.CertFile == "" || certConfig.KeyFile == "" {
			return "", "", fmt.Errorf("Cert file path or key file path not exist")
		}
		return certConfig.CertFile, certConfig.KeyFile, nil
	} else if certConfig.CertMode == "dns" {
		lego, err := legocmd.New()
		if err != nil {
			return "", "", err
		}
		certPath, keyPath, err := lego.DNSCert(certConfig.CertDomain, certConfig.Email, certConfig.Provider, certConfig.DNSEnv)
		if err != nil {
			return "", "", err
		}
		return certPath, keyPath, err
	} else if certConfig.CertMode == "http" {
		lego, err := legocmd.New()
		if err != nil {
			return "", "", err
		}
		certPath, keyPath, err := lego.HTTPCert(certConfig.CertDomain, certConfig.Email)
		if err != nil {
			return "", "", err
		}
		return certPath, keyPath, err
	}

	return "", "", fmt.Errorf("Unsupported certmode: %s", certConfig.CertMode)
}
