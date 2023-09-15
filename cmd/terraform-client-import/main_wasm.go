//go:build wasm && js

package main

import (
	"fmt"
	"path/filepath"
	"syscall/js"

	"github.com/hashicorp/go-hclog"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/magodo/tfadd/tfadd"
	"github.com/magodo/tfstate"
	"github.com/zclconf/go-cty/cty"
)

type fn func(this js.Value, args []js.Value) (any, error)

var (
	jsErr     js.Value = js.Global().Get("Error")
	jsPromise js.Value = js.Global().Get("Promise")
)

func main() {
	js.Global().Set("terraformImport", asyncFunc(wasmMain))
	<-make(chan interface{})
}

func asyncFunc(innerFunc fn) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(_ js.Value, promFn []js.Value) any {
			resolve, reject := promFn[0], promFn[1]

			go func() {
				defer func() {
					if r := recover(); r != nil {
						reject.Invoke(jsErr.New(fmt.Sprint("panic:", r)))
					}
				}()

				res, err := innerFunc(this, args)
				if err != nil {
					reject.Invoke(jsErr.New(err.Error()))
				} else {
					resolve.Invoke(res)
				}
			}()

			return nil
		})

		return jsPromise.New(handler)
	})
}

func wasmMain(_ js.Value, args []js.Value) (any, error) {
	//wasmName, wasmPath, rt, id, cfg string
	if len(args) != 5 {
		return nil, fmt.Errorf("expected 5 arguments, got %d", len(args))
	}
	fset := FlagSet{
		WasmName:     args[0].String(),
		WasmPath:     args[1].String(),
		ResourceType: args[2].String(),
		ResourceId:   args[3].String(),
		ProviderCfg:  args[4].String(),
	}
	logger := hclog.New(&hclog.LoggerOptions{
		Output: hclog.DefaultOutput,
		Level:  hclog.LevelFromString(fset.LogLevel),
		Name:   filepath.Base(fset.PluginPath),
	})
	state, schema, err := importResource(logger, fset)
	if err != nil {
		return nil, err
	}

	tfAddr := fset.ResourceType + ".example"
	hcl, err := json2hcl(tfAddr, fset.WasmName, *state, fset.ResourceType, *schema)
	return js.ValueOf(hcl), err
}

func json2hcl(tfAddress string, provider string, state cty.Value, rt string, schema tfjson.Schema) (string, error) {
	b, err := tfadd.GenerateForOneResource(
		&schema,
		tfstate.StateResource{
			Mode:         tfjson.ManagedResourceMode,
			Address:      tfAddress,
			Type:         rt,
			ProviderName: provider,
			Value:        state,
		},
		false)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
