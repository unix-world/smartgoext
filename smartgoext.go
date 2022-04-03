
// GO Lang :: SmartGo Extra :: Smart.Go.Framework
// (c) 2020-2022 unix-world.org
// r.20220403.2025 :: STABLE

package smartgoext

// REQUIRE: go 1.15 or later

import (
	smart 		"github.com/unix-world/smartgo"
	wdav  		"github.com/unix-world/smartgoext/webdavsrv"
	msgpaksrv 	"github.com/unix-world/smartgoext/websocketmsgpak"
	jsvm  		"github.com/unix-world/smartgoext/quickjsvm"
)


//-----

func WebdavServerRun(authUser string, authPass string, httpAddr string, httpPort uint16, httpsPort uint16, serveSecure bool, disableUnsecure bool, certifPath string, storagePath string) bool {
	//--
	isRunning := wdav.WebdavServerListenAndServe(authUser, authPass, httpAddr, httpPort, httpsPort, serveSecure, disableUnsecure, certifPath, storagePath)
	//--
	return isRunning
	//--
} //END FUNCTION

//-----

func MsgPakServerRun(serverID string, useTLS bool, certifPath string, httpAddr string, httpPort uint16, authUsername string, authPassword string, sharedEncPrivKey string, intervalMsgSeconds uint32, handleMessagesFunc msgpaksrv.HandleMessagesFunc) bool {
	//--
	serverID = smart.StrTrimWhitespaces(serverID)
	if(serverID == "") {
		serverID = msgpaksrv.MsgPakGenerateUUID()
	} //end if
	//--
	isRunning := msgpaksrv.MsgPakServerListenAndServe(serverID, useTLS, certifPath, httpAddr, httpPort, authUsername, authPassword, sharedEncPrivKey, intervalMsgSeconds, handleMessagesFunc)
	//--
	return isRunning
	//--
} //END FUNCTION

func MsgPakClientRun(serverPool []string, clientID string, tlsMode string, authUsername string, authPassword string, sharedEncPrivKey string, intervalMsgSeconds uint32, intervalReconnectSeconds uint32, handleMessagesFunc msgpaksrv.HandleMessagesFunc) bool {
	//--
	clientID = smart.StrTrimWhitespaces(clientID)
	if(clientID == "") {
		clientID = msgpaksrv.MsgPakGenerateUUID()
	} //end if
	//--
	isRunning := msgpaksrv.MsgPakClientListenAndConnectToServer(serverPool, clientID, tlsMode, authUsername, authPassword, sharedEncPrivKey, intervalMsgSeconds, intervalReconnectSeconds, handleMessagesFunc)
	//--
	return isRunning
	//--
} //END FUNCTION

//-----

func QuickJsRunCode(jsCode string, stopTimeout uint, jsMemMB uint16, jsInputData map[string]string, jsExtendMethods map[string]interface{}, jsBinaryCodePreload map[string][]byte) (jsEvErr string, jsEvRes string) {
	//--
	err, jsRes := jsvm.QuickJsVmRunCode(jsCode, stopTimeout, jsMemMB, jsInputData, jsExtendMethods, jsBinaryCodePreload)
	//--
	return err, jsRes
	//--
} //END FUNCTION

//-----


// #END
