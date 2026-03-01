package entity

import "testing"

func TestShortLabelAbbreviation(t *testing.T) {
	s := ServiceEntry{OfficialName: "Amazon EC2", Abbreviation: "EC2"}
	if got := s.ShortLabel(); got != "EC2" {
		t.Errorf("ShortLabel = %q, want %q", got, "EC2")
	}
}

func TestShortLabelStripsAmazon(t *testing.T) {
	s := ServiceEntry{OfficialName: "Amazon Simple Storage Service"}
	if got := s.ShortLabel(); got != "Simple Storage Service" {
		t.Errorf("ShortLabel = %q, want %q", got, "Simple Storage Service")
	}
}

func TestShortLabelStripsAWS(t *testing.T) {
	s := ServiceEntry{OfficialName: "AWS Lambda"}
	if got := s.ShortLabel(); got != "Lambda" {
		t.Errorf("ShortLabel = %q, want %q", got, "Lambda")
	}
}

func TestShortLabelNoPrefix(t *testing.T) {
	s := ServiceEntry{OfficialName: "Some Service"}
	if got := s.ShortLabel(); got != "Some Service" {
		t.Errorf("ShortLabel = %q, want %q", got, "Some Service")
	}
}

func TestShortLabelEmptyAbbreviationFallsBack(t *testing.T) {
	s := ServiceEntry{OfficialName: "Amazon DynamoDB", Abbreviation: ""}
	if got := s.ShortLabel(); got != "DynamoDB" {
		t.Errorf("ShortLabel = %q, want %q", got, "DynamoDB")
	}
}
