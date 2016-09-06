package font

import "testing"

func TestAll(t *testing.T) {
	if builtin == nil {
		t.Fatal("no builtin fonts are loaded")
	}
	t.Logf("%d builtin fonts loaded", builtin.Len())
}
