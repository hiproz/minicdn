package main

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const defaultWSURL = "/_ws/"

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	state = ServerState{
		ActiveDownload: 0,
		Closed:         false,
	}
	slaves = make(map[string]Slave, 10)
)

type Slave struct {
	Name           string
	Connection     *websocket.Conn
	ActiveDownload int
}

type ServerState struct {
	sync.Mutex
	ActiveDownload int
	Closed         bool
}

func (s *ServerState) addActiveDownload(n int) {
	s.Lock()
	defer s.Unlock()
	s.ActiveDownload += n
}

func (s *ServerState) Close() error {
	s.Closed = true
	time.Sleep(time.Millisecond * 5000) // 0.5s
	for {
		if s.ActiveDownload == 0 { // Wait until all download finished
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	return nil
}

func InitSlave() (err error) {
	u, err := url.Parse(*upstream)
	if err != nil {
		return
	}
	u.Path = defaultWSURL
	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return
	}
	client, _, err := websocket.NewClient(conn, u, nil, 1024, 1024)
	if err != nil {
		return
	}
	client.WriteMessage(1, []byte("hello"))
	return nil
}

func InitMaster() (err error) {
	http.HandleFunc(defaultWSURL, WSHandler)
	return nil
}

func init() {
	//var me = "http://127.0.0.1:5000"
	//var peers = groupcache.NewHTTPPool(me)

	//peers.Set("http://
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(conn.RemoteAddr())
	defer conn.Close()

	name := conn.RemoteAddr().String()
	slave := Slave{
		Name:           name,
		Connection:     conn,
		ActiveDownload: 0,
	}
	slaves[name] = slave

	for {
		msgType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read Msg ERR:", err)
			return
		}
		log.Println(msgType, string(p))
	}
}
