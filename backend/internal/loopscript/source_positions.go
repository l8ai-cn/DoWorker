package loopscript

func loopPosition(p *programPositions) sourcePosition {
	if p != nil {
		return p.loop
	}
	return sourcePosition{}
}

func limitsPosition(p *programPositions) sourcePosition {
	if p != nil {
		return p.limits
	}
	return sourcePosition{}
}

func repeatPosition(p *programPositions) sourcePosition {
	if p != nil {
		return p.repeat
	}
	return sourcePosition{}
}

func agentPosition(p *programPositions) sourcePosition {
	if p != nil {
		return p.agent
	}
	return sourcePosition{}
}

func verifierPosition(p *programPositions) sourcePosition {
	if p != nil {
		return p.verifier
	}
	return sourcePosition{}
}

func failurePosition(p *programPositions) sourcePosition {
	if p != nil {
		return p.failure
	}
	return sourcePosition{}
}
