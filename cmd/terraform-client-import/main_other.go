//go:build !wasm && !js

package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
)

func main() {
	var fset FlagSet
	flag.StringVar(&fset.WasmName, "wasmName", "", "The path to the plugin")
	flag.StringVar(&fset.WasmPath, "wasmPath", "", "The path to the plugin")
	flag.StringVar(&fset.PluginPath, "path", "", "The path to the plugin")
	flag.StringVar(&fset.ResourceType, "type", "", "The resource type")
	flag.StringVar(&fset.ResourceId, "id", "", "The resource id")
	flag.StringVar(&fset.LogLevel, "log-level", hclog.Error.String(), "Log level")
	flag.StringVar(&fset.ProviderCfg, "cfg", "{}", "The content of provider config block in JSON")
	flag.Var(&fset.StatePatches, "state-patch", "The JSON patch to the state after importing, which will then be used as the prior state for reading. Can be specified multiple times")
	flag.IntVar(&fset.TimeoutSec, "timeout", 0, "Timeout in second. Defaults to no timeout.")

	flag.Parse()

	logger := hclog.New(&hclog.LoggerOptions{
		Output: hclog.DefaultOutput,
		Level:  hclog.LevelFromString(fset.LogLevel),
		Name:   filepath.Base(fset.PluginPath),
	})

	if err := realMain(logger, fset); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
