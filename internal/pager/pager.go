// Package pager renders content through an external pager subprocess
// ($PAGER -> less -> more) so interactive output behaves like less:
// scroll, search, quit with q. It uses only the standard library; the pager
// is a child process, not a Go module dependency.
package pager

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// splitCommand splits a pager command line into fields. It is a thin wrapper
// around strings.Fields so PAGER="less -S" becomes ["less", "-S"]. An empty
// or all-whitespace string yields nil.
func splitCommand(s string) []string {
	return strings.Fields(s)
}

// lessArgs returns the flags to make less feel like a modern pager: -F
// (quit if one screen), -R (keep ANSI color), -X (no alternate screen). It
// returns them only when the resolved pager basename is exactly "less" and
// the LESS environment variable is unset; otherwise it returns nil so the
// pager (more, or a custom $PAGER) is run verbatim and a user's LESS is
// respected entirely.
func lessArgs(base, lessEnv string) []string {
	if filepath.Base(base) != "less" || lessEnv != "" {
		return nil
	}
	return []string{"-F", "-R", "-X"}
}

// Resolve picks a pager using the chain $PAGER -> less -> more. pagerEnv is
// the value of the PAGER environment variable (may contain arguments);
// lookPath locates executables (normally exec.LookPath, stubbed in tests).
// It returns the pager argv, whether the resolved base command is less, and
// whether any pager resolved at all. A $PAGER whose base is not found falls
// through to less, then more, rather than failing.
func Resolve(pagerEnv string, lookPath func(string) (string, error)) (argv []string, isLess bool, ok bool) {
	if pagerEnv != "" {
		if fields := splitCommand(pagerEnv); len(fields) > 0 {
			if _, err := lookPath(fields[0]); err == nil {
				return fields, filepath.Base(fields[0]) == "less", true
			}
		}
	}
	if _, err := lookPath("less"); err == nil {
		return []string{"less"}, true, true
	}
	if _, err := lookPath("more"); err == nil {
		return []string{"more"}, false, true
	}
	return nil, false, false
}

// setEnv returns env with the given key=value applied, replacing any existing
// entry for key so the result is deterministic regardless of the parent env.
func setEnv(env []string, kv string) []string {
	key := kv[:strings.IndexByte(kv, '=')+1]
	out := make([]string, 0, len(env)+1)
	for _, e := range env {
		if !strings.HasPrefix(e, key) {
			out = append(out, e)
		}
	}
	return append(out, kv)
}

// Command builds the *exec.Cmd for the resolved pager argv. When the pager
// is less and LESS is unset, the -FRX flags are appended and the child's
// environment gets LESS=-FRX (overriding any inherited value; our own
// environment is left untouched). The command's stdout and stderr are wired
// to ours; stdin is set by the caller.
func Command(argv []string, lessEnv string) *exec.Cmd {
	args := lessArgs(argv[0], lessEnv)
	full := append(append([]string{}, argv...), args...)
	cmd := exec.Command(full[0], full[1:]...)
	if len(args) > 0 {
		cmd.Env = setEnv(os.Environ(), "LESS="+strings.Join(args, ""))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// Page sends content through the resolved pager. If no pager resolves, or
// the pager fails to start, content is written to fallback (os.Stdout) so
// nothing is ever lost and nil is returned. A wrapped error is returned only
// for a genuine pager run failure; callers may log it but should not treat
// it as fatal.
func Page(content []byte, pagerEnv string, fallback io.Writer) error {
	argv, _, ok := Resolve(pagerEnv, exec.LookPath)
	if !ok {
		_, err := fallback.Write(content)
		return err
	}
	cmd := Command(argv, os.Getenv("LESS"))
	cmd.Stdin = bytes.NewReader(content)
	if err := cmd.Run(); err != nil {
		// Never lose content: fall back to a direct write.
		if _, werr := fallback.Write(content); werr != nil {
			return werr
		}
		return fmt.Errorf("pager: %w", err)
	}
	return nil
}
