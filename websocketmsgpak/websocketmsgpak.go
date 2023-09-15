
// GO Lang :: SmartGo Extra / WebSocket Message Pack - Server / Client :: Smart.Go.Framework
// (c) 2020-2023 unix-world.org
// r.20230915.0918 :: STABLE

// Req: go 1.16 or later (embed.FS is N/A on Go 1.15 or lower)
package websocketmsgpak

import (
	"os"
	"os/signal"
	"sync"

	"log"
	"fmt"
	"time"

	"net/http"

	fifolist "container/list"

	smart 			"github.com/unix-world/smartgo"
	uid 			"github.com/unix-world/smartgo/uuid"
	b58 			"github.com/unix-world/smartgo/base58"
	assets 			"github.com/unix-world/smartgo/web-assets"
	srvassets 		"github.com/unix-world/smartgo/web-srvassets"
	smarthttputils 	"github.com/unix-world/smartgo/web-httputils"
	smartcache 		"github.com/unix-world/smartgo/simplecache"

	dhkx 			"github.com/unix-world/smartgo/dhkx"
	websocket 		"github.com/unix-world/smartgoext/gorilla/websocket"
	crontab 		"github.com/unix-world/smartgoext/crontab"
)


//-- msgpak


const (
	VERSION string = "r.20230915.0918"

	DEBUG bool = false
	DEBUG_CACHE bool = false

	HANDSHAKE_TIMEOUT_SECONDS uint32 	= 45 		// default is 45
	WAIT_DHKX_LIMIT_SECONDS  uint32 	= 60 		// default is 60
	WAIT_CLOSE_LIMIT_SECONDS uint32 	=  2 		// default is 2
	MAX_NUM_ERR_MSG_RECV uint8 			=  8 		// default is 8 ; after this number of reecive errors, reset connection and force re-connect ; Implemented just for Client ; TODO: implement also for Server

	MAX_META_MSG_SIZE uint32 	=  1 * 1000 * 1000 	//  1 MB
	MAX_MSG_SIZE uint32 		= 16 * 1000 * 1000 	// 16 MB
	MAX_QUEUE_MESSAGES uint8 	= 100 				// must be between: 1..255

	LIMIT_INTERVAL_SECONDS_MIN uint32 = 10 			// {{{SYNC-MSGPAK-INTERVAL-LIMITS}}}
	LIMIT_INTERVAL_SECONDS_MAX uint32 = 3600 		// {{{SYNC-MSGPAK-INTERVAL-LIMITS}}}

	CERTIFICATES_DEFAULT_PATH string = "./ssl"
	CERTIFICATE_PEM_CRT string = "cert.crt"
	CERTIFICATE_PEM_KEY string = "cert.key"

	HTTP_AUTH_REALM string = "Smart.MsgPak Server"
)


type CronMsgTask struct {
	Timing string
	Cmd string
	Data string
}

type HandleMessagesFunc func(bool, string, string, string, string, string, string) (string, string)


type messagePack struct {
	Cmd        string `json:"cmd"`
	Data       string `json:"data"`
	CheckSum   string `json:"checksum"`
}


//-----


var websockWriteMutex sync.Mutex // connections allow concurrent reads but not concurrent writes, thus protect writes with a simple mutex (not with a RWMutex)

func connCloseSocket(conn *websocket.Conn) {
	//--
	defer smart.PanicHandler()
	//--
	if(conn != nil) {
		conn.Close()
		conn = nil
	} //end if
	//--
} //END FUNCTION


func connWriteCloseMsgToSocket(conn *websocket.Conn, msg []byte) error {
	//--
	defer smart.PanicHandler()
	//--
	websockWriteMutex.Lock()
	defer websockWriteMutex.Unlock()
	//--
	if(conn == nil) {
		return smart.CreateNewError("WARNING: Cannot write CloseMsg to Empty Connection")
	} //end if
	//--
	conn.SetWriteDeadline(time.Now().Add(time.Duration(WAIT_CLOSE_LIMIT_SECONDS) * time.Second))
	return conn.WriteMessage(websocket.CloseMessage, msg)
	//--
} //END FUNCTION


func connWriteTxtMsgToSocket(conn *websocket.Conn, msg []byte, maxLimitSeconds uint32) error {
	//--
	defer smart.PanicHandler()
	//--
	websockWriteMutex.Lock()
	defer websockWriteMutex.Unlock()
	//--
	if(conn == nil) {
		return smart.CreateNewError("WARNING: Cannot write TxtMsg to Empty Connection")
	} //end if
	//--
	if(maxLimitSeconds < LIMIT_INTERVAL_SECONDS_MIN) { // {{{SYNC-MSGPAK-INTERVAL-LIMITS}}}
		maxLimitSeconds = LIMIT_INTERVAL_SECONDS_MIN
	} else if(maxLimitSeconds > LIMIT_INTERVAL_SECONDS_MAX) { // {{{SYNC-MSGPAK-INTERVAL-LIMITS}}}
		maxLimitSeconds = LIMIT_INTERVAL_SECONDS_MAX
	} //end if
	//--
	conn.SetWriteDeadline(time.Now().Add(time.Duration(int(maxLimitSeconds - 1)) * time.Second))
	return conn.WriteMessage(websocket.TextMessage, msg)
	//--
} //END FUNCTION


func connReadFromSocket(conn *websocket.Conn, maxLimitSeconds uint32) (msgType int, msg []byte, err error) {
	//--
	defer smart.PanicHandler()
	//--
	if(conn == nil) {
		return -1, nil, smart.CreateNewError("WARNING: Cannot read Msg from Empty Connection")
	} //end if
	//--
	if(maxLimitSeconds < LIMIT_INTERVAL_SECONDS_MIN) { // {{{SYNC-MSGPAK-INTERVAL-LIMITS}}}
		maxLimitSeconds = LIMIT_INTERVAL_SECONDS_MIN
	} else if(maxLimitSeconds > LIMIT_INTERVAL_SECONDS_MAX) { // {{{SYNC-MSGPAK-INTERVAL-LIMITS}}}
		maxLimitSeconds = LIMIT_INTERVAL_SECONDS_MAX
	} //end if
	//--
	conn.SetReadLimit(int64(MAX_MSG_SIZE + MAX_META_MSG_SIZE))
	conn.SetReadDeadline(time.Now().Add(time.Duration(int(maxLimitSeconds + 1)) * time.Second))
	//--
	messageType, message, rdErr := conn.ReadMessage()
	//--
	return messageType, message, rdErr
	//--
} //END FUNCTION


//-----


func msgPakComposeMessage(cmd string, data string, sharedPrivateKey string, sharedSecret string) (msg string, errMsg string) {
	//--
	defer smart.PanicHandler()
	//--
	cmd = smart.StrTrimWhitespaces(cmd)
	if(cmd == "") {
		return "", "MsgPak: Command is empty"
	} //end if
	//--
	var dataEnc string = smart.ThreefishEncryptCBC(data, sharedPrivateKey + sharedSecret + ":" + smart.Sha384(cmd), false)
	sMsg := &messagePack{
		cmd,
		dataEnc,
		smart.Sha512(cmd + "\n" + dataEnc + "\n" + data),
	}
	//--
	dataEnc = smart.DataArchive(smart.JsonNoErrChkEncode(sMsg, false, true))
	sMsg = nil
	//--
	var crrLen int = len(dataEnc)
	if(crrLen > int(MAX_MSG_SIZE)) {
		return "", "MsgPak: Package is Oversized: Max allowed is: " + smart.ConvertIntToStr(int(MAX_MSG_SIZE)) + " < Current Package is: " + smart.ConvertIntToStr(crrLen) + " bytes"
	} //end if
	//--
	return dataEnc, ""
	//--
} //END FUNCTION


func msgPakParseMessage(msg string, sharedPrivateKey string, sharedSecret string) (msgStruct *messagePack, errMsg string) {
	//--
	defer smart.PanicHandler()
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
	D, DErr := smart.JsonObjDecode(msg)
	if((DErr != nil) || (D == nil)) {
		return nil, "MsgPak: Message Decoding FAILED: " + DErr.Error()
	} //end if
	//--
	sMsg := &messagePack{
		D["cmd"].(string),
		D["data"].(string),
		D["checksum"].(string),
	}
	sMsg.Data = smart.ThreefishDecryptCBC(sMsg.Data, sharedPrivateKey + sharedSecret + ":" + smart.Sha384(sMsg.Cmd), false)
	if(sMsg.CheckSum != smart.Sha512(sMsg.Cmd + "\n" + D["data"].(string) + "\n" + sMsg.Data)) {
		return nil, "MsgPak: Invalid Message Checksum"
	} //end if
	//--
	return sMsg, ""
	//--
} //END FUNCTION


func msgPakWriteMessage(conn *websocket.Conn, maxLimitSeconds uint32, cmd string, data string, sharedPrivateKey string, sharedSecret string) (ok bool, msgSize int, errMsg string) {
	//--
	defer smart.PanicHandler()
	//--
	cmd = smart.StrTrimWhitespaces(cmd)
	if(cmd == "") {
		return false, 0, ""
	} //end if
	//--
	msg, errMsg := msgPakComposeMessage(cmd, data, sharedPrivateKey, sharedSecret)
	if(errMsg != "") {
		return false, 0, "MsgPak: Write Message Compose Error: " + errMsg
	} //end if
	if(msg == "") {
		return false, 0, ""
	} //end if
	//--
	err := connWriteTxtMsgToSocket(conn, []byte(msg), maxLimitSeconds)
	if(err != nil) {
		return false, 0, "MsgPak: Errors encountered during write message to websocket: " + err.Error()
	} //end if
	//--
	return true, len(msg), ""
	//--
} //END FUNCTION


func dhkxCliHandler(remoteId string, isServer bool, cmd string, data string) (answerMsg string, answerData string, extraData string) {
	//--
	if(isServer == true) {
		return "<ERR:DHKX:CLI>", "Invalid Server Command: " + cmd, ""
	} //end if
	if(cmd != "<DHKX:CLI>") {
		return "<ERR:DHKX:CLI>", "Invalid Command: " + cmd, ""
	} //end if
	//--
	var clientRecvDhKxFromServer dhkx.HandleDhkxCliRecvFunc = func() (string, []byte, int) {
		//--
		arr := smart.Explode(":", smart.StrTrimWhitespaces(data))
		if(len(arr) != 2) {
			return "Invalid Message Format", nil, 0
		} //end if
		//--
		var grpId int = int(smart.ParseStrAsInt64(arr[0]))
		if(!dhkx.DhKxValidateGroup(grpId)) {
			return "Invalid Message Format: Group: " + arr[0], nil, 0
		} //end if
		var srvPubKey []byte = smart.BaseDecode(arr[1], "b62")
		//--
		return "", srvPubKey, grpId
		//--
	} //END FUNCTION
	//--
	var clientSendDhKxToServer dhkx.HandleDhkxCliSendFunc = func(cliPubKey []byte, cliExch []byte) string {
		//--
		// This will be handled back by dhkxCliHandler -> msgPakHandleMessage
		//--
		return ""
		//--
	} //END FUNCTION
	//--
	errCliRecvSend1, grpCli, privCli, pubCli, recvPubSrv, shardCli, shardExch := dhkx.DhKxClientExchange(clientRecvDhKxFromServer, clientSendDhKxToServer)
	if(errCliRecvSend1 != "") {
		return "<ERR:DHKX:CLI>", errCliRecvSend1, ""
	} //end if
	if(grpCli == nil) {
		return "<ERR:DHKX:CLI>", "Client Group is NULL", ""
	} //end if
	if(privCli == nil) {
		return "<ERR:DHKX:CLI>", "Client PrivKey is NULL", ""
	} //end if
	if(pubCli == nil) {
		return "<ERR:DHKX:CLI>", "Client PubKey is NULL", ""
	} //end if
	if(recvPubSrv == nil) {
		return "<ERR:DHKX:CLI>", "Received Server PubKey is NULL", ""
	} //end if
	if(shardCli == "") {
		return "<ERR:DHKX:CLI>", "Client SharedKey is Empty", ""
	} //end if
	if(shardExch == "") {
		return "<ERR:DHKX:CLI>", "Client SharedExchange is Empty", ""
	} //end if
	//--
	if(DEBUG == true) {
		log.Println("[DEBUG] DhKx SharedSecret:", shardCli)
	} //end if
	//--
	return "<DHKX:SRV>", smart.BlowfishEncryptCBC(smart.BaseEncode(pubCli, "b58") + ":" + smart.BaseEncode([]byte(shardExch), "b62"), smart.BaseEncode(recvPubSrv, "b92")), shardCli
	//--
} //END FUNCTION


func msgPakHandleMessage(conn *websocket.Conn, isServer bool, id string, remoteId string, msgHash string, maxLimitSeconds uint32, message string, sharedPrivateKey string, sharedSecret string, authUsername string, authPassword string, handleMessagesFunc HandleMessagesFunc) (okRecv bool, okRepl bool, errMsg string, extData string) {
	//--
	defer smart.PanicHandler()
	//--
	var isRecvOk bool = false
	//--
	msg, errMsg := msgPakParseMessage(message, sharedPrivateKey, sharedSecret)
	lenMessage := len(smart.StrTrimWhitespaces(message))
	message = ""
	if(errMsg != "") {
		return isRecvOk, false, errMsg, ""
	} //end if
	isRecvOk = true
	//--
	var area string = "client"
	var rarea string = "server"
	if(isServer == true) {
		area = "server"
		rarea = "client"
	} //end if
	//--
	var identRepl string = "*** MsgPak.Handler." + area
	if(DEBUG == true) {
		identRepl += "{" + id + "}"
	} //end if
	identRepl += " <- " + rarea + "[" + remoteId + "](" + msgHash + "):"
	//--
	log.Println("[INFO] " + identRepl + " Received Command `" + msg.Cmd + "` Data-Size: " + smart.ConvertIntToStr(len(msg.Data)) + " / Package-Size: " + smart.ConvertIntToStr(lenMessage) + " bytes")
	if(DEBUG == true) {
		log.Println("[DATA] " + identRepl + " Command `" + msg.Cmd + "` Data-Size:", len(msg.Data), " / Package-Size:", lenMessage, "bytes ; Data: `" + msg.Data + "`")
	} //end if else
	//--
	var answerMsg string = ""
	var answerData string = ""
	var extraData string = ""
	//--
	var shouldAnswer bool = true
	switch(msg.Cmd) {
		case "<DHKX:CLI>": // client DHKX Key Exchange
			if(isServer != true) {
				answerMsg, answerData, extraData = dhkxCliHandler(remoteId, isServer, msg.Cmd, msg.Data)
			} //end if
			break
		case "<PING>": // ping (zero)
			if(isServer != true) {
				answerMsg = "<OK:PING>"
				answerData = msg.Cmd
			} //end if
			break
		case "<PONG>": // pong (one)
			if(isServer == true) {
				answerMsg = "<OK:PONG>"
				answerData = msg.Cmd
			} //end if
			break
		case "<OK:PING>", "<OK:PONG>":
			if(DEBUG == true) {
				log.Println("[DEBUG] " + identRepl + " # Command `" + msg.Cmd + "` Confirmation for: " + remoteId)
			} //end if
			shouldAnswer = false
			break
		case "<OK>":
			log.Println("[OK] " + identRepl + " # Command `" + msg.Cmd + "` @ `" + msg.Data + "`")
			shouldAnswer = false
			break
		case "<INFO>":
			log.Println("[INFO] " + identRepl + " # Command `" + msg.Cmd + "` @ `" + msg.Data + "`")
			shouldAnswer = false
			break
		case "<ERR>":
			log.Println("[WARNING] " + identRepl + ": " + "Invalid Message ! <ERR> is reserved for internal use ...")
			shouldAnswer = false
			break
		default: // custom handler or unhandled
			if(smart.StrStartsWith(msg.Cmd, "<ERR:")) { // for commands starting with <ERR: just forward them to <INFO>
				log.Println("[WARNING] " + identRepl + ": " + msg.Cmd + " # " + msg.Data)
				shouldAnswer = false
			} else {
				/*
				handleMessagesFunc := func(isServer bool, id string, remoteId string, cmd string, data string, authUsername string, authPassword string) (bool, string, string) {
					//--
					defer smart.PanicHandler()
					//--
					var answerMsg string = ""
					var answerData string = ""
					//--
					switch(cmd) { // see below how to implement commands ...
						default: // unhandled
							answerMsg = "<ERR:UNHANDLED>" // return an error answer
							answerData = "Error description goes here"
					} //end switch
					//--
					// if both answerMsg and answerData are empty will return no answer
					// if answerMsg is empty and answerData is non-empty the answerData will be considered as an error message to display
					// if answerMsg is non-empty will reply back with answerMsg and answerData
					//--
					return answerMsg, answerData
					//--
				} //END FUNCTION
				*/
				answerMsg, answerData = handleMessagesFunc(isServer, id, remoteId, msg.Cmd, msg.Data, authUsername, authPassword)
				if(smart.StrStartsWith(answerMsg, "<ERR:")) {
					log.Println("[WARNING] " + identRepl + ": " + msg.Cmd + " # FAILED: " + answerMsg + " # " + answerData)
				} else if((answerMsg == "") && (answerData != "")) {
					log.Println("[ERROR] " + identRepl + ": " + msg.Cmd + " # FAILED: # " + answerData)
				} else if((answerMsg == "") && (answerData == "")) {
					shouldAnswer = false
				} //end if
			} //end if else
	} //end switch
	//--
	if(shouldAnswer != true) {
		if((answerMsg != "") || (answerData != "")) {
			log.Println("[WARNING] " + identRepl + ": " + msg.Cmd + " # Command is Marked as Should-Not-Answer but have a non-empty Message/Data: `" + answerMsg + "` / `" + answerData + "`")
		} //end if
		return isRecvOk, false, "", "" // there is no other message to be sent
	} //end if
	//--
	if(conn == nil) { // do not return any message in this case ...
		return isRecvOk, false, identRepl + " # Cannot Send Back Reply to `" + msg.Cmd + "` @ No connection available ...", ""
	} //end if
	wrOK, lenPakMsg, errWrMsg := msgPakWriteMessage(conn, maxLimitSeconds, answerMsg, answerData, sharedPrivateKey, sharedSecret)
	if((wrOK != true) || (errWrMsg != "")) {
		if(errWrMsg == "") {
			errWrMsg = "Unknown Error"
		} //end if
		if(DEBUG == true) {
			log.Println("[DEBUG] " + identRepl + " # Message Reply FAILED to [" + rarea + "] @ " + errWrMsg)
		} //end if
		return isRecvOk, true, errWrMsg, ""
	} //end if
	//--
	log.Println("[NOTICE] " + identRepl + " Message Reply to [" + rarea + "] # `" + answerMsg + "` ; Data-Size:", len(answerData), " / Package-Size:", lenPakMsg, "bytes")
	//--
	return isRecvOk, true, "", extraData
	//--
} //END FUNCTION


//-- helpers


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
	var uuid string = uid.Uuid1013Str(13) + "-" + uid.Uuid1013Str(10) + "-" + uid.Uuid1013Str(13)
	if(theTime != "") {
		uuid += "-" + b58.Encode([]byte(theTime))
	} //end if
	//--
	return uuid
	//--
} //END FUNCTION


func msgPakGenerateMessageHash(msg []byte) string {
	//--
	return smart.Crc32b(string(msg))
	//--
} //END FUNCTION


//-- server


func MsgPakSetServerTaskCmd(cmd string, data string, timeoutSec uint32, tlsMode string, certifPath string, httpAddr string, httpPort uint16, authUsername string, authPassword string) string {

	//--
	certifPath = smart.StrTrimWhitespaces(certifPath)
	if((certifPath == "") || (smart.PathIsBackwardUnsafe(certifPath) == true)) {
		certifPath = CERTIFICATES_DEFAULT_PATH
	} //end if
	certifPath = smart.PathGetAbsoluteFromRelative(certifPath)
	certifPath = smart.PathAddDirLastSlash(certifPath)
	//--

	//--
	var uri string = "http"
	var tlsInsecureSkipVerify bool = false
	var tlsServerCerts string = ""
	if(tlsMode == "tls:server") {
		uri += "s"
		uri += "://"
		crt, errCrt := smart.SafePathFileRead(certifPath + CERTIFICATE_PEM_CRT, true)
		if(errCrt != "") {
			return "Failed to read root certificate CRT: " + errCrt
		} //end if
		key, errKey := smart.SafePathFileRead(certifPath + CERTIFICATE_PEM_KEY, true)
		if(errKey != "") {
			return "Failed to read root certificate KEY: " + errKey
		} //end if
		tlsServerCerts = smart.StrTrimWhitespaces(string(crt)) + "\n" + smart.StrTrimWhitespaces(string(key))
	} else if(tlsMode == "tls:noverify") {
		uri += "s"
		uri += "://"
		tlsInsecureSkipVerify = true
	} else if(tlsMode == "tls") {
		uri += "s"
		uri += "://"
	} else { // insecure
		uri += "://"
	} //end if else
	//--
	uri += httpAddr
	uri += ":" + smart.ConvertUInt16ToStr(httpPort)
	uri += "/msgsend"
	//--
	var reqArr map[string][]string = map[string][]string{
		"cmd": { cmd },
		"data": { data },
	}
	//--

	//--
	httpResult := smarthttputils.HttpClientDoRequestPOST(uri, tlsServerCerts, tlsInsecureSkipVerify, nil, reqArr, timeoutSec, smarthttputils.HTTP_CLI_DEF_BODY_READ_SIZE, 0, authUsername, authPassword)
	//--
	if(httpResult.Errors != "") {
		return "SET Error # " + httpResult.Errors
	} else if(httpResult.HttpStatus != 202) {
		return "SET Failed # " + smart.ConvertIntToStr(httpResult.HttpStatus)
	} //end if
	//--

	//--
	return ""
	//--

} //END FUNCTION


func MsgPakServerRun(serverID string, useTLS bool, certifPath string, httpAddr string, httpPort uint16, allowedIPs string, authUsername string, authPassword string, sharedEncPrivKey string, intervalMsgSeconds uint32, handleMessagesFunc HandleMessagesFunc, allowedHttpCustomCmds map[string]bool, cronMsgTasks []CronMsgTask) bool {

	//--

	defer smart.PanicHandler()

	//-- checks

	serverID = smart.StrTrimWhitespaces(serverID)
	if(serverID == "") {
		serverID = MsgPakGenerateUUID()
		log.Println("[NOTICE] MsgPak Server: No Server ID provided, assigning an UUID as ID:", serverID)
	} //end if
	if(serverID == "") {
		log.Println("[ERROR] MsgPak Server: Empty Server ID")
		return false
	} //end if
	if(len(serverID) > 64) {
		log.Println("[ERROR] MsgPak Server: Server ID is too long")
		return false
	} //end if

	certifPath = smart.StrTrimWhitespaces(certifPath)
	if((certifPath == "") || (smart.PathIsBackwardUnsafe(certifPath) == true)) {
		certifPath = CERTIFICATES_DEFAULT_PATH
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

	if(intervalMsgSeconds < LIMIT_INTERVAL_SECONDS_MIN) { // {{{SYNC-MSGPAK-INTERVAL-LIMITS}}}
		log.Println("[ERROR] MsgPak Server: Min allowed Message Interval Seconds is", LIMIT_INTERVAL_SECONDS_MIN, "seconds but is set to:", intervalMsgSeconds)
		return false
	} else if(intervalMsgSeconds > LIMIT_INTERVAL_SECONDS_MAX) { // {{{SYNC-MSGPAK-INTERVAL-LIMITS}}}
		log.Println("[ERROR] MsgPak Server: Max allowed Message Interval Seconds is", LIMIT_INTERVAL_SECONDS_MAX, "seconds but is set to:", intervalMsgSeconds)
		return false
	} //end if

	var allowedHttpCmds sync.Map
	if(allowedHttpCustomCmds != nil) {
		for ks, vs := range allowedHttpCustomCmds {
			if(vs == true) { // if true can be schedduled also via HTTP(S) tasks manager, else only by cron tasks manager ; commands containing ":" cannot be schedduled {{{SYNC-MSGPAK-SPECIAL-COMMANDS}}}
				allowedHttpCmds.Store(ks, vs)
			} //end if
		} //end for
	} //end if

	//-- #

	var srvWebSockUpgrader = websocket.Upgrader{
		ReadBufferSize:    16384,
		WriteBufferSize:   16384,
		EnableCompression: false, // this is still experimental
	} // use default options

	var dhkxSrvKeysClients sync.Map
	var connectedClients sync.Map

	var memSrvMsgCache *smartcache.InMemCache = smartcache.NewCache("smart.websocketmsgpak.server.messages.inMemCache", time.Duration(int(intervalMsgSeconds + 1)) * time.Second, DEBUG_CACHE)

	var srvCustomMsgs map[string][]string = map[string][]string{}
	var mtxSrvCustomMsgs sync.RWMutex // use a RWMutex instead of Mutex ... currently uses no RLock/RUnlock but ... just in case ...

	const defaultMessageCmd = "<PING>"
	var defaultMessageDat = "PING, from the Server: [" + serverID + "]"

	//--

	setNewTask := func(theMsgCmd string, theMsgData string, theArea string) (err string) { // commands containing ":" cannot be schedduled {{{SYNC-MSGPAK-SPECIAL-COMMANDS}}}
		//--
		defer smart.PanicHandler()
		//--
		err = "" // initialize
		//--
		theMsgCmd = smart.StrTrimWhitespaces(smart.StrTrim(smart.StrTrimWhitespaces(theMsgCmd), "<>")) // min 1 char ; max 255 chars ; must contain only a-z A-Z 0-9 - . :
		theMsgData = smart.StrTrimWhitespaces(theMsgData)
		//--
		if((len(theMsgCmd) < 1) || (len(theMsgCmd) > 255) || (theMsgCmd == "") || (!smart.StrRegexMatchString(`^[a-zA-Z0-9\-\.\:]+$`, theMsgCmd))) { // {{{SYNC-MSGPAK-CMD-CHECKS-FORMAT}}}
			err = "Failed to Register new Task Command # Format is Invalid `" + theMsgCmd + "`"
			log.Println("[WARNING] !!!!!!! " + err)
			return
		} //end if
		if(smart.StrContains(theMsgCmd, ":")) { // indirect commands are dissalowed ... (must not contain `:`) // {{{SYNC-MSGPAK-CMD-CHECKS-SPECIALS}}}
			err = "Failed to Register new Task Command # Disallowed `" + theMsgCmd + "`"
			log.Println("[WARNING] !!!!!!! " + err)
			return
		} //end if
		//--
		var lenMsgData int = len(theMsgData)
		if(lenMsgData > int(MAX_MSG_SIZE)) {
			err = "Failed to Register new Task Command # Data is Oversized: " + smart.ConvertIntToStr(lenMsgData) + " bytes"
			log.Println("[WARNING] !!!!!!! " + err)
			return
		} //end if
		//--
		theMsgCmd = smart.StrToUpper(theMsgCmd)
		//--
		_, cmdExst := allowedHttpCmds.Load(theMsgCmd)
		if(!cmdExst) {
			err = "Failed to Register new Task Command # Disallowed `" + theMsgCmd + "`"
			log.Println("[WARNING] !!!!!!! " + err)
			return
		} //end if
		//--
		theMsgCmd = "<" + theMsgCmd + ">"
		//--
		var numConnCli int = 0
		//--
		var errConnCli int = 0
		connectedClients.Range(func(kk, vv interface{}) bool {
			//--
			numConnCli++
			//--
			k := string(fmt.Sprint(kk)) // convert from type interface to string
			if(DEBUG == true) {
				log.Println("[DEBUG] Task Command: Connected Client found # UUID:", k)
			} //end if
			//--
			mtxSrvCustomMsgs.Lock()
			//--
			if(len(srvCustomMsgs[k]) <= int(MAX_QUEUE_MESSAGES)) { // hardcoded
				srvCustomMsgs[k] = append(srvCustomMsgs[k], smart.Base64Encode(theMsgCmd) + "|" + smart.Base64Encode(theMsgData) + "|" + smart.Base64Encode(smart.DateNowIsoUtc()))
				if(DEBUG == true) {
					log.Println("[DEBUG] +++++++ Register Task Command for Client: `" + k + "` in Queue: `" + theMsgCmd + "`")
				} //end if
			} else {
				errConnCli++
				log.Println("[WARNING] !!!!!!! Failed to Register new Task Command for Client: `" + k + "` # Queue is full: `" + theMsgCmd + "`")
			} //end if else
			//--
			mtxSrvCustomMsgs.Unlock()
			//--
			return true
			//--
		})
		if(errConnCli > 0) {
			err = "Failed to Register new Task Command for " + smart.ConvertIntToStr(errConnCli) + " Clients # `" + theMsgCmd + "`"
			return
		} //end if
		//--
		log.Println("[OK] New Task Command was Set by {" + theArea + "} for", numConnCli, "connected client(s): `" + theMsgCmd + "` ; Data-Length:", lenMsgData, "bytes")
		return ""
		//--
	} //END FUNCTION

	//--

	if((cronMsgTasks != nil) && (len(cronMsgTasks) > 0)) { // commands containing ":" cannot be schedduled {{{SYNC-MSGPAK-SPECIAL-COMMANDS}}}
		ctab := crontab.New()
		for t:=0; t<len(cronMsgTasks); t++ {
			log.Println("[INFO] MsgPak Server :: Registering Self-Cron Job Tasks: `" + cronMsgTasks[t].Timing + "` # <" + cronMsgTasks[t].Cmd + "> @ [", len(cronMsgTasks[t].Data), "bytes ]")
			cronJoberr := ctab.AddJob(cronMsgTasks[t].Timing, func(idx int){
				log.Println("[NOTICE] ······· ······· MsgPak Server :: A New Client Task will be set via Self-Cron Job #" + smart.ConvertIntToStr(idx) + " (" + cronMsgTasks[idx].Timing + ") ······· <" + cronMsgTasks[idx].Cmd + ">")
				setNewTask(cronMsgTasks[idx].Cmd, cronMsgTasks[idx].Data, "Self-Cron Job #" + smart.ConvertIntToStr(idx))
			}, t)
			if(cronJoberr != nil) {
				log.Println("[ERROR] MsgPak Server :: Failed to Register a Task as Self-Cron Job #" + smart.ConvertIntToStr(t) + " :", cronJoberr)
				return false
			} //end if
		} //end for
	} //end if

	//--

	srvBroadcastMsg := func(conn *websocket.Conn, rAddr string) {
		//--
		defer smart.PanicHandler()
		//--
		var oneCustomMsg []string = []string{}
		var sendCustomMsgToThisClient bool = false
		var theCacheMsgHash string = ""
		//--
		var crrMessageCmd string = ""
		var crrMessageDat string = ""
		//--
		for {
			//--
			if(conn == nil) {
				break
			} //end if
			//--
			oneCustomMsg = []string{} // init
			theCacheMsgHash = "" // init
			sendCustomMsgToThisClient = false // init
			//--
			//===
			//--
			mtxSrvCustomMsgs.Lock() // use just lock for read and writes
			//--
			log.Println("[DEBUG] ≡≡≡≡≡≡≡ Task Commands Queue Length # Client(s):", len(srvCustomMsgs), "≡≡≡≡≡≡≡")
			if(DEBUG == true) {
				log.Println("[DATA] Message Queue:", srvCustomMsgs)
			} //end if
			//--
			if((srvCustomMsgs[rAddr] != nil) && (len(srvCustomMsgs[rAddr]) > 0)) { // if there are custom (task) messages in the queue, get first
				theCacheMsgHash = smart.Sha512B64(smart.StrTrimWhitespaces(srvCustomMsgs[rAddr][0]))
				oneCustomMsg = smart.ExplodeWithLimit("|", smart.StrTrimWhitespaces(srvCustomMsgs[rAddr][0]), 3) // cmd | data | dtime
				if(len(srvCustomMsgs[rAddr]) > 1) {
					var tmpList []string = srvCustomMsgs[rAddr][1:] // remove 1st element from list after read (key:0)
					srvCustomMsgs[rAddr] = tmpList
					tmpList = nil
				} else {
					srvCustomMsgs[rAddr] = []string{} // there was only one element, reset !
				} //end if else
				if(DEBUG == true) {
					log.Println("[DEBUG] srvBroadcastMsg: Found a Queued Task Command for Client `" + rAddr + "` ; Hash: `" + theCacheMsgHash + "`")
				} //end if
				if(len(oneCustomMsg) == 3) {
					sendCustomMsgToThisClient = true
				} //end if
			} //end if
			//--
			if(srvCustomMsgs[rAddr] != nil) {
				if(len(srvCustomMsgs[rAddr]) <= 0) {
					delete(srvCustomMsgs, rAddr)
					if(DEBUG == true) {
						log.Println("[DEBUG] srvBroadcastMsg: ------- Unregister Client: `" + rAddr + "` from Queue (cleanup, empty list) ...")
					} //end if
				} //end if
			} //end if
			//--
			mtxSrvCustomMsgs.Unlock()
			//--
			//===
			//--
			if(sendCustomMsgToThisClient == true) {
				//--
				if(DEBUG == true) {
					log.Println("[DEBUG] srvBroadcastMsg: Check in Cache for the specific Task Command for Client `" + rAddr + "` ; Hash: `" + theCacheMsgHash + "`")
				} //end if
				cacheExists, cachedObj, _ := memSrvMsgCache.Get(rAddr + "|" + theCacheMsgHash) // {{{SYNC-MSGPAK-CACHE-KEY}}}
				if(DEBUG_CACHE == true) {
					log.Println("[DATA] srvBroadcastMsg: Cached Info for the specific Task Command for Client `" + rAddr + "` ; Hash: `" + theCacheMsgHash + "` ; In-Cache:", cacheExists, "; Object:", cachedObj)
				} //end if
				//--
				if(cacheExists != true) { // send
					cachedObj.Id = rAddr + "|" + theCacheMsgHash // {{{SYNC-MSGPAK-CACHE-KEY}}}
					cachedObj.Data = smart.DateNowIsoUtc()
					memSrvMsgCache.Set(cachedObj, uint64(intervalMsgSeconds * 10)) // support up to 7 ( + 3 free loops) queued messages {{{SYNC-MAX-QUEUED-MSGPAK}}}
					if(DEBUG == true) {
						log.Println("[DEBUG] srvBroadcastMsg: Task Command Cached now (send) for Client `" + rAddr + "` ; Hash: `" + theCacheMsgHash + "`")
					} //end if
				} else { // skip
					sendCustomMsgToThisClient = false
					if(DEBUG == true) {
						log.Println("[DEBUG] srvBroadcastMsg: Task Command already Cached (skip) for Client `" + rAddr + "` ; Hash: `" + theCacheMsgHash + "`")
					} //end if
				} //end if
				//--
			} else {
				//--
				if(theCacheMsgHash != "") {
					log.Println("[ERROR] srvBroadcastMsg: Invalid Task Command for Client `" + rAddr + "` ; Hash: `" + theCacheMsgHash + "`")
				} //end if
				//--
			} //end if
			//--
			if(sendCustomMsgToThisClient == true) {
				crrMessageCmd = smart.Base64Decode(oneCustomMsg[0])
				crrMessageDat = smart.Base64Decode(oneCustomMsg[1])
			} else {
				crrMessageCmd = defaultMessageCmd
				crrMessageDat = defaultMessageDat
			} //end if else
			//--
			sendCustomMsgToThisClient = false // reset
			theCacheMsgHash = "" // reset
			oneCustomMsg = []string{} // reset
			//--
			log.Println("[NOTICE] @@@ Broadcasting " + crrMessageCmd + " Message to Client{" + rAddr + "}, Data-Size:", len(crrMessageDat), "bytes")
			//--
			cliShardIntf, cliShardExst := dhkxSrvKeysClients.Load(rAddr)
			var cliShardStr string = ""
			if(cliShardExst) {
				cliShardStr = string(fmt.Sprint(cliShardIntf)) // convert from type interface to string
			} //end if
			if(smart.StrTrimWhitespaces(cliShardStr) == "") {
				log.Println("[WARNING] @@@ Broadcasting # Client{" + rAddr + "} Shared Key is Empty")
				break
			} //end if
			//--
			msg, errMsg := msgPakComposeMessage(crrMessageCmd, crrMessageDat, sharedEncPrivKey, cliShardStr)
			//--
			if(errMsg != "") {
				//--
				log.Println("[ERROR] @@@ Broadcasting # Send Message to Client{" + rAddr + "}:", errMsg)
				break
				//--
			} else {
				//--
				errWrs := connWriteTxtMsgToSocket(conn, []byte(msg), intervalMsgSeconds)
				//--
				if(errWrs != nil) {
					//--
					log.Println("[ERROR] @@@ Broadcasting # Send Message to Client{" + rAddr + "} / Writing to websocket Failed:", errWrs)
					break
					//--
				} else {
					//--
					log.Println("[OK] @@@ Broadcasting # Send Message completed to Client{" + rAddr + "}")
					//--
				} //end if else
				//--
			} //end if else
			//--
			time.Sleep(time.Duration(intervalMsgSeconds) * time.Second)
			//--
		} //end for
		//--
		return
		//--
	} //end function

	//--

	srvHandlerMsgPack := func(w http.ResponseWriter, r *http.Request) {
		//-- check auth
		var authErr string = smarthttputils.HttpBasicAuthCheck(w, r, HTTP_AUTH_REALM, authUsername, authPassword, allowedIPs, false) // outputs: TEXT
		if(authErr != "") {
			log.Println("[WARNING] MessagePak Server / MsgPak Channel Area :: Authentication Failed:", authErr)
			return
		} //end if
		//-- upgrade the raw HTTP connection to a websocket based one ; below method must check credentials
		srvWebSockUpgrader.CheckOrigin = func(r *http.Request) bool {
			if(authErr != "") {
				return false
			} //end if
			return true
		} // this is for ths js client connected from another origin ...
		//--
		conn, err := srvWebSockUpgrader.Upgrade(w, r, nil)
		//--
		connectedClients.Store(r.RemoteAddr, conn)
		defer func() {
			connectedClients.Delete(r.RemoteAddr)
			connCloseSocket(conn)
		}()
		//--
		if(err != nil) {
			log.Println("[ERROR] Connection Upgrade Failed:", err)
			return
		} //end if
		//--
		log.Println("New Pre-Connection (DhKx Exchange) <-> <-> <-> to:", conn.LocalAddr(), "From:", r.RemoteAddr)
		time.Sleep(time.Duration(2) * time.Second)
		//--
		var serverSendDhKxToClient dhkx.HandleDhkxSrvSendFunc = func(srvPubKey []byte, grpID int) string {
			//--
			msg, errCompose := msgPakComposeMessage("<DHKX:CLI>", smart.ConvertIntToStr(grpID) + ":" + smart.BaseEncode(srvPubKey, "b62"), sharedEncPrivKey, "")
			if(errCompose != "") {
				return "Send (to Client) ERR (1): " + errCompose
			} //end if
			err := connWriteTxtMsgToSocket(conn, []byte(msg), WAIT_DHKX_LIMIT_SECONDS)
			if(err != nil) {
				return "Send (to Client) ERR (2): " + err.Error()
			} //end if
			return ""
			//--
		} //end function
		var groupID int = dhkx.DhKxGetRandomGroup(true) // high only
		errSrvStep1, grpSrv, privSrv, _ := dhkx.DhKxServerInitExchange(groupID, serverSendDhKxToClient)
		if(errSrvStep1 != "") {
			log.Println("[ERROR]: DhKx #1 " + errSrvStep1)
			return
		} //end if
		var serverRecvDhKxFromClient dhkx.HandleDhkxSrvRecvFunc = func(srvPubKey []byte) (string, []byte, []byte) {
			//--
			msgType, message, err := connReadFromSocket(conn, WAIT_DHKX_LIMIT_SECONDS)
			if(err != nil) {
				return "Recv (from Client) ERR: " + err.Error(), nil, nil
			} //end if
			if(msgType != websocket.TextMessage) {
				return "Recv (from Client) ERR: Not a Text Message", nil, nil
			} //end if
			//--
			msg, errMsg := msgPakParseMessage(string(message), sharedEncPrivKey, "")
			if(errMsg != "") {
				return "Recv (from Client) ERR: Invalid Message: " + errMsg, nil, nil
			} //end if
			if(smart.StrStartsWith(msg.Cmd, "<ERR:DHKX:")) {
				return "Recv (from Client) ERR: Message Cmd Failed: `" + msg.Cmd + "` # " + msg.Data, nil, nil
			} //end if
			if(msg.Cmd != "<DHKX:SRV>") {
				return "Recv (from Client) ERR: Invalid Message Cmd: `" + msg.Cmd + "`", nil, nil
			} //end if
			decdata := smart.StrTrimWhitespaces(smart.BlowfishDecryptCBC(msg.Data, smart.BaseEncode(srvPubKey, "b92")))
			if(decdata == "") {
				return "Recv (from Client) ERR: Invalid Message Data Encryption", nil, nil
			} //end if
			data := smart.Explode(":", decdata)
			if(len(data) != 2) {
				return "Recv (from Client) ERR: Invalid Message Data Structure", nil, nil
			} //end if
			var cliPubKey []byte = smart.BaseDecode(data[0], "b58")
			var cliExch []byte = smart.BaseDecode(data[1], "b62")
			//--
			return "", cliPubKey, cliExch
			//--
		} //END FUNCTION
		errSrvRecv1GenShardStep2, recvPubCli, shardSrv := dhkx.DhKxServerFinalizeExchange(grpSrv, privSrv, serverRecvDhKxFromClient)
		if(errSrvRecv1GenShardStep2 != "") {
			log.Println("[ERROR]: DhKx #2 " + errSrvRecv1GenShardStep2)
			return
		} //end if
		if(recvPubCli == nil) {
			log.Println("[ERROR]: DhKx #2 CliPubKey is NULL")
			return
		} //end if
		shardSrv = smart.StrTrimWhitespaces(shardSrv)
		if(shardSrv == "") {
			log.Println("[ERROR]: DhKx #2 SharedSecret is EMPTY")
			return
		} //end if
		if(smart.StrTrimWhitespaces(smart.Base64Decode(shardSrv)) == "") {
			log.Println("[ERROR]: DhKx #2 SharedSecret is INVALID")
			return
		} //end if
		if(DEBUG == true) {
			log.Println("[DEBUG] DhKx SharedSecret:", shardSrv)
		} //end if
		dhkxSrvKeysClients.Store(r.RemoteAddr, shardSrv)
		defer func() {
			dhkxSrvKeysClients.Delete(r.RemoteAddr)
		}()
		log.Println("[OK] <-> <-> <-> DhKx Exchange Completed:", conn.LocalAddr(), "<->", r.RemoteAddr, "/ Key-Length:", len(shardSrv), "bytes")
		time.Sleep(time.Duration(2) * time.Second)
		//--
		log.Println("New Connection to:", conn.LocalAddr(), "From:", r.RemoteAddr)
		//-- The event loop
		go srvBroadcastMsg(conn, r.RemoteAddr)
		var msgHash string = ""
		for {
			//--
			messageType, message, err := connReadFromSocket(conn, intervalMsgSeconds)
			if(err != nil) {
				log.Println("[ERROR] Message Reading Failed (interval", intervalMsgSeconds, "sec.):", err)
				break
			} //end if
			//--
			if(DEBUG == true) {
				log.Println("[DEBUG] Server: [", conn.LocalAddr(), "] # Got New Message from Client: {" + r.RemoteAddr + "} # Type:", messageType)
			} //end if
			//--
			if(messageType == websocket.TextMessage) {
				//--
				msgHash = msgPakGenerateMessageHash(message) // {{{SYNC-MSGPAK-MSGHASH}}}
				//--
				log.Println("[NOTICE] Message Received from Client{" + r.RemoteAddr + "} # Message-Hash: " + msgHash + " ; Package Size:", len(message), "bytes")
				//--
				cliShardIntf, cliShardExst := dhkxSrvKeysClients.Load(r.RemoteAddr)
				var cliShardStr string = ""
				if(cliShardExst) {
					cliShardStr = string(fmt.Sprint(cliShardIntf)) // convert from type interface to string
				} //end if
				if(smart.StrTrimWhitespaces(cliShardStr) == "") {
					log.Println("[WARNING] Client Shared Key is Empty")
					break
				} //end if
				mRecvOk, mRepl, errMsg, _ := msgPakHandleMessage(conn, true, serverID, r.RemoteAddr, msgHash, intervalMsgSeconds, string(message), sharedEncPrivKey, cliShardStr, authUsername, authPassword, handleMessagesFunc)
				message = nil
				if(mRecvOk != true) {
					log.Println("[ERROR] Invalid Message received from Client{" + r.RemoteAddr + "} # Message-Hash: " + msgHash + " ; Details: " + errMsg)
				} else { // recv ok
					log.Println("[OK] Valid Message received from Client{" + r.RemoteAddr + "} # Message-Hash: " + msgHash)
					if(errMsg != "") {
						log.Println("[ERROR] Failed to Reply back to Message from Client{" + r.RemoteAddr + "} # Message-Hash: " + msgHash + " ; Details: " + errMsg)
					} else {
						if(mRepl == true) {
							log.Println("[OK] Reply back to Message from Client{" + r.RemoteAddr + "} # Message-Hash: " + msgHash)
						} //end if else
					} //end if else
				} //end if else
				//--
				msgHash = ""
				//--
			} else {
				//--
				log.Println("[ERROR]: TextMessage is expected from Client{" + r.RemoteAddr + "}")
				//--
			} //end if else
			//--
		} //end for
		//--
		return
		//--
	} //end function

	srvHandlerCustomMsg := func(w http.ResponseWriter, r *http.Request) {
		//--
		if(allowedHttpCustomCmds == nil) {
			smarthttputils.HttpStatus503(w, r, "This service area is NOT Active", true)
			return
		} //end if
		//--
		var authErr string = smarthttputils.HttpBasicAuthCheck(w, r, HTTP_AUTH_REALM, authUsername, authPassword, allowedIPs, true) // outputs: HTML
		if(authErr != "") {
			log.Println("[WARNING] MessagePak Server / Task Commands Area :: Authentication Failed:", authErr)
			return
		} //end if
		//--
		var isRequestOk bool = true
		//--
		if(r.Method == http.MethodGet) { // GET
			r.ParseForm()
			_, hasParamCmd := r.Form["cmd"]
			_, hasParamData  := r.Form["data"]
			if((hasParamCmd != true) && (hasParamData != true)) {
				smarthttputils.HttpStatus200(w, r, srvassets.HtmlServerTemplate("Server: New Task Command", "", `<h1>Server: New Task Command &nbsp; <i class="sfi sfi-tab sfi-2x"></i></h1>` + `<form name="new-task-form" method="post" action="msgsend" class="ux-form">` + "\n" + `<div class="operation_success">` + `<input type="text" name="cmd" class="ux-field" placeholder="Cmd" title="Cmd" maxlength="255" style="width:300px;">` + `</div>` + "\n" + `<div class="operation_important">` + `<textarea name="data" class="ux-field" placeholder="Data" title="Data" maxlength="16000000" style="width:300px; height:200px;"></textarea>` + `</div>` + "\n" + `<button type="submit" disabled style="display:none;" aria-hidden="true" data-hint="Prevent Form Submit on Enter"></button>` + "\n" + `<button type="submit" class="ux-button ux-button-special"><i class="sfi sfi-new-tab"></i>&nbsp;&nbsp; Submit Task Command</button>` + "\n" + `</form>`), "index.html", "", -1, "", "no-cache", nil)
				return
			} //end if
		} else if(r.Method == http.MethodPost) { // POST
			r.ParseForm()
			r.ParseMultipartForm(int64(smarthttputils.HTTP_CLI_MAX_POST_DATA_SIZE))
		} else {
			isRequestOk = false
		} //end if else
		//--
		var customcmd string = r.FormValue("cmd")
		var customdata string = r.FormValue("data")
		if(DEBUG == true) {
			log.Println("[DEBUG] RequestVars:", "cmd", customcmd, ";", "data", customdata)
		} //end if
		//--
		if(isRequestOk == true) {
			customcmd = smart.StrToUpper(smart.StrTrimWhitespaces(smart.StrTrim(smart.StrTrimWhitespaces(customcmd), "<>")))
			if(customcmd == "") {
				isRequestOk = false
			} //end if
		} //end if
		//--
		var errSetTask string = ""
		//--
		if(isRequestOk == true) {
			errSetTask = setNewTask(customcmd, customdata, "HTTP(S) Task Command (" + r.RemoteAddr + ")")
			if(errSetTask != "") {
				isRequestOk = false
			} //end if
		} //end if
		//--
		if(isRequestOk != true) {
			if(errSetTask == "") {
				errSetTask = "Command is Invalid Or Disallowed"
			} //end if
			smarthttputils.HttpStatus400(w, r, "Invalid Request # Required Variables: [ `cmd` : string, `data` : string ] # " + errSetTask, true)
			return
		} //end if
		//--
		log.Println("[NOTICE] °°°°°°° °°°°°°° A New Task was set via HTTP(S) Task Command °°°°°°° by `" + authUsername + "` from IP Address [`" + r.RemoteAddr + "`] :: `<" + customcmd + ">`")
		//--
		smarthttputils.HttpStatus202(w, r, srvassets.HtmlServerTemplate("Server: Task Command was Set", "", `<h1>Server: Task Command was Set &nbsp; <i class="sfi sfi-tab sfi-2x"></i></h1>` + `<div class="operation_success" title="Command">` + smart.EscapeHtml("<" + customcmd + ">") + `</div>` + "\n" + `<div class="operation_important" title="Data">` + "\n" + `<textarea class="ux-field" style="width:300px; height:200px;" readonly>` + smart.EscapeHtml(customdata) + `</textarea>` + "\n" + `</div>` + "\n" + `<a href="msgsend" class="ux-button">New Task Command</a>`), "index.html", "", -1, "", "no-cache", nil)
		//--
	} //end function

	srvHandlerHome := func(w http.ResponseWriter, r *http.Request) {
		//--
		if(r.URL.Path != "/") {
			smarthttputils.HttpStatus404(w, r, "MsgPack Server: Resource Not Found: `" + r.URL.Path + "`", true)
			return
		} //end if
		//--
		headers := map[string]string{"refresh":"10"}
		smarthttputils.HttpStatus200(w, r, assets.HtmlStandaloneTemplate("MsgPak Server: HTTP(S)/WsMux", "", `<div class="operation_display">MsgPak Server: HTTP(S)/WsMux # ` + smart.EscapeHtml(VERSION) + `</div>` + `<div class="operation_info"><img width="48" height="48" src="lib/framework/img/loading-spin.svg"></div>` + `<hr>` + `<small>` + smart.EscapeHtml(smart.COPYRIGHT) + `</small>`), "index.html", "", -1, "", "no-cache", headers)
		//--
	} //end function

	webAssetsHttpHandler := func(w http.ResponseWriter, r *http.Request) {
		//--
		srvassets.WebAssetsHttpHandler(w, r, "", "cache:private") // use default mime disposition ; private cache mode
		//--
	} //end function

	var srvAddr string = httpAddr + fmt.Sprintf(":%d", httpPort)
	mux, srv := smarthttputils.HttpMuxServer(srvAddr, intervalMsgSeconds, true, "[MsgPak Server]") // force HTTP/1

	mux.HandleFunc("/msgpak", srvHandlerMsgPack)
	mux.HandleFunc("/msgsend", srvHandlerCustomMsg)
	mux.HandleFunc("/lib/", webAssetsHttpHandler)
	mux.HandleFunc("/", srvHandlerHome)

	//--

	if(useTLS == true) {
		log.Println("Starting MsgPak Server:", "wss://" + srvAddr + "/msgpak", "@", "HTTPS/WsMux/TLS", "#", VERSION)
		log.Println("[NOTICE] MsgPak Server Certificates Path:", certifPath)
	//	go log.Fatal("[ERROR]", http.ListenAndServeTLS(srvAddr, certifPath + CERTIFICATE_PEM_CRT, certifPath + CERTIFICATE_PEM_KEY, nil))
		go log.Fatal("[ERROR]", srv.ListenAndServeTLS(certifPath + CERTIFICATE_PEM_CRT, certifPath + CERTIFICATE_PEM_KEY))
	} else {
		log.Println("Starting MsgPak Server:", "ws://" + srvAddr + "/msgpak", "@", "HTTP/WsMux/Insecure", "#", VERSION)
	//	go log.Fatal("[ERROR]", http.ListenAndServe(srvAddr, nil))
		go log.Fatal("[ERROR]", srv.ListenAndServe())
	} //end if else

	//--

	return true

} //END FUNCTION


//-- client

var cliCustomMsgs *fifolist.List = fifolist.New()
var mtxCliCustomMsgs sync.Mutex


func MsgPakSetClientTaskCmd(cmd string, data string) string {
	//--
	cmd = smart.StrTrimWhitespaces(smart.StrTrim(smart.StrTrimWhitespaces(cmd), "<>"))
	data = smart.StrTrimWhitespaces(data)
	//--
	if((len(cmd) < 1) || (len(cmd) > 255) || (cmd == "") || (!smart.StrRegexMatchString(`^[a-zA-Z0-9\-\.\:]+$`, cmd))) { // {{{SYNC-MSGPAK-CMD-CHECKS-FORMAT}}}
		return "Format is Invalid `" + cmd + "`"
	} //end if
//	if(smart.StrContains(cmd, ":")) { // indirect commands are dissalowed ... (must not contain `:`) // {{{SYNC-MSGPAK-CMD-CHECKS-SPECIALS}}}
//		return "Disallowed `" + theMsgCmd + "`" // on client side, this must be allowed
//	} //end if
	//--
	var lenData int = len(data)
	if(lenData > int(MAX_MSG_SIZE)) {
		return "Data is Oversized: " + smart.ConvertIntToStr(lenData) + " bytes"
	} //end if
	//--
	cmd = "<" + smart.StrToUpper(cmd) + ">"
	//--
	mtxCliCustomMsgs.Lock()
	if(cliCustomMsgs.Len() <= int(MAX_QUEUE_MESSAGES)) {
		cliCustomMsgs.PushBack(smart.Base64Encode(cmd) + "|" + smart.Base64Encode(data) + "|" + smart.Base64Encode(smart.DateNowIsoUtc())) // add at the end
	} else {
		log.Println("[WARNING] !!!!!!! Failed to Register new Task Command # Queue is full: `" + cmd + "`")
	} //end if else
	mtxCliCustomMsgs.Unlock()
	//--
	return ""
	//--
} //END FUNCTION


func MsgPakClientRun(clientID string, serverPool []string, tlsMode string, certifPath string, authUsername string, authPassword string, sharedEncPrivKey string, intervalMsgSeconds uint32, handleMessagesFunc HandleMessagesFunc) bool {

	//--

	defer smart.PanicHandler()

	//--

	if(serverPool == nil) {
		serverPool = []string{}
	} //end if

	clientID = smart.StrTrimWhitespaces(clientID)
	if(clientID == "") {
		clientID = MsgPakGenerateUUID()
		log.Println("[NOTICE] MsgPak Server: No Client ID provided, assigning an UUID as ID:", clientID)
	} //end if
	if(clientID == "") {
		log.Println("[ERROR] MsgPak Client: Empty Client ID")
		return false
	} //end if
	if(len(clientID) > 64) {
		log.Println("[ERROR] MsgPak Client: Client ID is too long")
		return false
	} //end if

	certifPath = smart.StrTrimWhitespaces(certifPath)
	if((certifPath == "") || (smart.PathIsBackwardUnsafe(certifPath) == true)) {
		certifPath = CERTIFICATES_DEFAULT_PATH
	} //end if
	certifPath = smart.PathGetAbsoluteFromRelative(certifPath)
	certifPath = smart.PathAddDirLastSlash(certifPath)

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

	if(intervalMsgSeconds < 10) { // {{{SYNC-MSGPAK-INTERVAL-LIMITS}}}
		log.Println("[ERROR] MsgPak Client: Min allowed Message Interval Seconds is", LIMIT_INTERVAL_SECONDS_MIN, "seconds but is set to:", intervalMsgSeconds)
		return false
	} else if(intervalMsgSeconds > 3600) { // {{{SYNC-MSGPAK-INTERVAL-LIMITS}}}
		log.Println("[ERROR] MsgPak Client: Max allowed Message Interval Seconds is", LIMIT_INTERVAL_SECONDS_MAX, "seconds but is set to:", intervalMsgSeconds)
		return false
	} //end if

	//--

	var done chan interface{}
	var interrupt chan os.Signal

	var connectedServers sync.Map
	var dhkxCliKeysServers sync.Map

	var errHandlerReceive uint8 = 0

	receiveHandler := func(conn *websocket.Conn, theServerAddr string) {
		//--
		defer smart.PanicHandler()
		//--
		if(conn == nil) {
			log.Println("[ERROR] receiveHandler Failed:", "No Connection ...")
			// DO NOT INCREMENT errHandlerReceive here ; it is not a RECV Error, it is NO_CONNECTION and is handled by the WATCHDOG
			return
		} //end if
		//--
		defer close(done)
		//--
		var msgHash string = ""
		var firstMessageCompleted bool = false
		for {
			//--
			messageType, message, err := connReadFromSocket(conn, intervalMsgSeconds)
			if(err != nil) {
				log.Println("[ERROR] Message Receive Failed (interval", intervalMsgSeconds, "sec.):", err)
				errHandlerReceive++
				return
			} //end if
			//--
			if(DEBUG == true) {
				log.Println("[DEBUG] Client # Got New Message from Server:{", theServerAddr + "} # Type:", messageType)
			} //end if
			//--
			if(messageType == websocket.TextMessage) {
				//--
				msgHash = msgPakGenerateMessageHash(message) // {{{SYNC-MSGPAK-MSGHASH}}}
				//--
				log.Println("[NOTICE] Message Received from Server{" + theServerAddr + "} # Message-Hash: " + msgHash + " ; Package Size:", len(message), "bytes")
				//--
				srvShardIntf, srvShardExst := dhkxCliKeysServers.Load(theServerAddr)
				var srvShardStr string = ""
				if(srvShardExst) {
					srvShardStr = string(fmt.Sprint(srvShardIntf)) // convert from type interface to string
				} //end if
				if(firstMessageCompleted == true) {
					if(smart.StrTrimWhitespaces(srvShardStr) == "") {
						log.Println("[WARNING] Server{" + theServerAddr + "} Shared Key is Empty ...")
						errHandlerReceive++
						return
					} //end if
				} //end if
				//--
				mRecvOk, mRepl, errMsg, cliShared := msgPakHandleMessage(conn, false, clientID, theServerAddr, msgHash, intervalMsgSeconds, string(message), sharedEncPrivKey, srvShardStr, authUsername, authPassword, handleMessagesFunc)
				if(firstMessageCompleted != true) {
					if(cliShared != "") {
						dhkxCliKeysServers.Store(theServerAddr, cliShared)
						log.Println("[OK] <-> <-> <-> DhKx Exchange Completed:", clientID, "<->", theServerAddr, "/ Key-Length:", len(cliShared), "bytes")
					} //end if
				} //end if
				message = nil
				if(mRecvOk != true) {
					log.Println("[ERROR] Invalid Message received from Server{" + theServerAddr + "} # Message-Hash: " + msgHash + " ; Details: " + errMsg)
				} else { // recv ok
					log.Println("[OK] Valid Message received from Server{" + theServerAddr + "} # Message-Hash: " + msgHash)
					if(errMsg != "") {
						log.Println("[ERROR] Failed to Reply back to Message from Server{" + theServerAddr + "} # Message-Hash: " + msgHash + " ; Details: " + errMsg)
					} else {
						if(mRepl == true) {
							log.Println("[OK] Reply back to Message from Server{" + theServerAddr + "} # Message-Hash: " + msgHash)
						} //end if else
					} //end if else
				} //end if else
				//--
				msgHash = ""
				//--
			} else {
				//--
				log.Println("[ERROR]: TextMessage is expected from Server{" + theServerAddr + "}")
				//--
			} //end if
			//--
			if(firstMessageCompleted != true) {
				firstMessageCompleted = true
			} //end if
			//--
		} //end for
		//--
	} //end function

	connectToServer := func(addr string) {
		//--
		defer smart.PanicHandler()
		//--
		dhkxCliKeysServers.Delete(addr)
		log.Println("[NOTICE] Connecting to Server:", addr, "MODE:", tlsMode)
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
		var httpPort int64 = smart.ParseStrAsInt64(smart.StrTrimWhitespaces(arrAddr[1]))
		if((!smart.IsNetValidIpAddr(httpAddr)) && (!smart.IsNetValidHostName(httpAddr))) {
			log.Println("[ERROR] Invalid Server Address (Host):", addr)
			return
		} //end if
		if(!smart.IsNetValidPortNum(httpPort)) {
			log.Println("[ERROR] Invalid Server Address (Port):", addr)
			return
		} //end if
		//--
		if((tlsMode == "tls") || (tlsMode == "tls:noverify")) {
			log.Println("[NOTICE] Certificates Path:", certifPath)
		} //end if
		//--
		socketPrefix := "ws://"
		socketSuffix := "/msgpak"
		var theWebSocket websocket.Dialer
		if(tlsMode == "tls:server") {
			socketPrefix = "wss://"
			crt, errCrt := smart.SafePathFileRead(certifPath + CERTIFICATE_PEM_CRT, true)
			if(errCrt != "") {
				log.Fatal("[ERROR] Failed to read root certificate CRT: " + errCrt)
			} //end if
			key, errKey := smart.SafePathFileRead(certifPath + CERTIFICATE_PEM_KEY, true)
			if(errKey != "") {
				log.Fatal("[ERROR] Failed to read root certificate KEY: " + errKey)
			} //end if
			log.Println("Initializing Client:", socketPrefix + addr + socketSuffix, "@", "HTTPS/WsMux/TLS:WithServerCertificate")
			log.Println("[NOTICE] Server Certificates Path:", certifPath)
			theWebSocket = websocket.Dialer{
				HandshakeTimeout: time.Duration(HANDSHAKE_TIMEOUT_SECONDS) * time.Second,
				TLSClientConfig: smarthttputils.TlsConfigClient(false, smart.StrTrimWhitespaces(string(crt)) + "\n" + smart.StrTrimWhitespaces(string(key))),
			}
		} else if(tlsMode == "tls:noverify") {
			socketPrefix = "wss://"
			log.Println("Initializing Client:", socketPrefix + addr + socketSuffix, "@", "HTTPS/WsMux/TLS:InsecureSkipVerify")
			theWebSocket = websocket.Dialer{
				HandshakeTimeout: time.Duration(HANDSHAKE_TIMEOUT_SECONDS) * time.Second,
				TLSClientConfig: smarthttputils.TlsConfigClient(true, ""),
			}
		} else if(tlsMode == "tls") {
			socketPrefix = "wss://"
			log.Println("Initializing Client:", socketPrefix + addr + socketSuffix, "@", "HTTPS/WsMux/TLS")
			theWebSocket = websocket.Dialer{
				HandshakeTimeout: time.Duration(HANDSHAKE_TIMEOUT_SECONDS) * time.Second,
				TLSClientConfig: smarthttputils.TlsConfigClient(false, ""),
			}
		} else { // insecure
			log.Println("Initializing Client:", socketPrefix + addr + socketSuffix, "@", "HTTP/WsMux/Insecure")
			theWebSocket = websocket.Dialer{
				HandshakeTimeout: time.Duration(HANDSHAKE_TIMEOUT_SECONDS) * time.Second,
			}
		} //end if else
		h := smarthttputils.HttpClientAuthBasicHeader(authUsername, authPassword)
	//	h = nil
		//--
		conn, response, err := theWebSocket.Dial(socketPrefix + addr + socketSuffix, h)
	//	conn, response, err := websocket.DefaultDialer.Dial(socketPrefix + addr + socketSuffix, h)
		//--
		connectedServers.Store(addr, conn)
		defer func() {
			dhkxCliKeysServers.Delete(addr)
			connectedServers.Delete(addr)
			connCloseSocket(conn)
		}()
		//--
		if(err != nil) {
			var rStatusCode int = 0
			if(response != nil) {
				rStatusCode = response.StatusCode
			} //end if
			log.Println("[ERROR] Cannot connect to Websocket Server: HTTP Response StatusCode:", rStatusCode, "; Dial Errors:", err)
			return
		} //end if
		//--
		go receiveHandler(conn, addr)
		//-- the main loop for the client
		var firstMessageCompleted bool = false
		const defCliCmd  string = "<PONG>"
		var defCliData string = "PONG, from Client: `" + clientID + "`"
		var crrCliCmd  string = ""
		var crrCliData string = ""
		for {
			//--
			if(errHandlerReceive >= MAX_NUM_ERR_MSG_RECV) {
				log.Println("[INFO] Reset Connection to Server, Too many Message RECV Errors:", errHandlerReceive, "of Max Limit:", MAX_NUM_ERR_MSG_RECV)
				errHandlerReceive = 0 // reset !
				return // stop after any error from receive handler in order to force re-connect
			} //end if
			//--
			srvShardIntf, srvShardExst := dhkxCliKeysServers.Load(addr)
			var srvShardStr string = ""
			if(srvShardExst) {
				srvShardStr = string(fmt.Sprint(srvShardIntf)) // convert from type interface to string
			} //end if
			//--
			select {
				case <-time.After(time.Duration(intervalMsgSeconds) * time.Second):
					//--
					if(smart.StrTrimWhitespaces(srvShardStr) == "") {
						//--
						if(firstMessageCompleted == true) {
							log.Println("[WARNING] SKIP: Sending Message to Server{" + addr + "}, Server Shared Key is Empty ...")
						} //end if
						//--
					} else {
						//--
						mtxCliCustomMsgs.Lock()
						//--
						log.Println("[DEBUG] ≡≡≡≡≡≡≡ Task Commands Queue Length:", cliCustomMsgs.Len(), "≡≡≡≡≡≡≡")
						if(cliCustomMsgs.Len() > 0) {
							tmpMsg := cliCustomMsgs.Front() // get 1st element
							tmpValMsg := string(fmt.Sprint(tmpMsg.Value)) // convert from type interface to string
							tmpArrMsg := smart.ExplodeWithLimit("|", smart.StrTrimWhitespaces(tmpValMsg), 3) // cmd | data | dtime
							if(len(tmpArrMsg) == 3) {
								crrCliCmd = smart.Base64Decode(tmpArrMsg[0])
								crrCliData = smart.Base64Decode(tmpArrMsg[1])
							} else {
								log.Println("[ERROR] Malformed Custom Registered Task Command")
								crrCliCmd = defCliCmd
								crrCliData = defCliData
							} //end if else
							cliCustomMsgs.Remove(tmpMsg)
						} else {
							crrCliCmd = defCliCmd
							crrCliData = defCliData
						} //end if else
						//--
						mtxCliCustomMsgs.Unlock()
						//--
						log.Println("[NOTICE] @@@ Sending `" + crrCliCmd + "` Message to Server {" + addr + "}")
						msg, errMsg := msgPakComposeMessage(crrCliCmd, crrCliData, sharedEncPrivKey, srvShardStr)
						if(errMsg != "") {
							log.Println("[ERROR]:", errMsg)
							return
						} else {
							err := connWriteTxtMsgToSocket(conn, []byte(msg), intervalMsgSeconds)
							if(err != nil) {
								log.Println("[ERROR] Writing to websocket Failed:", err)
								return
							} //end if
						} //end if else
						//--
						msg = ""
						errMsg = ""
						//--
					} //end if else
					//--
					if(firstMessageCompleted != true) {
						firstMessageCompleted = true
					} //end if
					//--
				case <-interrupt: // received a SIGINT (Ctrl + C). Terminate gracefully...
					//--
					log.Println("[NOTICE] Received SIGINT interrupt signal. Closing all pending connections")
					err := connWriteCloseMsgToSocket(conn, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")) // close websocket connection
					if(err != nil) {
						log.Println("[ERROR] Writing the Close Message to websocket Failed:", err)
					} //end if
					//-- possible fix
				//	return
					// fix: if crashes comment below and uncomment the return above
					select {
						case <-done:
							log.Println("[NOTICE] Receiver Channel Closed...")
						case <-time.After(time.Duration(1) * time.Second):
							log.Println("[WARNING] Timeout in closing receiving channel...")
					} //end select
					//--
					return
					//-- #end fix
			} //end select
			//--
		} //end for
		//--
	} //end function

	connectWatchdog := func() {
		//--
		log.Println("Starting WS Client", "#", VERSION)
		//--
		var initConn bool = false
		//--
		for {
			//--
			log.Println("======= Connection WATCHDOG ======= is up and running for Client{" + clientID + "} ...")
			if(DEBUG == true) {
				log.Println("[DEBUG] Connected Servers:", connectedServers)
			} //end if
			//--
			for _, p := range serverPool {
				if _, exist := connectedServers.Load(p); exist {
					log.Println("[INFO] Client Connection appears REGISTERED with Server:", p)
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
			time.Sleep(time.Duration(int(intervalMsgSeconds + WAIT_CLOSE_LIMIT_SECONDS + WAIT_CLOSE_LIMIT_SECONDS)) * time.Second)
			//--
		} //end for
		//--
	} //end function

	done = make(chan interface{}) // Channel to indicate that the receiveHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully
	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	go connectWatchdog()

	return true

} //END FUNCTION


// #END
