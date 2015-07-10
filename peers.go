package main

import (
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

const defaultWSURL = "/_ws/"

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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var slaves map[string]Slave

type Slave struct {
	Name           string
	Connection     *websocket.Conn
	ActiveDownload int
}

func InitMaster() (err error) {
	slaves = make(map[string]Slave, 10)
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
