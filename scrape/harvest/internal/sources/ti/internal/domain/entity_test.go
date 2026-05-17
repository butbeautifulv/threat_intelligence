package domain

import (
	"testing"

	pkgdomain "github.com/butbeautifulv/veil/pkg/ti/domain"
)

func TestDomain_aliasesPkgTI(t *testing.T) {
	ioc := IOC{Type: IOCIP, Value: "203.0.113.5"}
	pkg := pkgdomain.IOC{Type: pkgdomain.IOCIP, Value: "203.0.113.5"}
	if ioc.NodeID() != pkg.NodeID() {
		t.Fatal("alias IOC should share pkg NodeID")
	}
}
