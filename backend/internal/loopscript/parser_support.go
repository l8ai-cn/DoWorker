package loopscript

import (
	"fmt"
	"strconv"
	"strings"
)

func (p *parser) failMissingStructure(limits, repeat, failure bool) {
	switch {
	case !limits:
		p.failCurrent("loop.structure.limits-count", "exactly one limits declaration is required", "")
	case !repeat:
		p.failCurrent("loop.structure.repeat-count", "exactly one repeat is required", "")
	case !failure:
		p.failCurrent("loop.structure.failure-count", "exactly one failure policy is required", "")
	}
}

func (p *parser) takeIdentifier() string {
	item := p.current()
	if item.kind != tokenIdent {
		p.failCurrent("loop.syntax.unexpected-token", "expected identifier", "")
		return ""
	}
	p.advance()
	return item.literal
}

func (p *parser) takeInteger(kind tokenKind) int64 {
	item := p.current()
	if item.kind != kind {
		p.failCurrent("loop.syntax.unexpected-token", fmt.Sprintf("expected %s", kind), "")
		return 0
	}
	p.advance()
	raw := strings.TrimSuffix(item.literal, "m")
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		p.failAt("loop.syntax.invalid-number", "invalid integer", "", positionOf(item))
		return 0
	}
	return value
}

func (p *parser) expect(kind tokenKind) {
	if p.failed() {
		return
	}
	if p.current().kind != kind {
		p.failCurrent(
			"loop.syntax.unexpected-token",
			fmt.Sprintf("expected %s, got %s", kind, p.current().kind),
			"",
		)
		return
	}
	p.advance()
}

func (p *parser) current() token {
	if p.pos >= len(p.tokens) {
		if len(p.tokens) > 0 {
			return p.tokens[len(p.tokens)-1]
		}
		return token{kind: tokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *parser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func (p *parser) failed() bool {
	return p.diagnostic != nil
}

func (p *parser) failCurrent(code, message, nodeID string) {
	p.failAt(code, message, nodeID, positionOf(p.current()))
}

func (p *parser) failAt(code, message, nodeID string, position sourcePosition) {
	if p.diagnostic == nil {
		diagnostic := newDiagnostic(code, message, nodeID, position)
		p.diagnostic = &diagnostic
	}
}

func positionOf(item token) sourcePosition {
	return sourcePosition{line: item.line, column: item.column}
}

func newDiagnostic(code, message, nodeID string, position sourcePosition) Diagnostic {
	return Diagnostic{
		Code: code, Message: message, NodeID: nodeID,
		Line: position.line, Column: position.column,
	}
}
