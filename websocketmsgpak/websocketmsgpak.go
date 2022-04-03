
// GO Lang :: SmartGo Extra / WebSocket Message Pack - Server / Client :: Smart.Go.Framework
// (c) 2020-2022 unix-world.org
// r.20220403.2025 :: STABLE

package websocketmsgpak

// REQUIRE: go 1.15 or later

import (
	"os"
	"os/signal"
	"runtime"

	"log"
	"fmt"
	"time"

	"crypto/subtle"
	"crypto/tls"
	"io/ioutil"
	"crypto/x509"

	"net/http"

	smart "github.com/unix-world/smartgo"
	uid   "github.com/unix-world/smartgo/uuid"
	b58   "github.com/unix-world/smartgo/base58"

	"github.com/unix-world/smartgoext/gorilla/websocket"
)


//-- msgpak


const VERSION = "r.20220403.2025"


type HandleMessagesFunc func(bool, string, string, string, string) string


type messagePack struct {
	Cmd        string `json:"cmd"`
	Data       string `json:"data"`
	CheckSum   string `json:"checksum"`
}


func msgPakComposeMessage(cmd string, data string, sharedPrivateKey string) (msg string, errMsg string) {
	//--
	cmd = smart.StrTrimWhitespaces(cmd)
	if(cmd == "") {
		return "", "MsgPak: Command is empty"
	} //end if
	//--
	var dataEnc string = smart.ThreefishEncryptCBC(data, sharedPrivateKey + ":" + smart.Sha384(cmd), false)
	sMsg := &messagePack{
		cmd,
		dataEnc,
		smart.Sha512(cmd + "\n" + dataEnc + "\n" + data),
	}
	//--
	return smart.DataArchive(smart.JsonEncode(sMsg)), ""
	//--
} //END FUNCTION


func msgPakParseMessage(msg string, sharedPrivateKey string) (msgStruct *messagePack, errMsg string) {
	//--
	msg = smart.StrTrimWhitespaces(msg)
	if(msg == "") {
		return nil, "MsgPak: Message is empty"
	} //end if
	//--
	msg = smart.DataUnArchive(msg)
	if(msg == "") {
		return nil, "MsgPak: Message Unarchiving FAILED"
	} //end if
	//--
	msg = smart.StrTrimWhitespaces(msg)
	if(msg == "") {
		return nil, "MsgPak: Message is empty after Unarchiving"
	} //end if
	//--
	D := smart.JsonDecode(msg)
	if(D == nil) {
		return nil, "MsgPak: Message Decoding FAILED"
	} //end if
	//--
	sMsg := &messagePack{
		D["cmd"].(string),
		D["data"].(string),
		D["checksum"].(string),
	}
	sMsg.Data = smart.ThreefishDecryptCBC(sMsg.Data, sharedPrivateKey + ":" + smart.Sha384(sMsg.Cmd), false)
	if(sMsg.CheckSum != smart.Sha512(sMsg.Cmd + "\n" + D["data"].(string) + "\n" + sMsg.Data)) {
		return nil, "MsgPak: Invalid Message Checksum"
	} //end if
	//--
	return sMsg, ""
	//--
} //END FUNCTION


func msgPakWriteMessage(conn *websocket.Conn, cmd string, data string, sharedPrivateKey string) (ok bool, errMsg string) {
	//--
	if(cmd == "") {
		return false, ""
	} //end if
	//--
	msg, errMsg := msgPakComposeMessage(cmd, data, sharedPrivateKey)
	if(errMsg != "") {
		return false, "MsgPak: Write Message Compose Error: " + errMsg
	} //end if
	if(msg == "") {
		return false, ""
	} //end if
	//--
	err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
	if(err != nil) {
		return false, "MsgPak: Errors encountered during write message to websocket: " + err.Error()
	} //end if
	//--
	return true, ""
	//--
} //END FUNCTION


func msgPakHandleMessage(conn *websocket.Conn, isServer bool, id string, remoteId string, message string, sharedPrivateKey string, handleMessagesFunc HandleMessagesFunc) (ok bool, errMsg string) {
	//--
	if(conn == nil) { // do not return any message in this case ...
		return false, "msgPakHandleMessage: No connection Provided ..."
	} //end if
	//--
	msg, errMsg := msgPakParseMessage(message, sharedPrivateKey)
	message = ""
	if(errMsg != "") {
		return false, errMsg
	} //end if
	//--
	var area string = "client"
	var rarea string = "server"
	if(isServer == true) {
		area = "server"
		rarea = "client"
	} //end if
	//--
	log.Println("[DEBUG] msgPakHandleMessage: Received Command to {" + area + "} from {" + rarea + "}[" + remoteId + "] `" + msg.Cmd + "`", "Data-Size:", len(msg.Data), "bytes")
	log.Println("[DATA] msgPakHandleMessage: Received Data: `", msg.Data, "`")
	//--
	/*
	handleMessagesFunc := func(isServer bool, id string, remoteId string, cmd string, data string) string {
		//--
		var answerMsg string = ""
		//--
		switch(cmd) { // see below how to implement commands ...
			default: // unhandled
				answerMsg = "<ERR:UNHANDLED> Error description goes here" // return an error answer
		} //end switch
		//--
		return answerMsg
		//--
	} //END FUNCTION
	*/
	//--
	var answerMsg string = ""
	var answerData string = "msgPakHandleMessage: Got Command from {" + area + "}[" + id + "]: " + msg.Cmd + "\n" + "Length=" + smart.ConvertIntToStr(len(msg.Data))
	switch(msg.Cmd) {
		case "<PING>": // ping (zero)
			if(isServer != true) {
				answerMsg = "<OK:PING>"
			} //end if else
			break;
		case "<PONG>": // pong (one)
			if(isServer == true) {
				answerMsg = "<OK:PONG>"
			} //end if else
			break;
		case "<OK:PING>", "<OK:PONG>", "<OK>":
			log.Println("[NOTICE] msgPakHandleMessage: " + msg.Cmd + " Command Confirmation ...")
			answerMsg = "" // no message to return
			break;
		case "<INFO>":
			if(smart.StrStartsWith(msg.Data, "<ERR:")) {
				log.Println("[ERROR] msgPakHandleMessage: " + msg.Data)
			} else {
				log.Println("[NOTICE] msgPakHandleMessage: " + msg.Cmd + " Command Status Info")
			} //end if
			answerMsg = ""
			break;
		case "<ERR>":
			answerMsg = "Invalid Message ! <ERR> is reserved for internal use ..."
			log.Println("[ERROR] msgPakHandleMessage: " + answerMsg)
			break;
		default: // custom handler or unhandled
			if(smart.StrStartsWith(msg.Cmd, "<ERR:")) {
				answerMsg = msg.Cmd
				if(smart.StrStartsWith(answerMsg, "<ERR:")) {
					log.Println("[ERROR] msgPakHandleMessage: " + answerMsg)
					answerData = answerMsg
					answerMsg = "<INFO>"
				} //end if
			} else {
				answerMsg = handleMessagesFunc(isServer, id, remoteId, msg.Cmd, msg.Data)
				if(smart.StrStartsWith(answerMsg, "<ERR:")) {
					log.Println("[ERROR] msgPakHandleMessage: " + answerMsg)
					answerData = answerMsg
					answerMsg = "<INFO>"
				} //end if
			} //end if else
	} //end switch
	if(smart.StrStartsWith(answerMsg, "<ERR:")) { // answers with ERRORS starts with "<ERR:" ; see the sample above ...
	//	return false, "msgPakHandleMessage: Failed to Handle `" + msg.Cmd + "` message ... ERROR: " + answerMsg
	} else if(answerMsg == "") { // there is no other message to be sent
		return true, ""
	} //end if
	//--
	_, errWrMsg := msgPakWriteMessage(conn, answerMsg, answerData, sharedPrivateKey)
	if(errWrMsg != "") {
		return false, errWrMsg
	} //end if
	//--
	return true, ""
	//--
} //END FUNCTION


//-- helper


func MsgPakGenerateUUID() string {
	//--
	var theTime string = ""
	dtObjUtc := smart.DateTimeStructUtc("")
	if(dtObjUtc.Status != "OK") {
		log.Println("[ERROR] MsgPak: Date Time Failed:", dtObjUtc.ErrMsg)
	} else {
		theTime = smart.ConvertInt64ToStr(dtObjUtc.Time)
	} //end if else
//	log.Println("[NOTICE] MsgPak/UUID Time Seed:", theTime)
	var uuid string = uid.Uuid1013Str(13) + "-" + uid.Uuid1013Str(10) + "-" + uid.Uuid1013Str(13);
	if(theTime != "") {
		uuid += "-" + b58.Encode([]byte(theTime))
	} //end if
	//--
	return uuid
	//--
} //END FUNCTION


//-- server


func MsgPakServerListenAndServe(serverID string, useTLS bool, certifPath string, httpAddr string, httpPort uint16, authUsername string, authPassword string, sharedEncPrivKey string, intervalMsgSeconds uint32, handleMessagesFunc HandleMessagesFunc) bool {

	//-- checks

	serverID = smart.StrTrimWhitespaces(serverID)
	if(serverID == "") {
		log.Println("[ERROR] MsgPak Server: Empty Server ID")
		return false
	} //end if

	certifPath = smart.StrTrimWhitespaces(certifPath)
	if((certifPath == "") || (smart.PathIsBackwardUnsafe(certifPath) == true)) {
		certifPath = "./ssl"
	} //end if
	certifPath = smart.PathGetAbsoluteFromRelative(certifPath)
	certifPath = smart.PathAddDirLastSlash(certifPath)

	httpAddr = smart.StrTrimWhitespaces(httpAddr)
	if((!smart.IsNetValidIpAddr(httpAddr)) && (!smart.IsNetValidHostName(httpAddr))) {
		log.Println("[ERROR] MsgPak Server: Empty or Invalid Bind Address")
		return false
	} //end if

	if(!smart.IsNetValidPortNum(int64(httpPort))) {
		log.Println("[ERROR] MsgPak Server: Empty or Invalid Bind Port")
		return false
	} //end if

	authUsername = smart.StrTrimWhitespaces(authUsername)
	if(authUsername == "") {
		log.Println("[ERROR] MsgPak Server: Empty Auth UserName")
		return false
	} //end if
	if((len(authUsername) < 5) || (len(authUsername) > 25)) { // {{{SYNC-GO-SMART-AUTH-USER-LEN}}}
		log.Println("[ERROR] MsgPak Server: Invalid Auth UserName Length: must be between 5 and 25 characters")
		return false
	} //end if

	// do not trim authPassword !
	if(smart.StrTrimWhitespaces(authPassword) == "") {
		log.Println("[ERROR] MsgPak Server: Empty Auth Password")
		return false
	} //end if
	if((len(smart.StrTrimWhitespaces(authPassword)) < 7) || (len(authPassword) > 30)) { // {{{SYNC-GO-SMART-AUTH-PASS-LEN}}}
		log.Println("[ERROR] MsgPak Server: Invalid Auth UserName Length: must be between 7 and 30 characters")
		return false
	} //end if

	sharedEncPrivKey = smart.StrTrimWhitespaces(sharedEncPrivKey)
	if(sharedEncPrivKey == "") {
		log.Println("[ERROR] MsgPak Server: Empty Auth UserName")
		return false
	} //end if
	if((len(sharedEncPrivKey) < 16) || (len(sharedEncPrivKey) > 256)) { // {{{SYNC-GO-SMART-SHARED-PRIV-KEY-LEN}}}
		log.Println("[ERROR] MsgPak Server: Invalid Auth UserName Length: must be between 16 and 256 characters")
		return false
	} //end if

	if(intervalMsgSeconds < 10) {
		log.Println("[ERROR] MsgPak Server: Min allowed Message Interval Seconds is 10 seconds but is set to:", intervalMsgSeconds)
		return false
	} else if(intervalMsgSeconds > 86400) {
		log.Println("[ERROR] MsgPak Server: Max allowed Message Interval Seconds is 86400 seconds (24 hours) but is set to:", intervalMsgSeconds)
		return false
	} //end if

	//-- #

	const srvConnReadLimit = 16 * 1000 * 1000 // 16 MB

	var srvWebSockUpgrader = websocket.Upgrader{
		ReadBufferSize:    16384,
		WriteBufferSize:   16384,
	//	EnableCompression: true,
	} // use default options

	var customMessageCmd string = ""
	var customMessageDat string = ""

	var crrMessageCmd string = ""
	var crrMessageDat string = ""

	const defaultMessageCmd = "<PING>"
	var defaultMessageDat = "PING, from the Server: `" + serverID + "`"

	srvBroadcastMsg := func(conn *websocket.Conn, rAddr string) {
		//--
		defer conn.Close()
		//--
		for {
			//--
			if(customMessageCmd != "") {
				crrMessageCmd = customMessageCmd
				crrMessageDat = customMessageDat
			} else {
				crrMessageCmd = defaultMessageCmd
				crrMessageDat = defaultMessageDat
			} //end if else
			//--
			log.Println("[NOTICE] Broadcasting " + crrMessageCmd + " Message to Client(s)")
			//--
			msg, errMsg := msgPakComposeMessage(crrMessageCmd, smart.JsonEncode(defaultMessageDat), sharedEncPrivKey)
			//--
			if(customMessageCmd != "") { // reset after send
				customMessageCmd = ""
				customMessageDat = ""
			} //end if
			//--
			if(errMsg != "") {
				//--
				log.Println("[ERROR] Send Message to Client:", errMsg)
				//--
				return
				//--
			} else {
				//--
				err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
				//--
				if(err != nil) {
					//--
					log.Println("[ERROR] Send Message to Client / Writing to websocket Failed:", err)
					//--
					return
					//--
				} else {
					//--
					log.Println("[OK] Send Message to Client:", rAddr)
					//--
				} //end if else
				//--
			} //end if else
			//--
			time.Sleep(time.Duration(intervalMsgSeconds) * time.Second)
			//--
		} //end for
		//--
	} //end function

	srvHandlerMsgPack := func(w http.ResponseWriter, r *http.Request) {
		//-- lock thread
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		//-- check auth
		var isAuth bool = false
		aUsr, aPass, aOK := r.BasicAuth()
		if(aOK != true) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized) // 401
			return
		} //end if
		if((aUsr == authUsername) && (aPass == authPassword)) {
			isAuth = true
		} //end if
		if(isAuth != true) {
			w.WriteHeader(http.StatusForbidden) // 403
			return
		} //end if
		//-- upgrade the raw HTTP connection to a websocket based one ; below method must check credentials
		srvWebSockUpgrader.CheckOrigin = func(r *http.Request) bool {
			if(isAuth != true) {
				return false
			} //end if
			return true
		} // this is for ths js client connected from another origin ...
		//--
		conn, err := srvWebSockUpgrader.Upgrade(w, r, nil)
		if(err != nil) {
			log.Println("[ERROR] Connection Upgrade Failed:", err)
			return
		} //end if
		conn.SetReadLimit(srvConnReadLimit)
		//--
		defer conn.Close()
		//--
		log.Println("New Connection to:", conn.LocalAddr(), "From:", r.RemoteAddr)
		//-- The event loop
		go srvBroadcastMsg(conn, r.RemoteAddr)
		var retMsg string = ""
		for {
			//--
			messageType, message, err := conn.ReadMessage()
			if(err != nil) {
				log.Println("[ERROR] Message Reading Failed:", err)
				break
			} //end if
			log.Println("[INFO] Got New Message from Client:", conn.LocalAddr())
			//--
			if(messageType == websocket.TextMessage) {
				log.Printf("[NOTICE] Message Received from Client {" + r.RemoteAddr + "}, Package Size: %d bytes\n", len(string(message)))
				ok, errMsg := msgPakHandleMessage(conn, true, serverID, r.RemoteAddr, string(message), sharedEncPrivKey, handleMessagesFunc)
				message = nil
				retMsg = "[OK] Valid Message from Client ..."
				if(ok != true) {
					if(errMsg == "") {
						errMsg = "Unknown Error !!!"
					} //end if
					retMsg = "[ERROR]: " + errMsg
				} //end if
				log.Println(retMsg, "ClientAddr:", r.RemoteAddr)
			} else {
				log.Println("[ERROR]: TextMessage is expected", "ClientAddr:", r.RemoteAddr)
			} //end if else
			//--
			retMsg = ""
			//--
		} //end for
		//--
	} //end function

	srvHandlerCustomMsg := func(w http.ResponseWriter, r *http.Request) {
		//--
		user, pass, ok := r.BasicAuth()
		if(!ok || subtle.ConstantTimeCompare([]byte(user), []byte(authUsername)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(authPassword)) != 1) {
			w.Header().Set("WWW-Authenticate", `Basic realm="MessagePak Server Custom Messages Area"`)
			w.WriteHeader(401) // status code must be after set headers
			w.Write([]byte("401 Unauthorized\n"))
			log.Printf("[WARNING] MessagePak Server :: AUTH.FAILED [%s %s %s] %s [%s] %s\n", r.Method, r.URL, r.Proto, "401", r.Host, r.RemoteAddr)
			return
		} //end if
		//--
		custommsg, ok := r.URL.Query()["msg"]
		if(!ok || (len(custommsg[0]) < 1) || (len(custommsg[0]) > 255) || (smart.StrTrimWhitespaces(custommsg[0]) == "") || (!smart.StrRegexMatchString(`^[_a-zA-Z0-9\-\.]+$`, custommsg[0]))) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400 Bad Request\n"))
			return
		} //end if
		//--
		customMessageCmd = "<" + smart.StrToUpper(smart.StrTrimWhitespaces(custommsg[0])) + ">"
		customMessageDat = smart.Sha1(customMessageCmd) // "" ; TODO: get also this from params
		//--
		w.Header().Set("Content-Type", smart.HTML_CONTENT_HEADER)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(smart.HtmlStaticTemplate("Custom Message", "", `<h1>Custom Message</h1>` + `<div class="operation_success">` + smart.EscapeHtml(customMessageCmd) + `</div>` + "\n" + `<div class="operation_important">` + smart.EscapeHtml(customMessageDat) + `</div>`)))
		//--
	} //end function

	srvHandlerHome := func(w http.ResponseWriter, r *http.Request) {
		//--
		w.Header().Set("Refresh", "10")
		w.Header().Set("Content-Type", smart.HTML_CONTENT_HEADER)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(smart.HtmlStaticTemplate("WS Server: HTTP(S)/WsMux", "", `<h1>WS Server: HTTP(S)/WsMux # ` + smart.EscapeHtml(VERSION) + `</h1>` + `<div><img width="48" height="48" src="data:image/svg+xml,` + smart.EscapeHtml(smart.EscapeUrl(smart.ReadAsset("svg/loading-spin.svg"))) + `"></div>` + `<hr>` + `<small>(c) 2021-2022 unix-world.org</small>`)))
		//--
	} //end function

	http.HandleFunc("/msgpak", srvHandlerMsgPack)
	http.HandleFunc("/msgsend", srvHandlerCustomMsg)
	http.HandleFunc("/", srvHandlerHome)

	var srvAddr string = httpAddr + fmt.Sprintf(":%d", httpPort)
	if(useTLS == true) {
		log.Println("Starting WS Server:", "wss://" + srvAddr + "/msgpak", "@", "HTTPS/WsMux/TLS", "#", VERSION)
		log.Println("[NOTICE] Certificates Path:", certifPath)
		go log.Fatal("[ERROR]", http.ListenAndServeTLS(srvAddr, certifPath + "cert.crt", certifPath + "cert.key", nil))
	} else {
		log.Println("Starting WS Server:", "ws://" + srvAddr + "/msgpak", "@", "HTTP/WsMux/Insecure", "#", VERSION)
		go log.Fatal("[ERROR]", http.ListenAndServe(srvAddr, nil))
	} //end if else

	return true

} //END FUNCTION


//-- client


func MsgPakClientListenAndConnectToServer(serverPool []string, clientID string, tlsMode string, authUsername string, authPassword string, sharedEncPrivKey string, intervalMsgSeconds uint32, intervalReconnectSeconds uint32, handleMessagesFunc HandleMessagesFunc) bool {

	//--

	if(serverPool == nil) {
		serverPool = []string{}
	} //end if

	clientID = smart.StrTrimWhitespaces(clientID)
	if(clientID == "") {
		log.Println("[ERROR] MsgPak Client: Empty Client ID")
		return false
	} //end if

	authUsername = smart.StrTrimWhitespaces(authUsername)
	if(authUsername == "") {
		log.Println("[ERROR] MsgPak Client: Empty Auth UserName")
		return false
	} //end if
	if((len(authUsername) < 5) || (len(authUsername) > 25)) { // {{{SYNC-GO-SMART-AUTH-USER-LEN}}}
		log.Println("[ERROR] MsgPak Client: Invalid Auth UserName Length: must be between 5 and 25 characters")
		return false
	} //end if

	// do not trim authPassword !
	if(smart.StrTrimWhitespaces(authPassword) == "") {
		log.Println("[ERROR] MsgPak Client: Empty Auth Password")
		return false
	} //end if
	if((len(smart.StrTrimWhitespaces(authPassword)) < 7) || (len(authPassword) > 30)) { // {{{SYNC-GO-SMART-AUTH-PASS-LEN}}}
		log.Println("[ERROR] MsgPak Client: Invalid Auth UserName Length: must be between 7 and 30 characters")
		return false
	} //end if

	sharedEncPrivKey = smart.StrTrimWhitespaces(sharedEncPrivKey)
	if(sharedEncPrivKey == "") {
		log.Println("[ERROR] MsgPak Client: Empty Auth UserName")
		return false
	} //end if
	if((len(sharedEncPrivKey) < 16) || (len(sharedEncPrivKey) > 256)) { // {{{SYNC-GO-SMART-SHARED-PRIV-KEY-LEN}}}
		log.Println("[ERROR] MsgPak Client: Invalid Auth UserName Length: must be between 16 and 256 characters")
		return false
	} //end if

	if(intervalMsgSeconds < 10) {
		log.Println("[ERROR] MsgPak Client: Min allowed Message Interval Seconds is 10 seconds but is set to:", intervalMsgSeconds)
		return false
	} else if(intervalMsgSeconds > 86400) {
		log.Println("[ERROR] MsgPak Client: Max allowed Message Interval Seconds is 86400 seconds (24 hours) but is set to:", intervalMsgSeconds)
		return false
	} //end if

	if(intervalReconnectSeconds < 30) {
		log.Println("[ERROR] MsgPak Client: Min allowed Reconnect Interval Seconds is 30 seconds but is set to:", intervalReconnectSeconds)
		return false
	} else if(intervalReconnectSeconds > 43200) {
		log.Println("[ERROR] MsgPak Client: Max allowed Reconnect Interval Seconds is 43200 seconds (12 hours) but is set to:", intervalReconnectSeconds)
		return false
	} //end if

	//--

	var done chan interface{}
	var interrupt chan os.Signal

	receiveHandler := func(conn *websocket.Conn, theServerAddr string) {
		//--
		defer smart.PanicHandler()
		//--
		if(conn == nil) {
			log.Println("[ERROR] Receive Failed:", "No Connection ...")
			return
		} //end if
		//--
		defer close(done)
		//--
		var retMsg string = ""
		for {
			//--
			messageType, message, err := conn.ReadMessage()
			if(err != nil) {
				log.Println("[ERROR] Message Receive Failed:", err)
				return
			} //end if
			//--
			log.Printf("[NOTICE] Message Received from Server, Package Size: %d bytes\n", len(string(message)))
			if(messageType == websocket.TextMessage) {
				//--
				var lenMsg int = len(string(message))
				ok, errMsg := msgPakHandleMessage(conn, false, clientID, theServerAddr, string(message), sharedEncPrivKey, handleMessagesFunc)
				message = nil
				//--
				log.Println("[INFO] Client Message sent to Server, Package Size:", lenMsg, "bytes")
				retMsg = "[OK] Valid Message from Server ..."
				//--
				if(ok != true) {
					if(errMsg == "") {
						errMsg = "Unknown Error !!!"
					} //end if
					retMsg = "[ERROR]: " + errMsg
				} //end if
				//--
				log.Println(retMsg, "ServerAddr:", theServerAddr)
				//--
			} else {
				//--
				log.Println("[ERROR]: TextMessage is expected from ServerAddr:", theServerAddr)
				//--
			} //end if
			//--
			retMsg = ""
			//--
		} //end for
	} //end function

	var connectedServers = map[string]*websocket.Conn{}

	connectToServer := func(addr string) {
		//--
		defer smart.PanicHandler()
		//--
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		//--
		log.Println("[NOTICE] Connecting to Server:", addr, "TLS-MODE:", tlsMode)
		//--
		addr = smart.StrTrimWhitespaces(addr)
		if(addr == "") {
			log.Println("[ERROR] Empty Server Address:", addr)
			return
		} //end if
		arrAddr := smart.Explode(":", addr)
		if(len(arrAddr) != 2) {
			log.Println("[ERROR] Invalid Server Address:", addr)
			return
		} //end if
		var httpAddr string = smart.StrTrimWhitespaces(arrAddr[0])
		var httpPort int64 = smart.ParseIntegerStrAsInt64(smart.StrTrimWhitespaces(arrAddr[1]))
		if((!smart.IsNetValidIpAddr(httpAddr)) && (!smart.IsNetValidHostName(httpAddr))) {
			log.Println("[ERROR] Invalid Server Address (Host):", addr)
			return
		} //end if
		if(!smart.IsNetValidPortNum(httpPort)) {
			log.Println("[ERROR] Invalid Server Address (Port):", addr)
			return
		} //end if
		//--
		socketPrefix := "ws://"
		socketSuffix := "/msgpak"
		var securewebsocket websocket.Dialer
		if(tlsMode == "tls") {
			socketPrefix = "wss://"
			roots := x509.NewCertPool()
			var rootPEM string = ""
			crt, errCrt := ioutil.ReadFile("./cert.crt")
			if(errCrt != nil) {
				log.Fatal("[ERROR] Failed to read root certificate CRT")
			} //end if
			key, errKey := ioutil.ReadFile("./cert.key")
			if(errKey != nil) {
				log.Fatal("[ERROR] to read root certificate KEY")
			} //end if
			rootPEM = string(crt) + "\n" + string(key)
			ok := roots.AppendCertsFromPEM([]byte(rootPEM))
			if(!ok) {
				log.Fatal("[ERROR] Failed to parse root certificate")
			} //end if
			log.Println("Starting Client:", socketPrefix + addr + socketSuffix, "@", "HTTPS/WsMux/TLS")
			securewebsocket = websocket.Dialer{TLSClientConfig: &tls.Config{RootCAs: roots}}
		} else if(tlsMode == "tls:noverify") {
			socketPrefix = "wss://"
			log.Println("Starting Client:", socketPrefix + addr + socketSuffix, "@", "HTTPS/WsMux/TLS:NoVerify")
			securewebsocket = websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		} else { // insecure
			log.Println("Starting Client:", socketPrefix + addr + socketSuffix, "@", "HTTP/WsMux/Insecure")
			securewebsocket = websocket.Dialer{}
		} //end if else
		h := http.Header{"Authorization": {"Basic " + smart.Base64Encode(authUsername + ":" + authPassword)}}
	//	h = nil
		conn, response, err := securewebsocket.Dial(socketPrefix + addr + socketSuffix, h)
	//	conn, response, err := websocket.DefaultDialer.Dial(socketPrefix + addr + socketSuffix, h)
		if(err != nil) {
			var rStatusCode int = 0;
			if(response != nil) {
				rStatusCode = response.StatusCode
			} //end if
			log.Println("[ERROR] Cannot connect to Websocket Server: HTTP Response StatusCode:", rStatusCode, "; Dial Errors:", err)
			if(conn != nil) {
				conn.Close()
			} //end if
			return
		} //end if
		defer conn.Close()
		connectedServers[addr] = conn
		go receiveHandler(conn, addr)
		//-- the main loop for the client
		for {
			//--
			select {
				case <-time.After(time.Duration(intervalMsgSeconds) * time.Second):
					log.Println("[NOTICE] Sending <PONG> Message to Server")
					msg, errMsg := msgPakComposeMessage("<PONG>", smart.JsonEncode("PONG, from Client: `" + clientID + "`"), sharedEncPrivKey)
					if(errMsg != "") {
						log.Println("[ERROR]:", errMsg)
						delete(connectedServers, addr)
						return
					} else {
						err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
						if(err != nil) {
							log.Println("[ERROR] Writing to websocket Failed:", err)
							delete(connectedServers, addr)
							return
						} //end if
					} //end if else
					msg = ""
					errMsg = ""
				case <-interrupt: // received a SIGINT (Ctrl + C). Terminate gracefully...
					log.Println("[NOTICE] Received SIGINT interrupt signal. Closing all pending connections")
					err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")) // Close our websocket connection
					if(err != nil) {
						log.Println("[ERROR] Closing websocket Failed:", err)
					} //end if
					//-- possible fix
				//	delete(connectedServers, addr)
				//	return
					// fix: if crashes comment below and uncomment above 2 lines
					select {
						case <-done:
							log.Println("[NOTICE] Receiver Channel Closed...")
						case <-time.After(time.Duration(1) * time.Second):
							log.Println("[WARNING] Timeout in closing receiving channel...")
					} //end select
					delete(connectedServers, addr)
					return
					//-- #end fix
			} //end select
			//--
		} //end for
		//--
	} //end function

	watchdog := func() {
		//--
		log.Println("Starting WS Client", "#", VERSION)
		//--
		var initConn bool = false
		//--
		for {
			//--
			log.Println("[NOTICE] Client Connection WATCHDOG is up and running ...")
			log.Println("[DEBUG] Connected Servers:", connectedServers)
			//--
			for _, p := range serverPool {
				if _, exist := connectedServers[p]; exist {
					log.Println("[OK] Client is Connected to Server:", p)
				} else {
					if(initConn == true) {
						log.Println("[WARNING] Client is NOT Connected to Server:", p)
					} //end if
					go connectToServer(p)
				} //end if else
			} //end for
			//--
			initConn = true
			//--
			time.Sleep(time.Duration(intervalReconnectSeconds) * time.Second)
			//--
		} //end for
	} //end function

	done = make(chan interface{}) // Channel to indicate that the receiveHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully
	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	go watchdog()

	return true

} //END FUNCTION


// #END
