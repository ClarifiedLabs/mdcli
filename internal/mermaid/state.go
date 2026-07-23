package mermaid

import (
	"regexp"
	"strings"
)

var (
	reStateTrans = regexp.MustCompile(`^(\[\*\]|[\w.-]+)\s*-->\s*(\[\*\]|[\w.-]+)\s*(?::\s*(.*))?$`)
	reStateAs    = regexp.MustCompile(`^state\s+"([^"]*)"\s+as\s+([\w.-]+)\s*(\{)?\s*$`)
	reStateAnn   = regexp.MustCompile(`^state\s+([\w.-]+)\s+<<(\w+)>>\s*$`)
	reStateBlock = regexp.MustCompile(`^state\s+([\w.-]+)\s*\{\s*$`)
	reStateDesc  = regexp.MustCompile(`^([\w.-]+)\s*:\s*(.*)$`)
	reStateNote  = regexp.MustCompile(`(?i)^note\s+(right|left)\s+of\s+[\w.-]+\s*(:.*)?$`)
)

// stateScope is one level of composite-state nesting. `region` counts the
// concurrent regions (separated by `--`) seen so far at this level.
type stateScope struct {
	name   string
	region int
}

// scopeKey identifies the region a `[*]` belongs to, so that each composite
// state keeps its own start and end pseudo-states instead of sharing one
// global pair.
func scopeKey(scope []stateScope) string {
	var b strings.Builder
	for i, s := range scope {
		if i > 0 {
			b.WriteByte('/')
		}
		b.WriteString(s.name)
		if s.region > 0 {
			b.WriteByte('#')
			b.WriteString(itoa(s.region))
		}
	}
	return b.String()
}

// parseState parses stateDiagram / stateDiagram-v2 sources. Composite states
// are flattened: their inner states are rendered as ordinary states, each
// composite keeping its own start/end markers.
func parseState(lines []string) (*graph, error) {
	g := newGraph()
	first := true
	inNote := false
	var scope []stateScope
	stNode := func(id string) *gnode {
		switch id {
		case "[*]":
			panic("unreachable")
		}
		return g.node(id)
	}
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if first {
			first = false
			continue // stateDiagram / stateDiagram-v2 header
		}
		if inNote {
			if strings.EqualFold(line, "end note") {
				inNote = false
			}
			continue
		}
		lower := strings.ToLower(line)
		switch {
		case line == "}":
			if len(scope) > 0 {
				scope = scope[:len(scope)-1]
			}
			continue
		case line == "--":
			// concurrent region divider: a new region of the same composite
			if len(scope) > 0 {
				scope[len(scope)-1].region++
			}
			continue
		case strings.HasPrefix(lower, "direction "):
			g.dir = normalizeDir(line[len("direction "):])
			continue
		case reStateNote.MatchString(line):
			if !strings.Contains(line, ":") {
				inNote = true
			}
			continue
		case strings.HasPrefix(lower, "classdef "),
			strings.HasPrefix(lower, "class "),
			strings.HasPrefix(lower, "style "),
			strings.HasPrefix(lower, "accdescr"),
			strings.HasPrefix(lower, "acctitle"):
			continue
		}
		if m := reStateAs.FindStringSubmatch(line); m != nil {
			n := stNode(m[2])
			n.lines = splitLabel(m[1])
			if m[3] == "{" {
				scope = append(scope, stateScope{name: m[2]})
			}
			continue
		}
		if m := reStateAnn.FindStringSubmatch(line); m != nil {
			n := stNode(m[1])
			switch strings.ToLower(m[2]) {
			case "choice":
				n.kind = boxDiamond
			case "fork", "join":
				n.kind = boxBar
				n.lines = nil
			}
			continue
		}
		if m := reStateBlock.FindStringSubmatch(line); m != nil {
			// composite state: flatten (children parsed as ordinary states)
			stNode(m[1])
			scope = append(scope, stateScope{name: m[1]})
			continue
		}
		if m := reStateTrans.FindStringSubmatch(line); m != nil {
			key := scopeKey(scope)
			from := stateEndpoint(g, m[1], true, key)
			to := stateEndpoint(g, m[2], false, key)
			g.addEdge(&gedge{from: from, to: to, label: strings.TrimSpace(m[3]), em: mArrow})
			continue
		}
		if m := reStateDesc.FindStringSubmatch(line); m != nil {
			n := stNode(m[1])
			desc := splitLabel(m[2])
			if n.lines[0] == n.id && len(n.lines) == 1 {
				n.lines = desc
			} else {
				n.lines = append(n.lines, desc...)
			}
			continue
		}
		if reIdent.MatchString(line) {
			stNode(line)
			continue
		}
		if strings.HasPrefix(lower, "state ") {
			// e.g. `state X` without braces
			id := strings.TrimSpace(line[len("state "):])
			if reIdent.MatchString(id) {
				stNode(id)
			}
			continue
		}
	}
	// states default to rounded boxes
	for _, n := range g.nodes {
		if n.kind == boxRect {
			n.kind = boxRound
		}
	}
	return g, nil
}

var reIdent = regexp.MustCompile(`^[\w.-]+$`)

// stateEndpoint resolves `[*]` to the start or end pseudo-state of the region
// it appears in, so nested composites do not all share one marker.
func stateEndpoint(g *graph, id string, isSource bool, key string) *gnode {
	if id != "[*]" {
		return g.node(id)
	}
	base, label := "__end", "(O)"
	if isSource {
		base, label = "__start", "(*)"
	}
	if key != "" {
		base += ":" + key
	}
	n := g.node(base)
	n.kind = boxBare
	n.lines = []string{label}
	return n
}
