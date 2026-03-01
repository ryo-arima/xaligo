package entity_test

import (
	"testing"

	"github.com/ryo-arima/xaligo/internal/entity"
)

func TestShortLabel_WithAbbreviation(t *testing.T) {
	svc := entity.ServiceEntry{OfficialName: "Amazon Elastic Compute Cloud", Abbreviation: "EC2"}
	if got := svc.ShortLabel(); got != "EC2" {
		t.Errorf("ShortLabel() = %q, want %q", got, "EC2")
	}
}

func TestShortLabel_StripAmazonPrefix(t *testing.T) {
	svc := entity.ServiceEntry{OfficialName: "Amazon S3"}
	if got := svc.ShortLabel(); got != "S3" {
		t.Errorf("ShortLabel() = %q, want %q", got, "S3")
	}
}

func TestShortLabel_StripAWSPrefix(t *testing.T) {
	svc := entity.ServiceEntry{OfficialName: "AWS Lambda"}
	if got := svc.ShortLabel(); got != "Lambda" {
		t.Errorf("ShortLabel() = %q, want %q", got, "Lambda")
	}
}

func TestShortLabel_NoPrefix(t *testing.T) {
	svc := entity.ServiceEntry{OfficialName: "Elastic Load Balancing"}
	if got := svc.ShortLabel(); got != "Elastic Load Balancing" {
		t.Errorf("ShortLabel() = %q, want %q", got, "Elastic Load Balancing")
	}
}

func TestShortLabel_EmptyName(t *testing.T) {
	svc := entity.ServiceEntry{}
	if got := svc.ShortLabel(); got != "" {
		t.Errorf("ShortLabel() = %q, want empty string", got)
	}
}
