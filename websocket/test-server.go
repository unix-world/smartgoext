// server

package main

import (
	"log"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const (
	SRV_ADDR string = "127.0.0.1:8887"
)

func main() {
	log.Println("Server address is", SRV_ADDR)

	http.ListenAndServe(SRV_ADDR, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			// handle error
		}
		go func() {
			defer conn.Close()

			for {
			//	msg, op, err := wsutil.ReadClientData(conn)
				msg, err := wsutil.ReadClientText(conn)
				if err != nil {
					// handle error
				}
				log.Println("Received Message from Client:", string(msg), conn.RemoteAddr())
			//	err = wsutil.WriteServerMessage(conn, op, msg)
				err = wsutil.WriteServerText(conn, msg)
				if err != nil {
					// handle error
				}
			}
		}()
	}))

	log.Println("------- end")

}
