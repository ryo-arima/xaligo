package controller_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ryo-arima/xaligo/pkg/controller"
)

func TestRunInit_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	if err := controller.RunInit(dir); err != nil {
		t.Fatalf("RunInit: %v", err)
	}
	path := filepath.Join(dir, "sample.xal")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected %s to exist", path)
	}
}

func TestRunInit_FileContainsFrame(t *testing.T) {
	dir := t.TempDir()
	if err := controller.RunInit(dir); err != nil {
		t.Fatalf("RunInit: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "sample.xal"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Error("sample.xal should not be empty")
	}
	content := string(data)
	if len(content) < 7 || content[:7] != "<frame " {
		// Just check that it contains <frame
		found := false
		for i := 0; i < len(content)-6; i++ {
			if content[i:i+7] == "<frame " {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("sample.xal does not contain <frame tag")
		}
	}
}

func TestRunInit_CreatesOutputDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")
	if err := controller.RunInit(dir); err != nil {
		t.Fatalf("RunInit: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("expected output dir %s to be created", dir)
	}
}

func TestInitInitCmd_NotNil(t *testing.T) {
	cmd := controller.InitInitCmd()
	if cmd == nil {
		t.Error("InitInitCmd() returned nil")
	}
}
