package tool

import "testing"

func TestInputSchema_withParameters(t *testing.T) {
	s := Spec{
		Parameters: []Param{
			{Name: "target", Type: "string", Required: true},
			{Name: "scan_type", Type: "string", Default: "-sV"},
		},
	}
	schema := s.InputSchema()
	req, ok := schema["required"].([]string)
	if !ok || len(req) != 1 || req[0] != "target" {
		t.Fatalf("required: %v", schema["required"])
	}
}
