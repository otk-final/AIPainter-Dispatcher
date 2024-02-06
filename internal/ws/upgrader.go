package ws

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/samber/lo"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var upgrade = websocket.Upgrader{
	HandshakeTimeout: time.Second * 30,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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

	rawValues := &url.Values{}
	rawValues.Add("client_id", h.userId)
	wsURL := &url.URL{
		Scheme:   "ws",
		Host:     target.Host,
		Path:     "/ws",
		RawQuery: rawValues.Encode(),
	}

	//链接主机地址
	slog.Info("ready connect service", wsURL.String())
	wsConn, _, err := websocket.DefaultDialer.DialContext(context.Background(), wsURL.String(), nil)
	if err != nil {
		slog.Error("websocket dail error", err)
		return
	}

	//启动
	hold := newConnHold(wsConn, target.Host, h.userId)
	hold.exitHandle = func(text string) {
		//移除当前
		h.slaves = lo.Filter(h.slaves, func(item *ConnHold, index int) bool {
			return item.mode != hold.mode && item.connId != hold.connId
		})
		//关闭连接
		hold.free()
	}
	hold.start()

	//监听从链接内容，发送至主链接
	go func(hold *ConnHold) {
		for {
			select {
			case packet, ok := <-hold.readCh:
				//slog.Info("read service packet", packet)
				//发送至主链接
				if ok && packet != nil {
					h.master.writeCh <- packet
				}
			case <-time.After(30 * time.Second):
				//1分钟无响应，则释放用户和主机直接的链接
				_ = hold.conn.Close()
				return
			}
		}
	}(hold)

	//cache
	h.slaves = append(h.slaves, hold)
}

func NewUpgrade(writer http.ResponseWriter, request *http.Request) {
	vars := request.URL.Query()
	clientId := vars.Get("client_id")

	//new client
	clientConn, err := upgrade.Upgrade(writer, request, request.Header)
	if err != nil {
		http.Error(writer, "client upgrade err", http.StatusInternalServerError)
		return
	}
	slog.Info("user address", " clientId", clientId, "host", clientConn.RemoteAddr())

	MUTEX.Lock()
	defer MUTEX.Unlock()

	master := newConnHold(clientConn, "user", clientId)
	master.exitHandle = func(text string) {
		//释放当前连接，并通知子连接
		FreeHold(clientId)
	}
	master.start()
	//cache
	DISPATCHER[clientId] = &UpgradeHold{
		userId: clientId,
		mutex:  &sync.Mutex{},
		master: master,
		slaves: make([]*ConnHold, 0),
	}
}

func FreeHold(id string) {
	MUTEX.Lock()
	defer MUTEX.Unlock()

	hold, ok := DISPATCHER[id]
	if ok {
		hold.master.free()
		for _, s := range hold.slaves {
			_ = s.conn.Close()
		}
		delete(DISPATCHER, hold.userId)
	}
}

func FindHold(id string) *UpgradeHold {
	MUTEX.Lock()
	defer MUTEX.Unlock()
	return DISPATCHER[id]
}
