package ws

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type ConnHold struct {
	mode    string
	connId  string
	conn    *websocket.Conn
	readCh  chan *UpgradePacket
	writeCh chan *UpgradePacket
}

func (h *ConnHold) free() {
	_ = h.conn.Close()
	close(h.readCh)
	close(h.writeCh)

	//remove
	delete(DISPATCHER, h.Key())
}

func (h *ConnHold) Key() string {
	return fmt.Sprintf("%s_%s", h.mode, h.connId)
}

func (h *ConnHold) start() {
	//read
	go func() {
		for {
			msgType, msg, err := h.conn.ReadMessage()
			if err != nil {
				return
			}
			h.readCh <- &UpgradePacket{
				Type: msgType,
				Byte: msg,
			}
		}
	}()

	//write
	go func() {
		for {
			select {
			case event, ok := <-h.writeCh:
				if !ok {
					return
				}
				_ = h.conn.WriteMessage(event.Type, event.Byte)
			case <-time.After(10 * time.Second):
			}
		}
	}()
}

type UpgradePacket struct {
	Type int
	Byte []byte
}

func NewUpgrade(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	clientId := vars["client_id"]

	//new client
	clientConn, err := upgrade.Upgrade(writer, request, request.Header)
	if err != nil {
		http.Error(writer, "client upgrade err", http.StatusInternalServerError)
		return
	}

	MUTEX.Lock()
	defer MUTEX.Unlock()

	master := newConnHold(clientConn, "user", clientId)
	master.start()

	//cache
	DISPATCHER[clientId] = &UpgradeHold{
		userId: clientId,
		mutex:  &sync.Mutex{},
		master: master,
		slaves: make([]*ConnHold, 0),
	}
}

func newConnHold(conn *websocket.Conn, mode string, connId string) *ConnHold {
	return &ConnHold{
		mode:    mode,
		connId:  connId,
		conn:    conn,
		readCh:  make(chan *UpgradePacket, 1),
		writeCh: make(chan *UpgradePacket, 1),
	}
}
