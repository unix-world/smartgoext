
// GO Lang :: SmartGo Extra :: Smart.Go.Framework
// (c) 2020-2022 unix-world.org
// r.20220314.1554 :: STABLE

package smartgoext

// REQUIRE: go 1.15 or later

import (
	"runtime"
	"errors"
	"log"

	"time"
	"sort"

	"github.com/unix-world/smartgoext/quickjs"
	smart "github.com/unix-world/smartgo"
)

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
