package finding

import "testing"

func valid() Input {
	return Input{
		Title: "Missing object-level authorization", Class: AuthZ,
		Path: "src/api/orders.ts", Line: 36, CWE: "CWE-639",
		OWASP: LabelsFor(AuthZ), Severity: High,
	}
}

func TestNewAcceptsEvidenceBackedInput(t *testing.T) {
	f, ok := New(valid())
	if !ok || f.Action != Block || f.Path != "src/api/orders.ts" || f.Line != 36 {
		t.Fatalf("New(valid) = %+v, %v", f, ok)
	}
}

func TestNewFailsClosed(t *testing.T) {
	cases := map[string]func(*Input){
		"no path":           func(in *Input) { in.Path = "" },
		"zero line":         func(in *Input) { in.Line = 0 },
		"bad cwe":           func(in *Input) { in.CWE = "CWE-XYZ" },
		"unknown class":     func(in *Input) { in.Class = "made-up" },
		"no owasp":          func(in *Input) { in.OWASP = nil },
		"invented owasp id": func(in *Input) { in.OWASP = []string{"A01:2021", "A11:2021"} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			in := valid()
			mutate(&in)
			if _, ok := New(in); ok {
				t.Fatalf("New accepted %s", name)
			}
		})
	}
}

func TestActionBySeverity(t *testing.T) {
	for sev, want := range map[Severity]Action{Critical: Block, High: Block, Medium: Review, Low: Review} {
		in := valid()
		in.Severity = sev
		if f, _ := New(in); f.Action != want {
			t.Errorf("%s -> %s, want %s", sev, f.Action, want)
		}
	}
}

func TestNewCopiesOWASP(t *testing.T) {
	in := valid()
	f, _ := New(in)
	in.OWASP[0] = "tampered"
	if f.OWASP[0] != "A01:2021" {
		t.Fatal("finding shares OWASP slice with input")
	}
}

func TestInjectionLabelsAreTheVerifiedIDs(t *testing.T) {
	want := map[Class][]string{
		SQLInjection:  {"A03:2021", "ASVS-5.0-1.2.4"},
		CmdInjection:  {"A03:2021", "ASVS-5.0-1.2.5"},
		PathTraversal: {"A01:2021", "ASVS-5.0-5.3.2"},
	}
	for c, ids := range want {
		got := LabelsFor(c)
		if len(got) != len(ids) || got[0] != ids[0] || got[1] != ids[1] {
			t.Errorf("LabelsFor(%s) = %v, want %v", c, got, ids)
		}
	}
}

func TestLabelsAreOfficial(t *testing.T) {
	for c := range classLabel {
		if LabelsFor(c) == nil {
			t.Errorf("class %s has no official labels", c)
		}
	}
}
