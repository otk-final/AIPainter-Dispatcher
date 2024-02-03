package ws

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/samber/lo"
	"log/slog"
	"net/url"
	"sync"
	"time"
)

var upgrade = websocket.Upgrader{
	HandshakeTimeout: time.Second * 30,
}

var DISPATCHER = make(map[string]*UpgradeHold)
var MUTEX = &sync.Mutex{}

type UpgradeHold struct {
	userId string
	mutex  *sync.Mutex
	master *ConnHold
	slaves []*ConnHold
}

func (h *UpgradeHold) NewBind(target *url.URL) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	//是否存在
	_, ok := lo.Find(h.slaves, func(item *ConnHold) bool {
		return item.mode == target.Host && item.connId == h.userId
	})
	if ok {
		return
	}

	//链接主机地址
	wsConn, _, err := websocket.DefaultDialer.DialContext(context.Background(), target.String(), nil)
	if err != nil {
		return
	}

	//启动
	hold := newConnHold(wsConn, target.Host, h.userId)
	hold.start()

	//监听从链接内容，发送至主链接
	go func(hold *ConnHold) {
		defer hold.free()
		for {
			select {
			case packet := <-hold.readCh:
				h.master.writeCh <- packet
			case <-time.After(1 * time.Minute):
				//空闲，则释放
				h.slaves = lo.Filter(h.slaves, func(item *ConnHold, index int) bool {
					return item.mode != hold.mode && item.connId != hold.connId
				})
			}
		}
	}(hold)

	//cache
	h.slaves = append(h.slaves, hold)
}

func (h *UpgradeHold) Free() {
	h.master.free()
	for _, s := range h.slaves {
		s.free()
	}
	delete(DISPATCHER, h.userId)
}

func Free(id string) {
	if hold, ok := DISPATCHER[id]; ok {
		hold.Free()
	} else {
		slog.Warn("not found connect", id)
	}
}

func GetHold(userId string) *UpgradeHold {
	MUTEX.Lock()
	defer MUTEX.Unlock()

	return DISPATCHER[userId]
}
