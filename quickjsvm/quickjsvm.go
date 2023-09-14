
// GO Lang :: SmartGo Extra / QuickJsVm :: Smart.Go.Framework
// (c) 2020-2023 unix-world.org
// r.20230914.1548 :: STABLE

// Req: go 1.16 or later
package quickjsvm

import (
	"runtime"
	"errors"
	"log"

	"time"
	"sort"

	"github.com/unix-world/smartgoext/quickjs"
	smart "github.com/unix-world/smartgo"
)

const VERSION string = "r.20230914.1548"


type quickJsVmEvalResult struct {
	jsEvErr string
	jsEvRes string
}


func quickJsVmEvalCode(jsCode string, jsMemMB uint16, jsInputData map[string]string, jsExtendMethods map[string]interface{}, jsBinaryCodePreload map[string][]byte) quickJsVmEvalResult {

	//-- safety lock: dealing with QuickJs C methods req. this
	runtime.LockOSThread()
//	defer runtime.UnlockOSThread() // if the calling goroutine exits without unlocking the thread, the thread will be terminated ; thus is safer to not reuse this kind of threads to avoid C garbage, so close any thread using this method after this method returns !!
	//-- #safety lock

	//--
	if(jsMemMB < 2) {
		return quickJsVmEvalResult{ jsEvErr: "ERR: Minimum Memory Size for JS Eval Code is 2MB", jsEvRes: "" }
	} else if(jsMemMB > 1024) {
		return quickJsVmEvalResult{ jsEvErr: "ERR: Minimum Memory Size for JS Eval Code is 1024MB", jsEvRes: "" }
	} //end if else
	//--

	//--
	if(jsInputData == nil) {
		jsInputData = map[string]string{}
	} //end if
	//--
	if(jsExtendMethods == nil) {
		jsExtendMethods = map[string]interface{}{}
	} //end if
	//--
	if(jsBinaryCodePreload == nil) {
		jsBinaryCodePreload = map[string][]byte{}
	} //end if
	//--

	//--
	quickjsCheck := func(err error, result quickjs.Value) (theErr string, theCause string, theStack string) {
		//--
		if err != nil {
			var evalErr *quickjs.Error
			var cause string = ""
			var stack string = ""
			if errors.As(err, &evalErr) {
				cause = evalErr.Cause
				stack = evalErr.Stack
			} //end if
			return err.Error(), cause, stack
		} //end if
		//--
		if(result.IsException()) {
			return "WARN: JS Exception !", "", ""
		} //end if
		//--
		return "", "", ""
		//--
	} //end function
	//--

	//--
	sleepTimeMs := func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		var mseconds uint64 = 0
		var msecnds string = ""
		for _, vv := range args {
			msecnds = vv.String()
			mseconds = smart.ParseStrAsUInt64(msecnds)
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
		} //end for
		jsonArgs := smart.JsonEncode(theArgs)
		//--
		jsonStruct, jsonErrStruct := smart.JsonObjDecode(jsonArgs)
		if(jsonErrStruct != nil) {
			log.Println("[ERROR] Malformed JSON Data:", jsonErrStruct)
		} else if(jsonStruct != nil) {
			var txt string = ""
			keys := make([]string, 0)
			for xx, _ := range jsonStruct {
				keys = append(keys, xx)
			} //end for
			sort.Strings(keys) // need to be added in order {{{SYNC-GOLANG-ORDERED-RANGE-BY-KEYS}}}
			for _, zz := range keys {
				txt += jsonStruct[zz].(string) + " "
			} //end for
			log.Println(txt)
		} //end if
		//--
	//	return ctx.String("")
		return ctx.String(jsonArgs)
		//--
	} //end function
	//--

	//--
	jsvm := quickjs.NewRuntime()
	defer jsvm.Free()
	var MB uint32 = 1 << 10 << 10
	var vmMem uint32 = uint32(jsMemMB) * MB
	log.Println("[DEBUG] Settings: JsVm Memory Limit:", jsMemMB, "MB")
	jsvm.SetMemoryLimit(vmMem)
	//--
	context := jsvm.NewContext()
	defer context.Free()
	//--
	globals := context.Globals()
	jsInputData["JSON"] = "GoLangQuickJsVm"
	json := context.String(smart.JsonEncode(jsInputData))
	globals.Set("jsonInput", json)
	//--
	globals.Set("SmartJsVm_consoleLog", context.Function(consoleLog)) // set method for Javascript
	globals.SetFunction("SmartJsVm_sleepTimeMs", sleepTimeMs) // the same as above ...
	//--
	for k, v := range jsExtendMethods {
		globals.SetFunction("SmartJsVmX_" + k, v.(func(*quickjs.Context, quickjs.Value, []quickjs.Value)(quickjs.Value)))
	} //end for
	//--
	keys := make([]string, 0)
	for xx, _ := range jsBinaryCodePreload {
		keys = append(keys, xx)
	} //end for
	sort.Strings(keys) // need to be loaded in order {{{SYNC-GOLANG-ORDERED-RANGE-BY-KEYS}}}
	var i int = 0
	for _, zz := range keys {
		log.Println("[DEBUG] JsVm Pre-Loading Binary Opcode JS:", zz, "@", i)
		i++
		bload, _ := context.EvalBinary(jsBinaryCodePreload[zz], quickjs.EVAL_GLOBAL)
		defer bload.Free()
	} //end for
	//--
	result, err := context.Eval(jsCode, quickjs.EVAL_GLOBAL) // quickjs.EVAL_STRICT
	jsvm.ExecutePendingJob() // req. to execute promises: ex: `new Promise(resolve => resolve('testPromise')).then(msg => console.log('******* Promise Solved *******', msg));`
	defer result.Free()
	jsErr, _, _ := quickjsCheck(err, result)
	if(jsErr != "") {
		return quickJsVmEvalResult{ jsEvErr: "ERR: JS Eval Error: " + "`" + jsErr + "`", jsEvRes: "" }
	} //end if
	//--

	//--
	return quickJsVmEvalResult{ jsEvErr: "", jsEvRes: result.String() }
	//--

} //END FUNCTION


func QuickJsVmRunCode(jsCode string, stopTimeout uint32, jsMemMB uint16, jsInputData map[string]string, jsExtendMethods map[string]interface{}, jsBinaryCodePreload map[string][]byte) (jsEvErr string, jsEvRes string) {
	//-- check if Js Code is Empty
	if(smart.StrTrimWhitespaces(jsCode) == "") {
		return "QuickJsVmRunCode: Empty Javascript Code ...", ""
	} //end if
	//-- a reasonable execution timeout is 5 minutes ... but it depends
	if(stopTimeout > 86400) {
		return "QuickJsVmRunCode: Max Javascript Code Execution TimeOut that can be set is 86400 second(s)", "" // max execution timeout: 24 hours
	} else if(stopTimeout < 1) {
		return "QuickJsVmRunCode: Min Javascript Code Execution TimeOut that can be set is 86400 second(s)", "" // min execution timeout: 1 second
	} //end if
	//--
	log.Println("[DEBUG] Settings: JsVm Execution Timeout Limit:", stopTimeout, "second(s)")
	//--
	c1 := make(chan quickJsVmEvalResult, 1)
	//--
	go func() {
		quickJsVmEvalResult := quickJsVmEvalCode(jsCode, jsMemMB, jsInputData, jsExtendMethods, jsBinaryCodePreload) // (jsEvErr string, jsEvRes string)
		c1 <- quickJsVmEvalResult
	}()
	//--
	select {
		case res := <-c1:
			return res.jsEvErr, res.jsEvRes
		case <-time.After(time.Duration(stopTimeout) * time.Second):
			return "QuickJsVmRunCode: Javascript Code Execution reached the Maximum allowed TimeOut Limit (as set), after " + smart.ConvertUInt32ToStr(stopTimeout) + " second(s) ...", ""
	} //end select
	//--
} //END FUNCTION


// #END
