package rules

import "testing"

func TestAllEscapedArgs(t *testing.T) {
	consts := map[string]bool{"openTag": true}
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{"rune literal", []string{"'+')"}, true},
		{"local constant", []string{"openTag)"}, true},
		{"escaper call with closing paren", []string{"html.EscapeString(p))"}, true},
		{"escaper call balanced", []string{"html.EscapeString(p)"}, true},
		{"raw variable", []string{"p)"}, false},
		{"unknown identifier", []string{"closeTag)"}, false},
	}
	for _, c := range cases {
		if got := allEscapedArgs(c.args, consts); got != c.want {
			t.Errorf("%s: allEscapedArgs(%q) = %v, want %v", c.name, c.args, got, c.want)
		}
	}
}
