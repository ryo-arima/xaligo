package excalidraw_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ryo-arima/xaligo/internal/excalidraw"
	"github.com/ryo-arima/xaligo/internal/layout"
	"github.com/ryo-arima/xaligo/internal/model"
	"github.com/ryo-arima/xaligo/internal/parser"
)

func buildBoxFromXAL(t *testing.T, xal string) *layout.Box {
	t.Helper()
	doc, err := parser.Parse(strings.NewReader(xal))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	box, err := layout.Build(doc)
	if err != nil {
		t.Fatalf("build layout: %v", err)
	}
	return box
}

func TestBuildJSON_NilRoot_Error(t *testing.T) {
	_, err := excalidraw.BuildJSON(nil, "", "", "", 48, nil)
	if err == nil {
		t.Error("expected error for nil root")
	}
}

func TestBuildJSON_SimpleFrame_ValidJSON(t *testing.T) {
	box := buildBoxFromXAL(t, `<frame width="800" height="600"><card title="A" /></frame>`)
	data, err := excalidraw.BuildJSON(box, "", "", "", 48, nil)
	if err != nil {
		t.Fatalf("BuildJSON: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

func TestBuildJSON_OutputType(t *testing.T) {
	box := buildBoxFromXAL(t, `<frame width="800" height="600"></frame>`)
	data, err := excalidraw.BuildJSON(box, "", "", "", 48, nil)
	if err != nil {
		t.Fatalf("BuildJSON: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out["type"] != "excalidraw" {
		t.Errorf("type = %v, want %q", out["type"], "excalidraw")
	}
}

func TestBuildJSON_ContainsElements(t *testing.T) {
	box := buildBoxFromXAL(t, `<frame width="800" height="600"><card title="A" /></frame>`)
	data, err := excalidraw.BuildJSON(box, "", "", "", 48, nil)
	if err != nil {
		t.Fatalf("BuildJSON: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	elements, ok := out["elements"].([]interface{})
	if !ok {
		t.Fatal("elements field missing or wrong type")
	}
	if len(elements) == 0 {
		t.Error("expected at least one element")
	}
}

func TestBuildJSON_WithConnections(t *testing.T) {
	box := buildBoxFromXAL(t, `<frame width="800" height="600">
		<card title="A" />
	</frame>`)
	connNode := &model.Node{Tag: "connection", Attrs: map[string]string{"src": "1", "dst": "2"}}
	data, err := excalidraw.BuildJSON(box, "", "", "", 48, []*model.Node{connNode})
	if err != nil {
		t.Fatalf("BuildJSON with connections: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}
