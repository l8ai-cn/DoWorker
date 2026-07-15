package loopscript

type parser struct {
	tokens     []token
	pos        int
	diagnostic *Diagnostic
	positions  programPositions
}

func Parse(source string) (*Program, []Diagnostic) {
	tokens := lex(source)
	for _, item := range tokens {
		if item.kind == tokenIllegal {
			return nil, []Diagnostic{newDiagnostic(
				"loop.syntax.invalid-token", item.literal, "", positionOf(item),
			)}
		}
	}

	p := &parser{tokens: tokens}
	program := p.parseProgram()
	if p.diagnostic != nil {
		return nil, []Diagnostic{*p.diagnostic}
	}
	diagnostics := validateProgram(program, &p.positions)
	if len(diagnostics) != 0 {
		return nil, diagnostics
	}
	return program, nil
}

func (p *parser) parseProgram() *Program {
	nodeID, nodePosition := p.parseNodeID()
	if p.failed() {
		return nil
	}
	p.positions.loop = nodePosition
	p.expect(tokenLoop)
	localID := p.takeIdentifier()
	p.expect(tokenLBrace)

	loop := LoopNode{NodeID: nodeID, LocalID: localID}
	var hasLimits, hasRepeat, hasFailure bool
	for !p.failed() && p.current().kind != tokenRBrace && p.current().kind != tokenEOF {
		switch p.current().kind {
		case tokenAt:
			childID, childPosition := p.parseNodeID()
			switch p.current().kind {
			case tokenRepeat:
				if hasRepeat {
					p.failCurrent("loop.structure.repeat-count", "exactly one repeat is allowed", childID)
					break
				}
				p.positions.repeat = childPosition
				loop.Repeat = p.parseRepeat(childID)
				hasRepeat = true
			default:
				p.failCurrent("loop.syntax.unknown", "unknown attributed loop node", childID)
			}
		case tokenLimits:
			if hasLimits {
				p.failCurrent("loop.structure.limits-count", "exactly one limits declaration is allowed", "")
				break
			}
			p.positions.limits = positionOf(p.current())
			loop.Limits = p.parseLimits()
			hasLimits = true
		case tokenOnFailure:
			if hasFailure {
				p.failCurrent("loop.structure.failure-count", "exactly one failure policy is allowed", "")
				break
			}
			p.positions.failure = positionOf(p.current())
			loop.FailurePolicy = p.parseFailurePolicy()
			hasFailure = true
		case tokenRepeat:
			p.failCurrent("loop.node-id.missing", "node must be preceded by @id(...)", "")
		default:
			p.failCurrent("loop.syntax.unknown", "unknown loop syntax", "")
		}
	}
	p.expect(tokenRBrace)
	p.expect(tokenEOF)
	if p.failed() {
		return nil
	}
	if !hasLimits || !hasRepeat || !hasFailure {
		p.failMissingStructure(hasLimits, hasRepeat, hasFailure)
		return nil
	}
	return &Program{SchemaVersion: 1, Loop: loop}
}

func (p *parser) parseNodeID() (string, sourcePosition) {
	start := positionOf(p.current())
	if p.current().kind != tokenAt {
		p.failCurrent("loop.node-id.missing", "node must be preceded by @id(...)", "")
		return "", start
	}
	p.advance()
	p.expect(tokenID)
	p.expect(tokenLParen)
	nodeID := p.takeIdentifier()
	p.expect(tokenRParen)
	return nodeID, start
}
