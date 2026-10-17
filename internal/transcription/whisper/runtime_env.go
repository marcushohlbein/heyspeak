package whisper

import (
	"os"
	"path/filepath"
)

// ensureRuntimeEnv configures whisper.cpp/ggml runtime paths for local development.
func ensureRuntimeEnv() {
	ensureMetalResources()
}

func ensureMetalResources() {
	if os.Getenv("GGML_METAL_PATH_RESOURCES") != "" {
		return
	}
	root := whisperSourceRoot()
	if root == "" {
		return
	}
	metal := filepath.Join(root, "ggml", "src", "ggml-metal")
	if _, err := os.Stat(filepath.Join(metal, "kernels")); err != nil {
		return
	}
	_ = os.Setenv("GGML_METAL_PATH_RESOURCES", metal)
}

func whisperSourceRoot() string {
	if r := os.Getenv("SPEEK_WHISPER_ROOT"); r != "" {
		return r
	}
	for _, rel := range []string{
		"third_party/whisper.cpp",
		filepath.Join("..", "..", "third_party", "whisper.cpp"),
	} {
		abs, err := filepath.Abs(rel)
		if err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(abs, "include", "whisper.h")); err == nil {
			return abs
		}
	}
	return ""
}
