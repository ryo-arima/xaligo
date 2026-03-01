package repository

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestFileID(t *testing.T) {
	id1 := FileID("some/path/file.svg")
	id2 := FileID("some/path/file.svg")
	if id1 != id2 {
		t.Errorf("FileID not deterministic: %q != %q", id1, id2)
	}
	if len(id1) != 16 {
		t.Errorf("FileID length = %d, want 16", len(id1))
	}
	id3 := FileID("other/path/file.svg")
	if id1 == id3 {
		t.Error("FileID should differ for different inputs")
	}
}

func TestSVGBGColorNoFill(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg"></svg>`
	dataURL := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
	color := SVGBGColor(dataURL)
	if color != "transparent" {
		t.Errorf("SVGBGColor = %q, want %q", color, "transparent")
	}
}

func TestSVGBGColorWithFill(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg"><rect fill="#FF9900" /></svg>`
	dataURL := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
	color := SVGBGColor(dataURL)
	if color != "#ff9900" {
		t.Errorf("SVGBGColor = %q, want %q", color, "#ff9900")
	}
}

func TestSVGBGColorSkipsWhite(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg"><rect fill="#ffffff" /><rect fill="#336699" /></svg>`
	dataURL := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
	color := SVGBGColor(dataURL)
	if color != "#336699" {
		t.Errorf("SVGBGColor = %q, want %q", color, "#336699")
	}
}

func TestSVGBGColorInvalidDataURL(t *testing.T) {
	color := SVGBGColor("not-a-data-url")
	if color != "transparent" {
		t.Errorf("SVGBGColor for invalid URL = %q, want %q", color, "transparent")
	}
}

func TestMakeTextFields(t *testing.T) {
	el := MakeText("txt-1", 10, 20, 100, 30, "Hello", 14, "#000000", false, 42)
	if el["type"] != "text" {
		t.Errorf("type = %v, want text", el["type"])
	}
	if el["text"] != "Hello" {
		t.Errorf("text = %v, want Hello", el["text"])
	}
	if el["fontSize"] != 14 {
		t.Errorf("fontSize = %v, want 14", el["fontSize"])
	}
	if el["strokeColor"] != "#000000" {
		t.Errorf("strokeColor = %v, want #000000", el["strokeColor"])
	}
}

func TestMakeTextBold(t *testing.T) {
	el := MakeText("txt-2", 0, 0, 50, 20, "Bold", 12, "#000000", true, 1)
	if el["fontStyle"] != "bold" {
		t.Errorf("fontStyle = %v, want bold", el["fontStyle"])
	}
}

func TestMakeTextNoBold(t *testing.T) {
	el := MakeText("txt-3", 0, 0, 50, 20, "Normal", 12, "#000000", false, 1)
	if _, ok := el["fontStyle"]; ok {
		t.Error("fontStyle should not be set for non-bold text")
	}
}

func TestMakeImageFields(t *testing.T) {
	el := MakeImage("img-1", 5, 10, 48, 48, "fileid123", "#ff0000", 7)
	if el["type"] != "image" {
		t.Errorf("type = %v, want image", el["type"])
	}
	if el["fileId"] != "fileid123" {
		t.Errorf("fileId = %v, want fileid123", el["fileId"])
	}
	if el["backgroundColor"] != "#ff0000" {
		t.Errorf("backgroundColor = %v, want #ff0000", el["backgroundColor"])
	}
}

func TestReadServiceListSingleColumn(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "list.csv")
	content := "# comment\nAmazon EC2\nAWS Lambda\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	entries, err := ReadServiceList(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].OfficialName != "Amazon EC2" {
		t.Errorf("entry[0].OfficialName = %q", entries[0].OfficialName)
	}
}

func TestReadServiceListTwoColumn(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "list.csv")
	content := "42,Amazon S3\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	entries, err := ReadServiceList(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].CatalogID != 42 {
		t.Errorf("CatalogID = %d, want 42", entries[0].CatalogID)
	}
	if entries[0].OfficialName != "Amazon S3" {
		t.Errorf("OfficialName = %q", entries[0].OfficialName)
	}
}

func TestReadServiceListThreeColumn(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "list.csv")
	content := "99,Amazon DynamoDB,DynamoDB\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	entries, err := ReadServiceList(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries[0].Abbreviation != "DynamoDB" {
		t.Errorf("Abbreviation = %q, want %q", entries[0].Abbreviation, "DynamoDB")
	}
}

func TestReadServiceListSkipsEmpty(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "list.csv")
	content := "\n\n# just comments\n\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	entries, err := ReadServiceList(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestReadServiceListNotFound(t *testing.T) {
	_, err := ReadServiceList("/nonexistent/path/list.csv")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
