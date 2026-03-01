package repository

import (
	"testing"
)

func TestMakeText_Fields(t *testing.T) {
	el := MakeText("txt-1", 10, 20, 100, 30, "Hello", 14, "#ff0000", false, 42)

	if el["type"] != "text" {
		t.Errorf("type = %v, want text", el["type"])
	}
	if el["text"] != "Hello" {
		t.Errorf("text = %v, want Hello", el["text"])
	}
	if el["fontSize"] != 14 {
		t.Errorf("fontSize = %v, want 14", el["fontSize"])
	}
	if el["strokeColor"] != "#ff0000" {
		t.Errorf("strokeColor = %v, want #ff0000", el["strokeColor"])
	}
	if _, ok := el["fontStyle"]; ok {
		t.Error("fontStyle should not be set when bold=false")
	}
	if el["x"] != float64(10) {
		t.Errorf("x = %v, want 10", el["x"])
	}
	if el["y"] != float64(20) {
		t.Errorf("y = %v, want 20", el["y"])
	}
	if el["width"] != float64(100) {
		t.Errorf("width = %v, want 100", el["width"])
	}
	if el["height"] != float64(30) {
		t.Errorf("height = %v, want 30", el["height"])
	}
	if el["id"] != "txt-1" {
		t.Errorf("id = %v, want txt-1", el["id"])
	}
}

func TestMakeText_BoldFlag(t *testing.T) {
	el := MakeText("txt-2", 0, 0, 50, 20, "Bold", 12, "#000000", true, 1)
	if el["fontStyle"] != "bold" {
		t.Errorf("fontStyle = %v, want bold", el["fontStyle"])
	}
}

func TestMakeText_CommonFields(t *testing.T) {
	el := MakeText("t", 0, 0, 10, 10, "", 10, "", false, 0)
	if el["opacity"] != 100 {
		t.Errorf("opacity = %v, want 100", el["opacity"])
	}
	if el["isDeleted"] != false {
		t.Errorf("isDeleted should be false")
	}
	if el["locked"] != false {
		t.Errorf("locked should be false")
	}
}

func TestMakeImage_Fields(t *testing.T) {
	el := MakeImage("img-1", 5, 15, 48, 48, "file-abc", "#ffffff", 7)

	if el["type"] != "image" {
		t.Errorf("type = %v, want image", el["type"])
	}
	if el["fileId"] != "file-abc" {
		t.Errorf("fileId = %v, want file-abc", el["fileId"])
	}
	if el["backgroundColor"] != "#ffffff" {
		t.Errorf("backgroundColor = %v, want #ffffff", el["backgroundColor"])
	}
	if el["status"] != "saved" {
		t.Errorf("status = %v, want saved", el["status"])
	}
	if el["id"] != "img-1" {
		t.Errorf("id = %v, want img-1", el["id"])
	}
	if el["x"] != float64(5) {
		t.Errorf("x = %v, want 5", el["x"])
	}
	if el["y"] != float64(15) {
		t.Errorf("y = %v, want 15", el["y"])
	}
}

func TestMakeImage_CommonFields(t *testing.T) {
	el := MakeImage("i", 0, 0, 10, 10, "", "", 0)
	if el["opacity"] != 100 {
		t.Errorf("opacity = %v, want 100", el["opacity"])
	}
	if el["isDeleted"] != false {
		t.Errorf("isDeleted should be false")
	}
}
