package validate

import (
	"testing"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

func TestValidIOCType(t *testing.T) {
	for _, typ := range []domain.IOCType{domain.IOCIP, domain.IOCDomain, domain.IOCURL, domain.IOCHash} {
		if !ValidIOCType(typ) {
			t.Fatalf("%q should be valid", typ)
		}
	}
	if ValidIOCType(domain.IOCType("email")) {
		t.Fatal("unknown type")
	}
}

func TestCheckIOCShape(t *testing.T) {
	if err := CheckIOCShape(domain.IOC{Type: domain.IOCIP, Value: "1.2.3.4"}); err != nil {
		t.Fatal(err)
	}
	if err := CheckIOCShape(domain.IOC{Type: domain.IOCIP, Value: "  "}); err != ErrEmptyValue {
		t.Fatalf("got %v", err)
	}
	if err := CheckIOCShape(domain.IOC{Type: "x", Value: "a"}); err != ErrUnknownType {
		t.Fatalf("got %v", err)
	}
}

func TestCheckActorName_and_ReportLink(t *testing.T) {
	if CheckActorName("APT29") != nil {
		t.Fatal("actor")
	}
	if CheckActorName(" ") != ErrEmptyActor {
		t.Fatal("empty actor")
	}
	if CheckReportLink("https://x") != nil {
		t.Fatal("report")
	}
	if CheckReportLink("") != ErrEmptyReport {
		t.Fatal("empty report")
	}
}
