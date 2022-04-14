
// GO Lang :: SmartGo Extra / WebDAV Server :: Smart.Go.Framework
// (c) 2020-2022 unix-world.org
// r.20220415.0128 :: STABLE

// Req: go 1.16 or later (embed.FS is N/A on Go 1.15 or lower)
package webdavsrv

import (
	"log"

	"fmt"
	"strconv"
	"bytes"

	"net/http"
	"golang.org/x/net/webdav"

	smart "github.com/unix-world/smartgo"
	assets "github.com/unix-world/smartgo/web-assets"
	smarthttputils 	"github.com/unix-world/smartgo/web-httputils"
)

const (
	THE_VERSION  string = "r.20220415.0128"

	CONN_HOST    string = "127.0.0.1"
	CONN_PORT    uint16 = 13087

	STORAGE_DIR  string = "./wdav"
	DAV_PATH     string = "/webdav"

	TLS_PATH     string = "./ssl"
	TLS_CERT     string = "cert.crt"
	TLS_KEY      string = "cert.key"
)


func WebdavServerRun(allowedIPs string, authUser string, authPass string, httpAddr string, httpPort uint16, timeoutSeconds uint32, serveSecure bool, certifPath string, storagePath string) bool {

	//-- auth user / pass

	authUser = smart.StrTrimWhitespaces(authUser)
	if(authUser == "") {
		log.Println("[ERROR] WebDAV Server: Empty Auth UserName")
		return false
	} //end if
	if((len(authUser) < 5) || (len(authUser) > 25)) { // {{{SYNC-GO-SMART-AUTH-USER-LEN}}}
		log.Println("[ERROR] WebDAV Server: Invalid Auth UserName Length: must be between 5 and 25 characters")
		return false
	} //end if

	// do not trim authPass !
	if(smart.StrTrimWhitespaces(authPass) == "") {
		log.Println("[ERROR] WebDAV Server: Empty Auth Password")
		return false
	} //end if
	if((len(smart.StrTrimWhitespaces(authPass)) < 7) || (len(authPass) > 30)) { // {{{SYNC-GO-SMART-AUTH-PASS-LEN}}}
		log.Println("[ERROR] WebDAV Server: Invalid Auth UserName Length: must be between 7 and 30 characters")
		return false
	} //end if

	//-- http(s) address and port(s)

	httpAddr = smart.StrTrimWhitespaces(httpAddr)
	if((!smart.IsNetValidIpAddr(httpAddr)) && (!smart.IsNetValidHostName(httpAddr))) {
		httpAddr = CONN_HOST
	} //end if

	if(!smart.IsNetValidPortNum(int64(httpPort))) {
		httpPort = CONN_PORT
	} //end if

	//-- paths

	if(serveSecure == true) {
		certifPath = smart.StrTrimWhitespaces(certifPath)
		if((certifPath == "") || (smart.PathIsBackwardUnsafe(certifPath) == true)) {
			certifPath = TLS_PATH
		} //end if
		certifPath = smart.PathGetAbsoluteFromRelative(certifPath)
		certifPath = smart.PathAddDirLastSlash(certifPath)
	} else {
		certifPath = TLS_PATH
	} //end if

	storagePath = smart.StrTrimWhitespaces(storagePath)
	if((storagePath == "") || (smart.PathIsBackwardUnsafe(storagePath) == true)) {
		storagePath = STORAGE_DIR
	} //end if
	storagePath = smart.PathGetAbsoluteFromRelative(storagePath)
	storagePath = smart.PathAddDirLastSlash(storagePath)

	//-- for web

	var serverSignature bytes.Buffer
	serverSignature.WriteString("GO WebDAV Server " + THE_VERSION + "\n")
	serverSignature.WriteString("(c) 2020-2022 unix-world.org" + "\n")
	serverSignature.WriteString("\n")

	if(serveSecure == true) {
		serverSignature.WriteString("<Secure URL> :: https://" + httpAddr + ":" + smart.ConvertUInt16ToStr(httpPort) + DAV_PATH + "/" + "\n")
	} else {
		serverSignature.WriteString("<URL> :: http://" + httpAddr + ":" + smart.ConvertUInt16ToStr(httpPort) + DAV_PATH + "/" + "\n")
	} //end if

	//-- for console

	fmt.Println("===========================================================================")
	fmt.Println("GO WebDAV Server " + THE_VERSION)
	fmt.Println("---------------------------------------------------------------------------")
//	fmt.Println("Current Path: " + string(path))
	fmt.Println("Certificates Path: " + certifPath)
	fmt.Println("DAV Folder: " + storagePath)
	fmt.Println("WebDAV Path: " + DAV_PATH)
	fmt.Println("---------------------------------------------------------------------------")
	if(serveSecure == true) {
		fmt.Println("Secure Listening TLS at https://" + httpAddr + ":" + smart.ConvertUInt16ToStr(httpPort) + "/")
	} else {
		fmt.Println("Unsecure Listening at http://" + httpAddr + ":" + smart.ConvertUInt16ToStr(httpPort) + "/")
	} //end if
	fmt.Println("===========================================================================")

	//-- server

	mux, srv := smarthttputils.HttpMuxServer(httpAddr + fmt.Sprintf(":%d", httpPort), timeoutSeconds, true) // force HTTP/1

	//-- webdav handler

	wdav := &webdav.Handler{
		Prefix:     DAV_PATH,
		FileSystem: webdav.Dir(STORAGE_DIR),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if(err != nil) {
				log.Printf("[WARNING] GO WebDAV Server :: WEBDAV.ERROR: %s [%s %s %s] %s [%s] %s\n", err, r.Method, r.URL, r.Proto, "*", r.Host, r.RemoteAddr)
			} else {
				log.Printf("[OK] GO WebDAV Server :: WEBDAV [%s %s %s] %s [%s] %s\n", r.Method, r.URL, r.Proto, "*", r.Host, r.RemoteAddr)
			} //end if else
		},
	}

	//-- other handlers

	// http root handler : 202 | 404
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if(r.URL.Path != "/") {
			smarthttputils.HttpStatus404(w, r, "WebDAV Resource Not Found: `" + r.URL.Path + "`", false)
			return
		} //end if
		var titleText string = "GO WebDAV Server " + THE_VERSION
		var headHtml string = "<style>" + "\n" + "div.status { text-align:center; margin:10px; cursor:help; }" + "\n" + "div.signature { background:#778899; color:#FFFFFF; font-size:2rem; font-weight:bold; text-align:center; border-radius:3px; padding:10px; margin:20px; }" + "\n" + "</style>"
		var bodyHtml string = `<div class="status"><img alt="Status: Up and Running ..." title="Status: Up and Running ..." width="64" height="64" src="data:image/svg+xml,` + smart.EscapeHtml(smart.EscapeUrl(assets.ReadWebAsset("lib/framework/img/loading-spin.svg"))) + `"></div>` + "\n" + `<div class="signature">` + "\n" + "<pre>" + "\n" + smart.EscapeHtml(serverSignature.String()) + "</pre>" + "\n" + "</div>"
		log.Printf("[OK] GO WebDAV Server :: DEFAULT [%s %s %s] %s [%s] %s\n", r.Method, r.URL, r.Proto, strconv.Itoa(202), r.Host, r.RemoteAddr)
		smarthttputils.HttpStatus202(w, r, assets.HtmlStandaloneTemplate(titleText, headHtml, bodyHtml), "index.html", "", -1, "", "no-cache", nil)
	})

	// http version handler : 203
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[OK] GO WebDAV Server :: VERSION [%s %s %s] %s [%s] %s\n", r.Method, r.URL, r.Proto, strconv.Itoa(203), r.Host, r.RemoteAddr)
		smarthttputils.HttpStatus203(w, r, "GO WebDAV Server " + THE_VERSION + "\n", "version.txt", "", -1, "", "no-cache", nil)
	})

	// webdav handler : all webdav status codes ...
	mux.HandleFunc(DAV_PATH + "/", func(w http.ResponseWriter, r *http.Request) {
		var authErr string = smarthttputils.HttpBasicAuthCheck(w, r, "WebDAV Server: Storage Area", authUser, authPass, allowedIPs, false) // outputs: TEXT
		if(authErr != "") {
			log.Println("[WARNING] WebDAV Server / Storage Area :: Authentication Failed:", authErr)
			return
		} //end if
		wdav.ServeHTTP(w, r) // if all ok above (basic auth + credentials ok, serve ...)
	})

	// serve logic

	if(serveSecure == true) { // serve HTTPS
		var theTlsCertPath string = certifPath + TLS_CERT
		var theTlsKeyPath  string = certifPath + TLS_KEY
		if(!smart.PathIsFile(theTlsCertPath)) {
			log.Println("[ERROR] INIT TLS: No certificate crt found in current directory. Please provide a valid cert:", theTlsCertPath)
			return false
		} //end if
		if(!smart.PathIsFile(theTlsKeyPath)) {
			log.Println("[ERROR]: INIT TLS No certificate key found in current directory. Please provide a valid cert:", theTlsKeyPath)
			return false
		} //end if
		log.Println("[NOTICE] ... serving HTTPS / TLS at " + httpAddr + " on port", httpPort)
		go srv.ListenAndServeTLS(theTlsCertPath, theTlsKeyPath)
	} else { // serve HTTP
		log.Println("[NOTICE] ... serving HTTP at " + httpAddr + " on port", httpPort)
		go srv.ListenAndServe()
	} //end if

	//--
	return true
	//--

} //END FUNCTION


// #END
