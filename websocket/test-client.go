// client

package main

import (
	"log"
	"context"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const (
	SRV_ADDR string = "127.0.0.1:8887"
)

func main() {
	log.Println("Connect to server", SRV_ADDR)

	ctx := context.Background()
	conn, _, _, err := ws.DefaultDialer.Dial(ctx, "ws://" + SRV_ADDR)
	if err != nil {
		log.Println("ERR:", err)
		return
	}
	defer conn.Close()

	for {
	//	errRd := wsutil.WriteClientMessage(conn, ws.OpText, []byte("just a message ..."))
		errRd := wsutil.WriteClientText(conn, []byte("just a message ..."))
		if(errRd != nil) {
			log.Println("ERR/Read:", errRd)
			return
		}

	//	msg, op, errWr := wsutil.ReadServerData(conn)
		msg, errWr := wsutil.ReadServerText(conn)
		if(errWr != nil) {
			log.Println("ERR/Write:", errWr)
			return
		}
		log.Println("Received Message from Server:", string(msg))

		time.Sleep(time.Duration(1) * time.Second)
	}

	log.Println("------- end")

}
