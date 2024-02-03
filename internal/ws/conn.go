package ws

import (
	"github.com/gorilla/websocket"
	"log/slog"
	"time"
)

type ConnHold struct {
	mode       string
	connId     string
	conn       *websocket.Conn
	exitCh     chan string
	exitHandle func(signal string)
	readCh     chan *UpgradePacket
	writeCh    chan *UpgradePacket
}

func (h *ConnHold) free() {
	close(h.readCh)
	close(h.writeCh)
	close(h.exitCh)
}

func (h *ConnHold) start() {
	//read
	go func() {
		for {
			msgType, msg, err := h.conn.ReadMessage()
			//连接关闭，发送退出信号
			if err != nil {
				h.exitCh <- err.Error()
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
				slog.Info("send message", "mode", h.mode, "client_id", h.connId, "data", event.Byte)
				_ = h.conn.WriteMessage(event.Type, event.Byte)
			case <-time.After(10 * time.Second):
			}
		}
	}()

	//退出事件
	go func() {
		text := <-h.exitCh
		slog.Info("exit", "mode", h.mode, "connId", h.connId, "reason", text)
		h.exitHandle(text)
	}()
}

type UpgradePacket struct {
	Type int
	Byte []byte
}

func newConnHold(conn *websocket.Conn, mode string, connId string) *ConnHold {
	return &ConnHold{
		mode:    mode,
		connId:  connId,
		conn:    conn,
		exitCh:  make(chan string),
		readCh:  make(chan *UpgradePacket, 1),
		writeCh: make(chan *UpgradePacket, 1),
	}
}
