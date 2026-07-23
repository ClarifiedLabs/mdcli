package pager

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestSplitCommand(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   \t  ", nil},
		{"less", []string{"less"}},
		{"less -S", []string{"less", "-S"}},
		{"  less   -S  --chop-long-lines ", []string{"less", "-S", "--chop-long-lines"}},
		{"/usr/local/bin/most", []string{"/usr/local/bin/most"}},
	}
	for _, c := range cases {
		got := splitCommand(c.in)
		if len(got) != len(c.want) {
			t.Errorf("splitCommand(%q) = %v, want %v", c.in, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitCommand(%q) = %v, want %v", c.in, got, c.want)
				break
			}
		}
	}
}

func TestLessArgs(t *testing.T) {
	t.Parallel()
	frx := []string{"-F", "-R", "-X"}
	cases := []struct {
		name    string
		base    string
		lessEnv string
		want    []string
	}{
		{"plain less", "less", "", frx},
		{"absolute less", "/usr/bin/less", "", frx},
		{"LESS set is respected", "less", "-S", nil},
		{"LESS set non-empty even to -FRX", "less", "-FRX", nil},
		{"more never gets flags", "more", "", nil},
		{"absolute more never gets flags", "/usr/bin/more", "", nil},
		{"custom pager never gets flags", "most", "", nil},
		{"custom pager with args never gets flags", "/usr/local/bin/most", "", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := lessArgs(c.base, c.lessEnv); !reflect.DeepEqual(got, c.want) {
				t.Errorf("lessArgs(%q, %q) = %v, want %v", c.base, c.lessEnv, got, c.want)
			}
		})
	}
}

// stubLookPath returns a LookPath func that resolves only the given names.
func stubLookPath(found ...string) func(string) (string, error) {
	set := map[string]bool{}
	for _, n := range found {
		set[n] = true
	}
	return func(name string) (string, error) {
		if set[name] {
			return "/usr/bin/" + name, nil
		}
		return "", errors.New("not found: " + name)
	}
}

func TestResolve(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		pagerEnv string
		found    []string
		wantArgv []string
		wantLess bool
		wantOK   bool
	}{
		{"PAGER verbatim no args", "most", []string{"most", "less", "more"}, []string{"most"}, false, true},
		{"PAGER with args", "less -S", []string{"less", "more"}, []string{"less", "-S"}, true, true},
		{"PAGER wins over less", "most", []string{"most", "less", "more"}, []string{"most"}, false, true},
		{"PAGER absolute path", "/usr/local/bin/most", []string{"/usr/local/bin/most", "less"}, []string{"/usr/local/bin/most"}, false, true},
		{"PAGER base missing falls to less", "nonexistent", []string{"less", "more"}, []string{"less"}, true, true},
		{"PAGER missing falls to more when no less", "nonexistent", []string{"more"}, []string{"more"}, false, true},
		{"PAGER whitespace only falls through", "   ", []string{"less"}, []string{"less"}, true, true},
		{"no PAGER prefers less", "", []string{"less", "more"}, []string{"less"}, true, true},
		{"no PAGER no less uses more", "", []string{"more"}, []string{"more"}, false, true},
		{"none resolve", "", []string{}, nil, false, false},
		{"PAGER set but nothing found", "nonexistent", []string{}, nil, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			argv, isLess, ok := Resolve(c.pagerEnv, stubLookPath(c.found...))
			if ok != c.wantOK {
				t.Fatalf("Resolve(%q) ok = %v, want %v", c.pagerEnv, ok, c.wantOK)
			}
			if isLess != c.wantLess {
				t.Errorf("Resolve(%q) isLess = %v, want %v", c.pagerEnv, isLess, c.wantLess)
			}
			if !reflect.DeepEqual(argv, c.wantArgv) {
				t.Errorf("Resolve(%q) argv = %v, want %v", c.pagerEnv, argv, c.wantArgv)
			}
		})
	}
}

func TestCommand(t *testing.T) {
	t.Parallel()

	// less with LESS unset gets -FRX appended and LESS=-FRX in the child env
	// (overriding any inherited LESS).
	cmd := Command([]string{"less"}, "")
	wantArgs := []string{"less", "-F", "-R", "-X"}
	if !reflect.DeepEqual(cmd.Args, wantArgs) {
		t.Errorf("less Command args = %v, want %v", cmd.Args, wantArgs)
	}
	if got := envValue(cmd.Env, "LESS"); got != "-F-R-X" {
		t.Errorf("less Command LESS = %q, want %q", got, "-F-R-X")
	}

	// less with LESS set is run verbatim and does not override the env.
	cmd = Command([]string{"less", "-S"}, "-S")
	wantArgs = []string{"less", "-S"}
	if !reflect.DeepEqual(cmd.Args, wantArgs) {
		t.Errorf("less with LESS set args = %v, want %v", cmd.Args, wantArgs)
	}
	if cmd.Env != nil {
		t.Errorf("less with LESS set should leave env alone, got %v", cmd.Env)
	}

	// more is run verbatim with no flags and no env override.
	cmd = Command([]string{"more"}, "")
	wantArgs = []string{"more"}
	if !reflect.DeepEqual(cmd.Args, wantArgs) {
		t.Errorf("more Command args = %v, want %v", cmd.Args, wantArgs)
	}
	if cmd.Env != nil {
		t.Errorf("more should leave env alone, got %v", cmd.Env)
	}

	// a custom $PAGER with args is run verbatim.
	cmd = Command([]string{"most", "-w"}, "")
	wantArgs = []string{"most", "-w"}
	if !reflect.DeepEqual(cmd.Args, wantArgs) {
		t.Errorf("custom pager Command args = %v, want %v", cmd.Args, wantArgs)
	}
	if cmd.Env != nil {
		t.Errorf("custom pager should leave env alone, got %v", cmd.Env)
	}
}

// envValue returns the value of the last key= entry in env, or "" if absent.
func envValue(env []string, key string) string {
	val := ""
	for _, e := range env {
		if strings.HasPrefix(e, key+"=") {
			val = e[len(key)+1:]
		}
	}
	return val
}
