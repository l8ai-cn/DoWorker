package loopscript

func (p *parser) parseLimits() Limits {
	p.expect(tokenLimits)
	p.expect(tokenLParen)
	p.expect(tokenIterations)
	p.expect(tokenColon)
	iterations := p.takeInteger(tokenNumber)
	p.expect(tokenComma)
	p.expect(tokenTokens)
	p.expect(tokenColon)
	tokens := p.takeInteger(tokenNumber)
	p.expect(tokenComma)
	p.expect(tokenTimeout)
	p.expect(tokenColon)
	timeout := p.takeInteger(tokenDuration)
	p.expect(tokenComma)
	p.expect(tokenNoProgress)
	p.expect(tokenColon)
	noProgress := p.takeInteger(tokenNumber)
	p.expect(tokenComma)
	p.expect(tokenSameError)
	p.expect(tokenColon)
	sameError := p.takeInteger(tokenNumber)
	p.expect(tokenRParen)
	return Limits{
		Iterations: iterations, Tokens: tokens, TimeoutMins: timeout,
		NoProgress: noProgress, SameError: sameError,
	}
}

func (p *parser) parseRepeat(nodeID string) RepeatNode {
	p.expect(tokenRepeat)
	localID := p.takeIdentifier()
	p.expect(tokenLParen)
	p.expect(tokenMax)
	p.expect(tokenColon)
	maximum := p.takeInteger(tokenNumber)
	p.expect(tokenComma)
	p.expect(tokenUntil)
	p.expect(tokenColon)
	until := p.parseReference()
	p.expect(tokenRParen)
	p.expect(tokenLBrace)

	repeat := RepeatNode{NodeID: nodeID, LocalID: localID, Max: maximum, Until: until}
	var hasAgent, hasVerifier bool
	for !p.failed() && p.current().kind != tokenRBrace && p.current().kind != tokenEOF {
		if p.current().kind == tokenCustomBlock {
			if repeat.CustomBlock != nil {
				p.failCurrent("loop.custom-block.count", "at most one custom block reference is allowed", nodeID)
				break
			}
			if hasAgent || hasVerifier {
				p.failCurrent(
					"loop.custom-block.position",
					"custom block reference must precede the expanded nodes",
					nodeID,
				)
				break
			}
			p.positions.customBlock = positionOf(p.current())
			repeat.CustomBlock = p.parseCustomBlockRef()
			continue
		}
		if p.current().kind != tokenAt {
			switch p.current().kind {
			case tokenAgent, tokenVerify:
				p.failCurrent("loop.node-id.missing", "node must be preceded by @id(...)", "")
			default:
				p.failCurrent("loop.syntax.unknown", "unknown repeat syntax", "")
			}
			break
		}
		childID, childPosition := p.parseNodeID()
		switch p.current().kind {
		case tokenAgent:
			if hasAgent {
				p.failCurrent("loop.structure.agent-count", "exactly one agent is allowed", childID)
				break
			}
			p.positions.agent = childPosition
			repeat.Agent = p.parseAgent(childID)
			hasAgent = true
		case tokenVerify:
			if hasVerifier {
				p.failCurrent("loop.structure.verifier-count", "exactly one verifier is allowed", childID)
				break
			}
			p.positions.verifier = childPosition
			repeat.Verifier = p.parseVerifier(childID)
			hasVerifier = true
		default:
			p.failCurrent("loop.syntax.unknown", "unknown attributed repeat node", childID)
		}
	}
	p.expect(tokenRBrace)
	if !p.failed() && (!hasAgent || !hasVerifier) {
		if !hasAgent {
			p.failCurrent("loop.structure.agent-count", "exactly one agent is required", nodeID)
		} else {
			p.failCurrent("loop.structure.verifier-count", "exactly one verifier is required", nodeID)
		}
	}
	return repeat
}

func (p *parser) parseCustomBlockRef() *CustomBlockRef {
	p.expect(tokenCustomBlock)
	p.expect(tokenLParen)
	p.expect(tokenNodeID)
	p.expect(tokenColon)
	nodeID := p.takeIdentifier()
	p.expect(tokenComma)
	p.expect(tokenDefinitionID)
	p.expect(tokenColon)
	definitionID := p.takeString()
	p.expect(tokenComma)
	p.expect(tokenSlug)
	p.expect(tokenColon)
	slug := p.takeIdentifier()
	p.expect(tokenComma)
	p.expect(tokenVersion)
	p.expect(tokenColon)
	version := p.takeInteger(tokenNumber)
	p.expect(tokenComma)
	p.expect(tokenDigest)
	p.expect(tokenColon)
	digest := p.takeString()
	p.expect(tokenRParen)
	return &CustomBlockRef{
		NodeID:           nodeID,
		DefinitionID:     definitionID,
		Slug:             slug,
		Version:          version,
		DefinitionDigest: digest,
	}
}

func (p *parser) parseReference() Reference {
	localID := p.takeIdentifier()
	p.expect(tokenDot)
	field := p.takeIdentifier()
	return Reference{LocalID: localID, Field: field}
}

func (p *parser) parseAgent(nodeID string) AgentNode {
	p.expect(tokenAgent)
	localID := p.takeIdentifier()
	p.expect(tokenLBrace)
	p.expect(tokenPromptKey)
	prompt, redacted := p.takeGuardedText(nodeID, "prompt string", tokenPrompt, tokenString)
	p.redactions.agentPrompt = p.redactions.agentPrompt || redacted
	p.expect(tokenRBrace)
	return AgentNode{
		NodeID: nodeID, LocalID: localID, Prompt: prompt,
	}
}

func (p *parser) parseVerifier(nodeID string) VerifierNode {
	p.expect(tokenVerify)
	localID := p.takeIdentifier()
	p.expect(tokenLBrace)
	p.expect(tokenCommand)
	command, commandRedacted := p.takeGuardedText(nodeID, "string", tokenString)
	p.redactions.verifierCommand = p.redactions.verifierCommand || commandRedacted
	p.expect(tokenAccept)
	accept, acceptRedacted := p.takeGuardedText(nodeID, "string", tokenString)
	p.redactions.verifierAccept = p.redactions.verifierAccept || acceptRedacted
	p.expect(tokenRBrace)
	return VerifierNode{
		NodeID: nodeID, LocalID: localID,
		Command: command, Accept: accept,
	}
}

func (p *parser) parseFailurePolicy() FailurePolicy {
	p.expect(tokenOnFailure)
	switch p.current().kind {
	case tokenPause:
		p.advance()
		return FailurePause
	case tokenFail:
		p.advance()
		return FailureFail
	default:
		p.failCurrent("loop.failure-policy.invalid", "failure policy must be pause or fail", "")
		return ""
	}
}
