package repository_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ryo-arima/xaligo/internal/repository"
)

func TestReadServiceList_SingleColumn(t *testing.T) {
	content := "Amazon S3\nAmazon EC2\n"
	f := writeTempFile(t, content)

	entries, err := repository.ReadServiceList(f)
	if err != nil {
		t.Fatalf("ReadServiceList: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].OfficialName != "Amazon S3" {
		t.Errorf("entries[0].OfficialName = %q, want %q", entries[0].OfficialName, "Amazon S3")
	}
}

func TestReadServiceList_TwoColumnsWithID(t *testing.T) {
	content := "10,Amazon S3\n20,Amazon EC2\n"
	f := writeTempFile(t, content)

	entries, err := repository.ReadServiceList(f)
	if err != nil {
		t.Fatalf("ReadServiceList: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].CatalogID != 10 {
		t.Errorf("entries[0].CatalogID = %d, want 10", entries[0].CatalogID)
	}
	if entries[0].OfficialName != "Amazon S3" {
		t.Errorf("entries[0].OfficialName = %q, want %q", entries[0].OfficialName, "Amazon S3")
	}
}

func TestReadServiceList_SkipsComments(t *testing.T) {
	content := "# this is a comment\nAmazon S3\n"
	f := writeTempFile(t, content)

	entries, err := repository.ReadServiceList(f)
	if err != nil {
		t.Fatalf("ReadServiceList: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestReadServiceList_SkipsEmptyLines(t *testing.T) {
	content := "Amazon S3\n\n\nAmazon EC2\n"
	f := writeTempFile(t, content)

	entries, err := repository.ReadServiceList(f)
	if err != nil {
		t.Fatalf("ReadServiceList: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestReadServiceList_ThreeColumnsWithAbbreviation(t *testing.T) {
	content := "100,Amazon Elastic Compute Cloud,EC2\n"
	f := writeTempFile(t, content)

	entries, err := repository.ReadServiceList(f)
	if err != nil {
		t.Fatalf("ReadServiceList: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Abbreviation != "EC2" {
		t.Errorf("Abbreviation = %q, want %q", entries[0].Abbreviation, "EC2")
	}
}

func TestReadServiceList_NonExistentFile_Error(t *testing.T) {
	_, err := repository.ReadServiceList("/nonexistent/services.csv")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "services.csv")
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return f
}
