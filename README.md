# XRayR
[![](https://img.shields.io/badge/TgChat-@XrayR讨论-blue.svg)](https://t.me/XrayR_project)
[![](https://img.shields.io/badge/Channel-@XrayR通知-blue.svg)](https://t.me/XrayR_channel)

A Xray backend framework that can easily support many panels.

一个基于Xray的后端框架，支持V2ay,Trojan,Shadowsocks协议，极易扩展，支持多面板对接。

如果您喜欢本项目，可以右上角点个star+watch，持续关注本项目的进展。

## 免责声明

本项目只是本人个人学习开发并维护，本人不保证任何可用性，也不对使用本软件造成的任何后果负责。

## 特点
* 永久开源且免费。
* 支持V2ray，Trojan， Shadowsocks多种协议。
* 支持Vless和XTLS等新特性。
* 支持单实例对接多面板、多节点，无需重复启动。
* 支持限制在线IP
* 支持节点端口级别、用户级别限速。
* 配置简单明了。
* 修改配置自动重启实例。
* 方便编译和升级，可以快速更新核心版本， 支持Xray-core新特性。



## TODO
- [x] 在线用户统计和限制
- [x] 限速实现
- [ ] 审计规则
- [ ] 对接ProxyPanel
- [ ] 对接v2board

## 功能介绍

| 功能            | v2ray | trojan | shadowsocks |
| --------------- | ----- | ------ | ----------- |
| 获取节点信息    | √     | √      | √           |
| 获取用户信息    | √     | √      | √           |
| 用户流量统计    | √     | √      | √           |
| 服务器信息上报  | √     | √      | √           |
| 自动申请tls证书 | √     | √      | √           |
| 自动续签tls证书 | √     | √      | √           |
| 在线人数统计    | √     | √      | √           |
| 在线用户限制    | √     | √      | √           |
| 审计规则        | TODO  | TODO   | TODO        |
| 节点端口限速    | √     | √      | √           |
| 按照用户限速    | √     | √      | √           |

## 支持前端

| 前端        | v2ray | trojan | shadowsocks                    |
| ----------- | ----- | ------ | ------------------------------ |
| sspanel-uim | √     | √      | √ (Shadowsocks - V2Ray-Plugin) |
| ProxyPanel  | TODO  | TODO   | TODO                           |
| v2board     | TODO  | TODO   | TODO                           |

## Thanks

* [Project X](https://github.com/XTLS/)
* [V2Fly](https://github.com/v2fly)
* [VNet-V2ray](https://github.com/ProxyPanel/VNet-V2ray)
* [Air-Universe](https://github.com/crossfw/Air-Universe)

## Licence

[Mozilla Public License Version 2.0](https://github.com/RManLuo/XrayR/blob/main/LICENSE)

## Telgram

[XrayR后端讨论](https://t.me/XrayR_project)

[XrayR通知](https://t.me/XrayR_channel)

## 下载并使用
1. 在此处，根据自身系统选择合适的版本：[Release](https://github.com/RManLuo/XrayR/releases)
2. 解压压缩包，之后运行：`./XrayR -config config.yml`

## 编译并使用
1. go 1.16.0
2. 依次运行
```bash
git clone https://github.com/RManLuo/XrayR
cd XrayR/main
go mod tidy
go build -o XrayR -ldflags "-s -w"
./XrayR -config config.yml
```
## 配置文件
1. `cp config.yml.example config.yml`
2. nano config.yml

配置文件基本格式，Nodes下可以同时添加多个面板，多个节点配置信息，只需添加相同格式的Nodes item即可。
```
Log:
  Level: debug # Log level: none, error, warning, info, debug 
  AccessPath: # ./access.Log
  ErrorPath: # ./error.log
Nodes:
  -
    PanelType: "SSpanel" # Panel type: SSpanel
    ApiConfig:
      ApiHost: "http://127.0.0.1:667"
      ApiKey: "123"
      NodeID: 41
      NodeType: V2ray # Node type: V2ray, Shadowsocks, Trojan
    ControllerConfig:
      UpdatePeriodic: 60 # Time to update the nodeinfo, how many sec.
      CertConfig:
        CertMode: dns # Option about how to get certificate: none, file, http, dns
        CertDomain: "node1.test.com" # Domain to cert
        CertFile: ./cert/node1.test.com.cert # Provided if the CertMode is file
        KeyFile: ./cert/node1.test.com.key
        Provider: alidns # DNS cert provider, Get the full support list here: https://go-acme.github.io/lego/dns/
        Email: test@me.com
        DNSEnv: # DNS ENV option used by DNS provider
          ALICLOUD_ACCESS_KEY: aaa
          ALICLOUD_SECRET_KEY: bbb
```
## 前端配置
### 限速说明
1. 节点限速：请在SSpanel的节点限速处填写，单位mbps。
2. 用户限速：请在SSpanel的用户设置处填写，单位mbps。
3. 限速值设为0，则为不限速。
### V2ray

| 协议      | 支持情况                                             |
| --------- | ---------------------------------------------------- |
| VMess     | tcp, tcp+tls/xtls, ws, ws+tls/xtls, h2c, h2+tls/xtls |
| VMessAEAD | tcp, tcp+tls/xtls, ws, ws+tls/xtls, h2c, h2+tls/xtls |
| VLess     | tcp, tcp+tls/xtls, ws, ws+tls/xtls, h2c, h2+tls/xtls |
#### SSpanel-uim 节点地址格式
```
IP;监听端口;alterId;(tcp或ws);(tls或xtls或不填);path=/xxx|host=xxxx.com|server=xxx.com|outside_port=xxx|enable_vless=(true或false)
```
alterId设为0，则自动启用VMessAEAD。
#### TCP示例
```
ip;12345;2;tcp;;server=域名
```
```
示例：1.3.5.7;12345;2;tcp;;server=hk.domain.com
```
#### tcp + tls 示例
```
ip;12345;2;tcp;tls;server=域名|host=域名
```
```
示例：1.3.5.7;12345;2;tcp;tls;server=hk.domain.com|host=hk.domain.com
```
#### ws示例
```
ip;80;2;ws;;path=/xxx|server=域名|host=CDN域名
```
```
示例：1.3.5.7;80;2;ws;;path=/v2ray|server=hk.domain.com|host=hk.domain.com
```
#### ws + tls 示例
```
ip;443;2;ws;tls;path=/xxx|server=域名|host=CDN域名
```
```
示例：1.3.5.7;443;2;ws;tls;path=/v2ray|server=hk.domain.com|host=hk.domain.com
```
#### 中转端口
在任一配置组合后增加`|outside_port=xxx`,此项为用户连接端口。
```
示例：1.3.5.7;80;2;ws;;path=/v2ray|server=hk.domain.com|host=hk.domain.com|outside_port=12345
```
#### 启用Vless **(此项为实验性功能)**

在任一配置组合后增加`|enable_vless=true`.
```
示例：1.3.5.7;12345;2;tcp;tls;server=hk.domain.com|host=hk.domain.com|enable_vless=true
```
请开启vless同时务必使用tls或者xtls。
#### 启用xtls **(此项为实验性功能)**
替换tls为xtls
```
示例：1.3.5.7;12345;2;tcp;xtls;server=hk.domain.com|host=hk.domain.com
```
### Trojan
| 协议   | 支持情况 |
| ------ | -------- |
| Trojan | √        |
#### SSpanel-uim 节点地址格式
```
域名或IP;port=监听端口#用户连接端口|host=xx
```
#### 直连示例
```
示例：gz.aaa.com;port=443|host=gz.aaa.com
```
#### 中转示例
```
示例：gz.aaa.com;port=443#12345|host=hk.aaa.com
```
#### 启用xtls **(此项为实验性功能)**
在配置后增加`|enable_xtls=true`.
```
示例：gz.aaa.com;port=443#12345|host=hk.aaa.com|enable_xtls=true
```

### Shadowsocks
| 协议            | 支持情况 | 加密方法                                    |
| --------------- | -------- | ------------------------------------------- |
| ShadowsocksAEAD | √        | aes-128-gcm, aes-256-gcm, chacha20-poly1305 |
#### SSpanel-uim 节点地址格式
请注意，节点类型请选择：`Shadowsocks - V2Ray-Plugin`
```
域名或IP;监听端口|server=xx
```
#### Shadowsocks 示例
```
示例：gz.aaa.com;12345|server=gz.aaa.com
```


