package controller_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ryo-arima/xaligo/pkg/controller"
)

func TestRunRender_SimpleXAL(t *testing.T) {
	xal := `<frame width="800" height="600">
	<card title="Hello" />
</frame>`
	tmpDir := t.TempDir()
	inPath := filepath.Join(tmpDir, "test.xal")
	outPath := filepath.Join(tmpDir, "out.excalidraw")

	if err := os.WriteFile(inPath, []byte(xal), 0644); err != nil {
		t.Fatalf("write xal: %v", err)
	}

	if err := controller.RunRender(inPath, outPath); err != nil {
		t.Fatalf("RunRender: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(data), "excalidraw") {
		t.Error("output should contain 'excalidraw'")
	}
}

func TestRunRender_NonExistentInput_Error(t *testing.T) {
	err := controller.RunRender("/nonexistent/input.xal", "/tmp/out.excalidraw")
	if err == nil {
		t.Error("expected error for non-existent input file")
	}
}

func TestRunRender_InvalidXAL_Error(t *testing.T) {
	tmpDir := t.TempDir()
	inPath := filepath.Join(tmpDir, "bad.xal")
	outPath := filepath.Join(tmpDir, "out.excalidraw")

	if err := os.WriteFile(inPath, []byte("<container>not a frame</container>"), 0644); err != nil {
		t.Fatalf("write xal: %v", err)
	}

	if err := controller.RunRender(inPath, outPath); err == nil {
		t.Error("expected error for invalid XAL (non-frame root)")
	}
}

func TestInitRenderCmd_NotNil(t *testing.T) {
	cmd := controller.InitRenderCmd()
	if cmd == nil {
		t.Error("InitRenderCmd() returned nil")
	}
}
