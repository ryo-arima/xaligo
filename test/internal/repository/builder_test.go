package repository_test

import (
	"testing"

	"github.com/ryo-arima/xaligo/internal/repository"
)

func TestMakeText_Type(t *testing.T) {
	el := repository.MakeText("id1", 0, 0, 100, 20, "hello", 14, "#000000", false, 1)
	if el["type"] != "text" {
		t.Errorf("type = %v, want %q", el["type"], "text")
	}
}

func TestMakeText_Text(t *testing.T) {
	el := repository.MakeText("id1", 0, 0, 100, 20, "hello", 14, "#000000", false, 1)
	if el["text"] != "hello" {
		t.Errorf("text = %v, want %q", el["text"], "hello")
	}
	if el["rawText"] != "hello" {
		t.Errorf("rawText = %v, want %q", el["rawText"], "hello")
	}
}

func TestMakeText_ID(t *testing.T) {
	el := repository.MakeText("my-id", 10, 20, 100, 30, "txt", 12, "#000000", false, 2)
	if el["id"] != "my-id" {
		t.Errorf("id = %v, want %q", el["id"], "my-id")
	}
}

func TestMakeText_FontSize(t *testing.T) {
	el := repository.MakeText("id1", 0, 0, 100, 20, "text", 18, "#000000", false, 1)
	if el["fontSize"] != 18 {
		t.Errorf("fontSize = %v, want 18", el["fontSize"])
	}
}

func TestMakeText_StrokeColor(t *testing.T) {
	el := repository.MakeText("id1", 0, 0, 100, 20, "text", 14, "#ff0000", false, 1)
	if el["strokeColor"] != "#ff0000" {
		t.Errorf("strokeColor = %v, want %q", el["strokeColor"], "#ff0000")
	}
}

func TestMakeImage_Type(t *testing.T) {
	el := repository.MakeImage("img1", 0, 0, 64, 64, "file-id", "transparent", 1)
	if el["type"] != "image" {
		t.Errorf("type = %v, want %q", el["type"], "image")
	}
}

func TestMakeImage_FileID(t *testing.T) {
	el := repository.MakeImage("img1", 0, 0, 64, 64, "file-id-abc", "transparent", 1)
	if el["fileId"] != "file-id-abc" {
		t.Errorf("fileId = %v, want %q", el["fileId"], "file-id-abc")
	}
}

func TestMakeImage_BackgroundColor(t *testing.T) {
	el := repository.MakeImage("img1", 0, 0, 64, 64, "file-id", "#ff5500", 1)
	if el["backgroundColor"] != "#ff5500" {
		t.Errorf("backgroundColor = %v, want %q", el["backgroundColor"], "#ff5500")
	}
}

func TestMakeImage_ID(t *testing.T) {
	el := repository.MakeImage("my-img", 5, 10, 32, 32, "file-id", "transparent", 3)
	if el["id"] != "my-img" {
		t.Errorf("id = %v, want %q", el["id"], "my-img")
	}
}

func TestMakeImage_Status(t *testing.T) {
	el := repository.MakeImage("img1", 0, 0, 64, 64, "file-id", "transparent", 1)
	if el["status"] != "saved" {
		t.Errorf("status = %v, want %q", el["status"], "saved")
	}
}
