package main
import (
    "github.com/gorilla/websocket"
)
// ここで使っているチャネルはバッファのあるチャネルで，メッセージのための待ち行列のようなものです，
// メモリ上に配置され，
// スレッドセーフな性質を備えています．
// 複数の送信者や受信者が同時に読み書きでき，
// ブロックされる事はありません．
// (バッファのないチャネルもしくはバッファの空きがなければブロックされ，送信と受信の同期が行われます．)


//clientはチャットを行っている1人のユーザーを表します
type client struct{
    // socketはこのクライアントのためのWebSocketです．
    socket *websocket.Conn
    // sendはメッセージが送られるチャネルです．
    send chan []byte
    // roomはこのクライアントが参加しているチャットルームです．
    room *room
}

//WebSocketの読み書きを行うメソッド
func (c *client) read() {
    for {
        if _, msg, err := c.socket.ReadMessage(); err==nil {
            c.room.forward <- msg
        }else {//webSocketの異常終了などが原因でエラーが発生した場合
            break
        }
    }
    c.socket.Close()
}
func (c *client) write() {
    for msg := range c.send {
        if err := c.socket.WriteMessage(websocket.TextMessage, msg);
            err!=nil{//webSocketへの書き込みが失敗した場合
            break
        }
    }
    c.socket.Close()
}