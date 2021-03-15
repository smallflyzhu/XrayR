package controller

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/RManLuo/XrayR/api"
	"github.com/RManLuo/XrayR/common/legocmd"
	"github.com/RManLuo/XrayR/common/serverstatus"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/task"
	"github.com/xtls/xray-core/core"
)

type Controller struct {
	server                  *core.Instance
	config                  *Config
	clientInfo              api.ClientInfo
	apiClient               api.API
	nodeInfo                *api.NodeInfo
	userList                *[]api.UserInfo
	nodeStatus              *api.NodeStatus
	onlineUsers             *[]api.OnlineUser
	userTraffic             *[]api.UserTraffic
	nodeInfoMonitorPeriodic *task.Periodic
	userReportPeriodic      *task.Periodic
}

// New return a Controller service with default parameters.
func New(server *core.Instance, api api.API, config *Config) *Controller {
	controller := &Controller{
		server:    server,
		config:    config,
		apiClient: api,
	}
	return controller
}

// Start implement the Start() function of the service interface
func (c *Controller) Start() error {
	c.clientInfo = c.apiClient.Describe()
	// First fetch Node Info
	newNodeInfo, err := c.apiClient.GetNodeInfo()
	if err != nil {
		return err
	}
	// Add new tag
	err = c.addNewTag(newNodeInfo)
	if err != nil {
		log.Panic(err)
		return err
	}
	// Update user
	userInfo, err := c.apiClient.GetUserList()
	if err != nil {
		return err
	}
	err = c.addNewUser(userInfo, newNodeInfo)
	if err != nil {
		return err
	}
	c.nodeInfo = newNodeInfo
	c.userList = userInfo
	tag := fmt.Sprintf("%s_%d", c.nodeInfo.NodeType, c.nodeInfo.Port)
	// Add Limiter
	if err := c.AddInboundLimiter(tag, newNodeInfo.SpeedLimit, userInfo); err != nil {
		log.Print(err)
	}
	c.nodeInfoMonitorPeriodic = &task.Periodic{
		Interval: time.Duration(c.config.UpdatePeriodic) * time.Second,
		Execute:  c.nodeInfoMonitor,
	}
	c.userReportPeriodic = &task.Periodic{
		Interval: time.Duration(c.config.UpdatePeriodic) * time.Second,
		Execute:  c.userInfoMonitor,
	}
	log.Print("Start monitor node status")
	c.nodeInfoMonitorPeriodic.Start()
	log.Print("Start report node status")
	c.userReportPeriodic.Start()
	return nil
}

// Close implement the Close() function of the service interface
func (c *Controller) Close() error {
	if c.nodeInfoMonitorPeriodic != nil {
		err := c.nodeInfoMonitorPeriodic.Close()
		if err != nil {
			log.Panicf("node info periodic close failed: %s", err)
		}
	}

	if c.nodeInfoMonitorPeriodic != nil {
		err := c.userReportPeriodic.Close()
		if err != nil {
			log.Panicf("user report periodic close failed: %s", err)
		}
	}
	return nil
}

func (c *Controller) nodeInfoMonitor() (err error) {
	// First fetch Node Info
	newNodeInfo, err := c.apiClient.GetNodeInfo()
	if err != nil {
		return err
	}
	var nodeInfoChanged bool = false
	// If nodeInfo changed
	if !reflect.DeepEqual(c.nodeInfo, newNodeInfo) {
		// Remove old tag
		oldtag := fmt.Sprintf("%s_%d", c.nodeInfo.NodeType, c.nodeInfo.Port)
		err := c.removeOldTag(oldtag)
		if err != nil {
			log.Print(err)
		}
		// Add new tag
		err = c.addNewTag(newNodeInfo)
		if err != nil {
			log.Panic(err)
		}
		nodeInfoChanged = true
		c.nodeInfo = newNodeInfo
		// Remove Old limiter
		if err = c.DeleteInboundLimiter(oldtag); err != nil {
			log.Print(err)
		}
	}
	// Check Cert
	if c.nodeInfo.EnableTLS && (c.config.CertConfig.CertMode == "dns" || c.config.CertConfig.CertMode == "http") {
		lego, err := legocmd.New()
		if err != nil {
			log.Print(err)
		}
		// Xray-core supports the OcspStapling certification hot renew
		_, _, err = lego.RenewCert(c.config.CertConfig.CertDomain, c.config.CertConfig.Email, c.config.CertConfig.CertMode, c.config.CertConfig.Provider, c.config.CertConfig.DNSEnv)
		if err != nil {
			log.Print(err)
		}
	}
	// Update User
	newUserInfo, err := c.apiClient.GetUserList()
	if err != nil {
		log.Print(err)
	}
	if nodeInfoChanged {
		err = c.addNewUser(newUserInfo, newNodeInfo)
		if err != nil {
			log.Print(err)
		}
		// Add Limiter
		tag := fmt.Sprintf("%s_%d", c.nodeInfo.NodeType, c.nodeInfo.Port)
		if err := c.AddInboundLimiter(tag, newNodeInfo.SpeedLimit, newUserInfo); err != nil {
			log.Print(err)
		}
	} else {
		deleted, added := compareUserList(c.userList, newUserInfo)
		if len(deleted) > 0 {
			deletedEmail := make([]string, len(deleted))
			for i, u := range deleted {
				deletedEmail[i] = u.Email
			}
			tag := fmt.Sprintf("%s_%d", c.nodeInfo.NodeType, c.nodeInfo.Port)
			err := c.removeUsers(deletedEmail, tag)
			if err != nil {
				log.Print(err)
			}
		}
		if len(added) > 0 {
			err = c.addNewUser(&added, c.nodeInfo)
			if err != nil {
				log.Print(err)
			}
			// Update Limiter
			tag := fmt.Sprintf("%s_%d", c.nodeInfo.NodeType, c.nodeInfo.Port)
			if err := c.AddInboundLimiter(tag, newNodeInfo.SpeedLimit, &added); err != nil {
				log.Print(err)
			}
		}
		log.Printf("%d user deleted, %d user added", len(deleted), len(added))
	}
	c.userList = newUserInfo
	return nil
}

func (c *Controller) removeOldTag(oldtag string) (err error) {
	err = c.removeInbound(oldtag)
	if err != nil {
		return err
	}
	err = c.removeOutbound(oldtag)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) addNewTag(newNodeInfo *api.NodeInfo) (err error) {
	inboundConfig, err := InboundBuilder(newNodeInfo, c.config.CertConfig)
	if err != nil {
		return err
	}
	err = c.addInbound(inboundConfig)
	if err != nil {

		return err
	}
	outBoundConfig, err := OutboundBuilder(newNodeInfo)
	if err != nil {

		return err
	}
	err = c.addOutbound(outBoundConfig)
	if err != nil {

		return err
	}
	return nil
}

func (c *Controller) addNewUser(userInfo *[]api.UserInfo, nodeInfo *api.NodeInfo) (err error) {
	users := make([]*protocol.User, 0)
	if nodeInfo.NodeType == "V2ray" {
		if nodeInfo.EnableVless {
			users = buildVlessUser(userInfo)
		} else {
			users = buildVmessUser(userInfo, nodeInfo.AlterID)
		}
	} else if nodeInfo.NodeType == "Trojan" {
		users = buildTrojanUser(userInfo)
	} else if nodeInfo.NodeType == "Shadowsocks" {
		users = buildSSUser(userInfo)
	} else {
		return fmt.Errorf("Unsupported node type: %s", nodeInfo.NodeType)
	}
	tag := fmt.Sprintf("%s_%d", nodeInfo.NodeType, nodeInfo.Port)
	err = c.addUsers(users, tag)
	if err != nil {
		return err
	}
	log.Printf("Added %d new users", len(*userInfo))
	return nil
}

func compareUserList(old, new *[]api.UserInfo) (deleted, added []api.UserInfo) {
	msrc := make(map[api.UserInfo]byte) //按源数组建索引
	mall := make(map[api.UserInfo]byte) //源+目所有元素建索引

	var set []api.UserInfo //交集

	//1.源数组建立map
	for _, v := range *old {
		msrc[v] = 0
		mall[v] = 0
	}
	//2.目数组中，存不进去，即重复元素，所有存不进去的集合就是并集
	for _, v := range *new {
		l := len(mall)
		mall[v] = 1
		if l != len(mall) { //长度变化，即可以存
			l = len(mall)
		} else { //存不了，进并集
			set = append(set, v)
		}
	}
	//3.遍历交集，在并集中找，找到就从并集中删，删完后就是补集（即并-交=所有变化的元素）
	for _, v := range set {
		delete(mall, v)
	}
	//4.此时，mall是补集，所有元素去源中找，找到就是删除的，找不到的必定能在目数组中找到，即新加的
	for v := range mall {
		_, exist := msrc[v]
		if exist {
			deleted = append(deleted, v)
		} else {
			added = append(added, v)
		}
	}

	return deleted, added
}

func (c *Controller) userInfoMonitor() (err error) {
	// Get server status
	CPU, Mem, Disk, Uptime, err := serverstatus.GetSystemInfo()
	if err != nil {
		log.Print(err)
	}
	err = c.apiClient.ReportNodeStatus(
		&api.NodeStatus{
			CPU:    CPU,
			Mem:    Mem,
			Disk:   Disk,
			Uptime: Uptime,
		})
	if err != nil {
		log.Print(err)
	}
	// Get User traffic
	userTraffic := make([]api.UserTraffic, 0)
	for _, user := range *c.userList {
		up, down := c.getTraffic(user.Email)
		if up > 0 || down > 0 {
			userTraffic = append(userTraffic, api.UserTraffic{
				UID:      user.UID,
				Email:    user.Email,
				Upload:   up,
				Download: down})
		}
	}
	if len(userTraffic) > 0 {
		err = c.apiClient.ReportUserTraffic(&userTraffic)
		if err != nil {
			log.Print(err)
		}
	}

	// TODO Report Online info
	return nil
}
