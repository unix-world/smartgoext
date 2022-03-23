
// GO Lang :: SmartGo Extra :: Smart.Go.Framework
// (c) 2020-2022 unix-world.org
// r.20220323.1640 :: STABLE

package smartgoext

// REQUIRE: go 1.15 or later

import (
	"os"
	"runtime"
	"errors"
	"log"

	"time"
	"sort"

	"fmt"
	"strconv"
	"bytes"

	"net/http"
	"crypto/subtle"
	"golang.org/x/net/webdav"

	"github.com/unix-world/smartgoext/quickjs"
	smart "github.com/unix-world/smartgo"
)


//-----

func WebdavServer(authUser string, authPass string, httpAddr string, httpPort uint16, httpsPort uint16, serveSecure bool, disableUnsecure bool) bool {

	//-- local defines

	const (
		THE_VERSION = "r.20220323.1440"

		CONN_HOST = "0.0.0.0"
		CONN_PORT = 13080
		CONN_TLSPORT = 13443

		ADMIN_USER = "admin"
		ADMIN_PASSWORD = "pass"

		STORAGE_DIR = "./wdav"
		DAV_PATH = "/webdav"

		TLS_CERT = "./ssl/cert.crt"
		TLS_KEY = "./ssl/cert.key"
	)

	//-- fixes

	if(smart.StrTrimWhitespaces(authUser) == "") || (smart.StrTrimWhitespaces(authPass) == "") {
		authUser = ADMIN_USER
		authPass = ADMIN_PASSWORD
	} //end if

	httpAddr = smart.StrTrimWhitespaces(httpAddr)
	if(httpAddr == "") {
		httpAddr = CONN_HOST
	} //end if

	if(httpPort <= 0 || httpPort > 65535) {
		httpPort = CONN_PORT
	} //end if

	if(httpsPort <= 0 || httpsPort > 65535) {
		httpsPort = CONN_TLSPORT
	} //end if

	//-- get cwd path

	path, err := os.Getwd()
	if(err != nil) {
		log.Println("[ERROR] WebDAV Server: Cannot Get Current Path: ", err)
		return false
	} //end if

	//-- test params

	if(disableUnsecure == true && serveSecure != true) {
		log.Println("[ERROR] WebDAV Server INIT: The both HTTP and HTTPS modes are disabled ... server will exit ...")
		return false
	} //end if

	//-- for web

	var serverSignature bytes.Buffer
	serverSignature.WriteString("WebDAV GO Server " + THE_VERSION + "\n")
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
	fmt.Println("WebDAV GO Server " + THE_VERSION)
	fmt.Println("---------------------------------------------------------------------------")
	fmt.Println("Current Path: " + string(path))
	fmt.Println("DAV Folder: " + STORAGE_DIR)
	fmt.Println("WebDAV Path: " + DAV_PATH)
	fmt.Println("---------------------------------------------------------------------------")
	if disableUnsecure != true {
		fmt.Println("Unsecure Listening at http://" + httpAddr + ":" + smart.ConvertUInt16ToStr(httpPort) + "/")
	}
	if serveSecure == true {
		fmt.Println("Secure Listening TLS at https://" + httpAddr + ":" + smart.ConvertUInt16ToStr(httpsPort) + "/")
	}
	fmt.Println("===========================================================================")

	//--

	srv := &webdav.Handler{
		Prefix:     DAV_PATH,
		FileSystem: webdav.Dir(STORAGE_DIR),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("[WARNING] WebDAV GO Server :: WEBDAV.ERROR: %s [%s %s %s] %s [%s] %s\n", err, r.Method, r.URL, r.Proto, "*", r.Host, r.RemoteAddr)
			} else {
				log.Printf("[OK] WebDAV GO Server :: WEBDAV [%s %s %s] %s [%s] %s\n", r.Method, r.URL, r.Proto, "*", r.Host, r.RemoteAddr)
			}
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var statusCode = 202
		if r.URL.Path != "/" {
			statusCode = 404
			w.WriteHeader(statusCode)
			w.Write([]byte("404 Not Found\n"))
			log.Printf("[ERROR] WebDAV GO Server :: DEFAULT.ERROR [%s %s %s] %s [%s] %s\n", r.Method, r.URL, r.Proto, strconv.Itoa(statusCode), r.Host, r.RemoteAddr)
			return
		}
		log.Printf("[OK] WebDAV GO Server :: DEFAULT [%s %s %s] %s [%s] %s\n", r.Method, r.URL, r.Proto, strconv.Itoa(statusCode), r.Host, r.RemoteAddr)
		w.Header().Set("Content-Type", smart.HTML_CONTENT_HEADER)
		w.WriteHeader(statusCode) // status code must be after content type
		var titleText string = "WebDAV GO Server " + THE_VERSION
		var headHtml string = "<style>" + "\n" + "div.status { text-align:center; margin:10px; cursor:help; }" + "\n" + "div.signature { background:#778899; color:#FFFFFF; font-size:2rem; font-weight:bold; text-align:center; border-radius:3px; padding:10px; margin:20px; }" + "\n" + "</style>"
		var bodyHtml string = `<div class="status"><img alt="Status: Up and Running ..." title="Status: Up and Running ..." width="96" height="96" src="data:image/svg+xml,` + smart.EscapeHtml(smart.EscapeUrl(smart.SVG_SPIN)) + `"></div>` + "\n" + `<div class="signature">` + "\n" + "<pre>" + "\n" + smart.EscapeHtml(serverSignature.String()) + "</pre>" + "\n" + "</div>"
		w.Write([]byte(smart.HtmlSimpleTemplate(titleText, headHtml, bodyHtml)))
	})

	// http version handler
	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		var statusCode = 203
		log.Printf("[OK] WebDAV GO Server :: VERSION [%s %s %s] %s [%s] %s\n", r.Method, r.URL, r.Proto, strconv.Itoa(statusCode), r.Host, r.RemoteAddr)
		// plain/text
		w.WriteHeader(statusCode)
		w.Write([]byte("WebDAV GO Server " + THE_VERSION + "\n"))
	})

	// webdav handler
	http.HandleFunc(DAV_PATH+"/", func(w http.ResponseWriter, r *http.Request) {
		// test if basic auth
		user, pass, ok := r.BasicAuth()
		// check if basic auth and if credentials match
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(authUser)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(authPass)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="WebDAV GO Server Storage Area"`)
			w.WriteHeader(401) // status code must be after set headers
			w.Write([]byte("401 Unauthorized\n"))
			log.Printf("[WARNING] WebDAV GO Server :: AUTH.FAILED [%s %s %s] %s [%s] %s\n", r.Method, r.URL, r.Proto, "401", r.Host, r.RemoteAddr)
			return
		}
		// if all ok above (basic auth + credentials ok, serve ...)
		srv.ServeHTTP(w, r)
	})

	if serveSecure == true {
		if(!smart.PathIsFile(TLS_CERT)) {
			log.Println("[ERROR] INIT TLS: No certificate crt found in current directory. Please provide a valid cert:", TLS_CERT)
			return false
		}
		if(!smart.PathIsFile(TLS_KEY)) {
			log.Println("[ERROR]: INIT TLS No certificate key found in current directory. Please provide a valid cert:", TLS_KEY)
			return false
		}
		log.Println("[NOTICE] ... serving HTTPS / TLS on port", httpsPort)
		go http.ListenAndServeTLS(httpAddr + fmt.Sprintf(":%d", httpsPort), TLS_CERT, TLS_KEY, nil)
	}
	if disableUnsecure != true {
		log.Println("[NOTICE] ... serving HTTP on port", httpPort)
		go http.ListenAndServe(httpAddr + fmt.Sprintf(":%d", httpPort), nil)
	}

	if disableUnsecure == true && serveSecure != true {
		log.Println("[ERROR] ... WebDAV NOT Started")
		return false
	}

	//--
	return true
	//--

} //END FUNCTION

//-----

func QuickJsEvalCode(jsCode string, jsMemMB uint16, jsInputData map[string]string, jsExtendMethods map[string]interface{}, jsBinaryCodePreload map[string][]byte) (jsEvErr string, jsEvRes string) {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if(jsMemMB < 2) {
		return "ERR: Minimum Memory Size for JS Eval Code is 2MB", ""
	} else if(jsMemMB > 1024) {
		return "ERR: Minimum Memory Size for JS Eval Code is 1024MB", ""
	} //end if else

	if(jsInputData == nil) {
		jsInputData = map[string]string{}
	}

	//--
	quickjsCheck := func(err error, result quickjs.Value) (theErr string, theCause string, theStack string) {
		if err != nil {
			var evalErr *quickjs.Error
			var cause string = ""
			var stack string = ""
			if errors.As(err, &evalErr) {
				cause = evalErr.Cause
				stack = evalErr.Stack
			}
			return err.Error(), cause, stack
		}
		if(result.IsException()) {
			return "WARN: JS Exception !", "", ""
		}
		return "", "", ""
	} //end function
	//--

	//--
	sleepTimeMs := func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		var mseconds uint64 = 0
		var msecnds string = ""
		for _, vv := range args {
			msecnds = vv.String()
			mseconds = smart.ParseIntegerStrAsUInt64(msecnds)
		} //end for
		if(mseconds > 1 && mseconds < 3600 * 1000) {
			time.Sleep(time.Duration(mseconds) * time.Millisecond)
		} //end if
		return ctx.String(msecnds)
	} //end function
	//--
	consoleLog := func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		//--
		theArgs := map[string]string{}
		for kk, vv := range args {
			theArgs["arg:" + smart.ConvertIntToStr(kk)] = vv.String()
		}
		jsonArgs := smart.JsonEncode(theArgs)
		//--
		jsonStruct := smart.JsonDecode(jsonArgs)
		if(jsonStruct != nil) {
			var txt string = ""
			keys := make([]string, 0)
			for xx, _ := range jsonStruct {
				keys = append(keys, xx)
			}
			sort.Strings(keys)
			for _, zz := range keys {
				txt += jsonStruct[zz].(string) + " "
			}
			log.Println(txt)
		} //end if
		//--
	//	return ctx.String("")
		return ctx.String(jsonArgs)
	} //end function
	//--

	//--
	jsvm := quickjs.NewRuntime()
	defer jsvm.Free()
	var MB uint32 = 1 << 10 << 10
	var vmMem uint32 = uint32(jsMemMB) * MB
	log.Println("[DEBUG] Settings: JsVm Memory Limit to", jsMemMB, "MB")
	jsvm.SetMemoryLimit(vmMem)
	//--
	context := jsvm.NewContext()
	defer context.Free()
	//--
	globals := context.Globals()
	jsInputData["JSON"] = "GoLang"
	json := context.String(smart.JsonEncode(jsInputData))
	globals.Set("jsonInput", json)
	//--
	globals.Set("sleepTimeMs", context.Function(sleepTimeMs))
	globals.SetFunction("consoleLog", consoleLog) // the same as above
	//--
	for k, v := range jsExtendMethods {
		globals.SetFunction(k, v.(func(*quickjs.Context, quickjs.Value, []quickjs.Value)(quickjs.Value)))
	}
	//--
	for y, b := range jsBinaryCodePreload {
		if(b != nil) {
			log.Println("[DEBUG] Pre-Loading Binary Opcode JS:", y)
			bload, _ := context.EvalBinary(b, quickjs.EVAL_GLOBAL)
			defer bload.Free()
		}
	}
	result, err := context.Eval(jsCode, quickjs.EVAL_GLOBAL) // quickjs.EVAL_STRICT
	jsvm.ExecutePendingJob() // req. to execute promises: ex: `new Promise(resolve => resolve('testPromise')).then(msg => console.log('******* Promise Solved *******', msg));`
	defer result.Free()
	jsErr, _, _ := quickjsCheck(err, result)
	if(jsErr != "") {
		return "ERR: JS Eval Error: " + "`" + jsErr + "`", ""
	} //end if
	//--

	//--
	return "", result.String()
	//--

} //END FUNCTION


//-----


// #END
