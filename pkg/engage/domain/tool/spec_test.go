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
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties: %T", schema["properties"])
	}
	scan, ok := props["scan_type"].(map[string]any)
	if !ok || scan["default"] != "-sV" {
		t.Fatalf("scan_type prop: %v", props["scan_type"])
	}
	target, ok := props["target"].(map[string]any)
	if !ok || target["type"] != "string" {
		t.Fatalf("target prop: %v", props["target"])
	}
}

func TestInputSchema_emptyParameters_fallback(t *testing.T) {
	schema := (Spec{}).InputSchema()
	req, ok := schema["required"].([]string)
	if !ok || len(req) != 1 || req[0] != "target" {
		t.Fatalf("required: %v", schema["required"])
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("missing properties")
	}
	if _, ok := props["target"]; !ok {
		t.Fatal("missing target")
	}
	if _, ok := props["additional_args"]; !ok {
		t.Fatal("missing additional_args")
	}
}

func TestInputSchema_emptyType_defaultsString(t *testing.T) {
	schema := (Spec{Parameters: []Param{{Name: "port", Required: true}}}).InputSchema()
	props := schema["properties"].(map[string]any)
	port := props["port"].(map[string]any)
	if port["type"] != "string" {
		t.Fatalf("type: %v", port["type"])
	}
}

func TestInputSchema_noRequired_omitsKey(t *testing.T) {
	schema := (Spec{Parameters: []Param{{Name: "opt", Type: "string"}}}).InputSchema()
	if _, ok := schema["required"]; ok {
		t.Fatalf("unexpected required: %v", schema["required"])
	}
}

func TestDefaultParameters(t *testing.T) {
	tests := []struct {
		name string
		spec Spec
		want map[string]string
	}{
		{
			name: "catalog defaults",
			spec: Spec{Parameters: []Param{
				{Name: "threads", Default: "10"},
				{Name: "target", Required: true},
			}},
			want: map[string]string{"threads": "10"},
		},
		{
			name: "empty defaults",
			spec: Spec{Parameters: []Param{{Name: "target", Required: true}}},
			want: map[string]string{},
		},
		{
			name: "no parameters",
			spec: Spec{},
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.DefaultParameters()
			if len(got) != len(tt.want) {
				t.Fatalf("got %v want %v", got, tt.want)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Fatalf("got %v want %v", got, tt.want)
				}
			}
		})
	}
}
