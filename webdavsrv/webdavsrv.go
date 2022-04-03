
// GO Lang :: SmartGo Extra / WebDAV Server :: Smart.Go.Framework
// (c) 2020-2022 unix-world.org
// r.20220403.1947 :: STABLE

package webdavsrv

// REQUIRE: go 1.15 or later

import (
	"log"

	"fmt"
	"strconv"
	"bytes"

	"net/http"
	"crypto/subtle"
	"golang.org/x/net/webdav"

	smart "github.com/unix-world/smartgo"
)

const (
	THE_VERSION = "r.20220403.1947"

	CONN_HOST = "127.0.0.1"
	CONN_PORT = 13080
	CONN_TLSPORT = 13443

	STORAGE_DIR = "./wdav"
	DAV_PATH = "/webdav"

	TLS_PATH = "./ssl"
	TLS_CERT = "cert.crt"
	TLS_KEY = "cert.key"
)


func WebdavServerListenAndServe(authUser string, authPass string, httpAddr string, httpPort uint16, httpsPort uint16, serveSecure bool, disableUnsecure bool, certifPath string, storagePath string) bool {

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

	if(!smart.IsNetValidPortNum(int64(httpsPort))) {
		httpsPort = CONN_TLSPORT
	} //end if

	//-- paths

	certifPath = smart.StrTrimWhitespaces(certifPath)
	if((certifPath == "") || (smart.PathIsBackwardUnsafe(certifPath) == true)) {
		certifPath = TLS_PATH
	} //end if
	certifPath = smart.PathGetAbsoluteFromRelative(certifPath)
	certifPath = smart.PathAddDirLastSlash(certifPath)

	storagePath = smart.StrTrimWhitespaces(storagePath)
	if((storagePath == "") || (smart.PathIsBackwardUnsafe(storagePath) == true)) {
		storagePath = STORAGE_DIR
	} //end if
	storagePath = smart.PathGetAbsoluteFromRelative(storagePath)
	storagePath = smart.PathAddDirLastSlash(storagePath)

	//-- test params

	if(disableUnsecure == true && serveSecure != true) {
		log.Println("[ERROR] WebDAV Server INIT: The both HTTP and HTTPS modes are disabled ... server will exit ...")
		return false
	} //end if

	//-- for web

	var serverSignature bytes.Buffer
	serverSignature.WriteString("GO WebDAV Server " + THE_VERSION + "\n")
	serverSignature.WriteString("(c) 2020-2022 unix-world.org" + "\n")
	serverSignature.WriteString("\n")
	if(disableUnsecure != true) {
		serverSignature.WriteString("<URL> :: http://" + httpAddr + ":" + smart.ConvertUInt16ToStr(httpPort) + DAV_PATH + "/" + "\n")
	} //end if
	if(serveSecure == true) {
		serverSignature.WriteString("<Secure URL> :: https://" + httpAddr + ":" + smart.ConvertUInt16ToStr(httpsPort) + DAV_PATH + "/" + "\n")
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
	if(disableUnsecure != true) {
		fmt.Println("Unsecure Listening at http://" + httpAddr + ":" + smart.ConvertUInt16ToStr(httpPort) + "/")
	} //end if
	if(serveSecure == true) {
		fmt.Println("Secure Listening TLS at https://" + httpAddr + ":" + smart.ConvertUInt16ToStr(httpsPort) + "/")
	} //end if
	fmt.Println("===========================================================================")

	//-- handler

	srv := &webdav.Handler{
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

	//-- handle methods

	// http root handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var statusCode = 202
		if(r.URL.Path != "/") {
			statusCode = 404
			w.WriteHeader(statusCode)
			w.Write([]byte("404 Not Found\n"))
			log.Printf("[ERROR] GO WebDAV Server :: DEFAULT.ERROR [%s %s %s] %s [%s] %s\n", r.Method, r.URL, r.Proto, strconv.Itoa(statusCode), r.Host, r.RemoteAddr)
			return
		} //end if
		log.Printf("[OK] GO WebDAV Server :: DEFAULT [%s %s %s] %s [%s] %s\n", r.Method, r.URL, r.Proto, strconv.Itoa(statusCode), r.Host, r.RemoteAddr)
		w.Header().Set("Content-Type", smart.HTML_CONTENT_HEADER)
		w.WriteHeader(statusCode) // status code must be after content type
		var titleText string = "GO WebDAV Server " + THE_VERSION
		var headHtml string = "<style>" + "\n" + "div.status { text-align:center; margin:10px; cursor:help; }" + "\n" + "div.signature { background:#778899; color:#FFFFFF; font-size:2rem; font-weight:bold; text-align:center; border-radius:3px; padding:10px; margin:20px; }" + "\n" + "</style>"
		var bodyHtml string = `<div class="status"><img alt="Status: Up and Running ..." title="Status: Up and Running ..." width="64" height="64" src="data:image/svg+xml,` + smart.EscapeHtml(smart.EscapeUrl(smart.ReadAsset("svg/loading-spin.svg"))) + `"></div>` + "\n" + `<div class="signature">` + "\n" + "<pre>" + "\n" + smart.EscapeHtml(serverSignature.String()) + "</pre>" + "\n" + "</div>"
		w.Write([]byte(smart.HtmlStaticTemplate(titleText, headHtml, bodyHtml)))
	})

	// http version handler
	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		var statusCode = 203
		log.Printf("[OK] GO WebDAV Server :: VERSION [%s %s %s] %s [%s] %s\n", r.Method, r.URL, r.Proto, strconv.Itoa(statusCode), r.Host, r.RemoteAddr)
		// plain/text
		w.WriteHeader(statusCode)
		w.Write([]byte("GO WebDAV Server " + THE_VERSION + "\n"))
	})

	// webdav handler
	http.HandleFunc(DAV_PATH+"/", func(w http.ResponseWriter, r *http.Request) {
		// test if basic auth
		user, pass, ok := r.BasicAuth()
		// check if basic auth and if credentials match
		if(!ok || subtle.ConstantTimeCompare([]byte(user), []byte(authUser)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(authPass)) != 1) {
			w.Header().Set("WWW-Authenticate", `Basic realm="GO WebDAV Server Storage Area"`)
			w.WriteHeader(401) // status code must be after set headers
			w.Write([]byte("401 Unauthorized\n"))
			log.Printf("[WARNING] GO WebDAV Server :: AUTH.FAILED [%s %s %s] %s [%s] %s\n", r.Method, r.URL, r.Proto, "401", r.Host, r.RemoteAddr)
			return
		} //end if
		// if all ok above (basic auth + credentials ok, serve ...)
		srv.ServeHTTP(w, r)
	})

	// serve logic

	var theTlsCertPath string = certifPath + TLS_CERT
	var theTlsKeyPath  string = certifPath + TLS_KEY

	if(serveSecure == true) {
		if(!smart.PathIsFile(theTlsCertPath)) {
			log.Println("[ERROR] INIT TLS: No certificate crt found in current directory. Please provide a valid cert:", theTlsCertPath)
			return false
		} //end if
		if(!smart.PathIsFile(theTlsKeyPath)) {
			log.Println("[ERROR]: INIT TLS No certificate key found in current directory. Please provide a valid cert:", theTlsKeyPath)
			return false
		} //end if
		log.Println("[NOTICE] ... serving HTTPS / TLS at " + httpAddr + " on port", httpsPort)
		go http.ListenAndServeTLS(httpAddr + fmt.Sprintf(":%d", httpsPort), theTlsCertPath, theTlsKeyPath, nil)
	} //end if
	if(disableUnsecure != true) {
		log.Println("[NOTICE] ... serving HTTP at " + httpAddr + " on port", httpPort)
		go http.ListenAndServe(httpAddr + fmt.Sprintf(":%d", httpPort), nil)
	} //end if

	//-- final checks

	if(disableUnsecure == true && serveSecure != true) {
		log.Println("[ERROR] ... WebDAV NOT Started")
		return false
	} //end if

	//--
	return true
	//--

} //END FUNCTION


// #END
