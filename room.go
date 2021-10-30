package main

import (
	"net/http"
    "github.com/gorilla/websocket"
	"log"
)

type room struct {
    // forward は他のクライアントに転送するためのメッセージを保持するチャネルです．
    forward chan []byte
    // joinはチャットルームに参加しようとしているクライアントのためのチャネルです．
    join chan *client
    // leaveはチャットルームから退室しようとしているクライアントのためのチャネルです．
    leave chan *client
    // clientsには在室しているすべてのクライアントが保持されます．
    clients map[*client]bool
}
func (r *room) run() {
    //無限ループ
    // goroutineとして実行される場合は，アプリ内の他の処理をブロックすることがないため大丈夫
    for {
        // select文は強力な並行処理の機能
        // 共有化されているメモリに対して同期化や変更がいくつか必要な任意の箇所でselect文を利用できる
        // チャネルに送信された値に応じて，異なる操作を行うことも可能

        // このソースコードでは，join,leave,forwardの3つのチャンネルを監視
        // いずれかのチャネルにメッセージが届くと，select文の中でそれぞれに対応するcase文が実行される
        // このcase節のコードは同時に実行されることはないため，マップr.clientsへの変更が同時に発生するということが防がれている
        select {
        case client := <- r.join:
            //参加
            r.clients[client] = true
        case client := <- r.leave:
            //退室
            delete(r.clients, client)
            close(client.send)
        case msg := <-r.forward:
            //すべてのクライアントにメッセージを転送
            for client := range r.clients{
                select {
                case client.send <- msg:
                    //メッセージを送信
                default:
                    //送信に失敗
                    delete(r.clients, client)
                    close(client.send)
                }
            }
        }
    }
}
const (
	socketBufferSize = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:	socketBufferSize, 
	WriteBufferSize: socketBufferSize,
}

func (r* room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:",err)
		return
	}
	client := &client {
		socket:socket,
		send:make(chan []byte, messageBufferSize),
		room: r,
	}
	r.join <- client
	defer func() {r.leave <- client } ()
	go client.write() //goroutineとして実行される
	client.read()
}
// newRoomはすぐに利用できるチャットルームを生成して返します．
func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join: make(chan *client),
		leave: make(chan *client),
		clients: make(map[*client] bool),
	}
}