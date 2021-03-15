// Package limiter is to control the links that go into the dispather
package limiter

import (
	sync "sync"
	"time"

	"github.com/RManLuo/XrayR/api"
	"github.com/juju/ratelimit"
)

type InboundInfo struct {
	Tag            string
	NodeSpeedLimit uint64
	UserInfo       *sync.Map // Key: Email value: api.UserInfo
	BucketHub      *sync.Map // key: Email, value: *ratelimit.Bucket
	UserOnlineIP   *sync.Map // Key: Email Value: *sync.Map: Key: IP, Value: UID
}

type Limiter struct {
	InboundInfo *sync.Map // Key: Tag, Value: *InboundInfo
}

func New() *Limiter {
	return &Limiter{
		InboundInfo: new(sync.Map),
	}
}

func (l *Limiter) AddInboundLimiter(tag string, nodeSpeedLimit uint64, userList *[]api.UserInfo) error {
	inboundInfo := &InboundInfo{
		Tag:            tag,
		NodeSpeedLimit: nodeSpeedLimit,
		BucketHub:      new(sync.Map),
		UserOnlineIP:   new(sync.Map),
	}
	userMap := new(sync.Map)
	for _, user := range *userList {
		userMap.Store(user.Email, &user)
	}
	inboundInfo.UserInfo = userMap
	l.InboundInfo.Store(tag, inboundInfo) // Replace the old inbound info
	return nil
}

func (l *Limiter) UpdateInboundLimiter(tag string, updatedNodeSpeedLimit uint64, updatedUserList *[]api.UserInfo) error {

	if value, ok := l.InboundInfo.Load(tag); ok {
		inboundInfo := value.(*InboundInfo)
		// Update Node info
		if inboundInfo.NodeSpeedLimit != updatedNodeSpeedLimit {
			inboundInfo.BucketHub = new(sync.Map)
		}
		inboundInfo.NodeSpeedLimit = updatedNodeSpeedLimit
		// Update User info
		for _, u := range *updatedUserList {
			inboundInfo.UserInfo.Store(u.Email, &u)
			limit := determineRate(updatedNodeSpeedLimit, u.SpeedLimit)                                              // If need the limit
			limiter := ratelimit.NewBucketWithQuantum(time.Duration(int64(time.Second)), int64(limit), int64(limit)) // Byte/s
			inboundInfo.BucketHub.Store(u.Email, limiter)
		}
	}
	return nil
}

func (l *Limiter) DeleteInboundLimiter(tag string) error {
	l.InboundInfo.Delete(tag)
	return nil
}

func (l *Limiter) GetUserBucket(tag string, email string) (*ratelimit.Bucket, bool) {
	if value, ok := l.InboundInfo.Load(tag); ok {
		inboundInfo := value.(*InboundInfo)
		nodeLimit := inboundInfo.NodeSpeedLimit
		var userLimit uint64 = 0
		if v, ok := inboundInfo.UserInfo.Load(email); ok {
			u := v.(*api.UserInfo)
			userLimit = u.SpeedLimit
		}
		limit := determineRate(nodeLimit, userLimit) // If need the limit
		if limit > 0 {
			limiter := ratelimit.NewBucketWithQuantum(time.Duration(int64(time.Second)), int64(limit), int64(limit)) // Byte/s
			if v, ok := inboundInfo.BucketHub.LoadOrStore(email, limiter); ok {
				bucket := v.(*ratelimit.Bucket)
				return bucket, true
			} else {
				return limiter, true
			}
		} else {
			return nil, false
		}
	} else {
		newError("Get Inbound Limiter information failed").AtDebug().WriteToLog()
		return nil, false
	}
}

// determineRate returns the minimum non-zero rate
func determineRate(nodeLimit, userLimit uint64) (limit uint64) {
	if nodeLimit == 0 || userLimit == 0 {
		if nodeLimit > userLimit {
			return nodeLimit
		} else if nodeLimit < userLimit {
			return userLimit
		} else {
			return 0
		}
	} else {
		if nodeLimit > userLimit {
			return userLimit
		} else if nodeLimit < userLimit {
			return nodeLimit
		} else {
			return nodeLimit
		}
	}
}
