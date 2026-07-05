package capability

import "testing"

func TestScanDeclarations(t *testing.T) {
	src := "AGENT x\nCAPABILITY resume cli\nCAPABILITY streaming true\n"
	caps := ScanDeclarations(src)
	if caps["resume"] != "cli" || caps["streaming"] != "true" {
		t.Fatalf("unexpected caps: %v", caps)
	}
	if ScanDeclarations("") != nil {
		t.Fatal("empty source should return nil")
	}
}
