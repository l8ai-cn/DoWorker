package parser

import (
	"strings"

	"github.com/l8ai-cn/agentcloud/agentfile/capability"
	"github.com/l8ai-cn/agentcloud/agentfile/lexer"
)

func (p *Parser) parseSetupDecl(pos Position) *SetupDecl {
	p.advance()
	decl := &SetupDecl{Position: pos, Timeout: 300}

	if p.currentIs(lexer.IDENT) && p.current().Literal == "timeout" {
		p.advance()
		p.expect(lexer.ASSIGN)
		decl.Timeout = p.expectInt()
	}
	if p.currentIs(lexer.HEREDOC_START) {
		p.advance()
		if p.currentIs(lexer.HEREDOC_BODY) {
			decl.Script = p.current().Literal
			p.advance()
		}
	}
	p.skipNewlines()
	return decl
}

// parseRemoveDecl: REMOVE ENV <name> | REMOVE SKILLS <slug> | REMOVE CONFIG <name> | REMOVE arg <name> | REMOVE file <path>
func (p *Parser) parseRemoveDecl(pos Position) *RemoveDecl {
	p.advance() // skip REMOVE
	tok := p.current()
	var target string
	switch tok.Type {
	case lexer.KW_ENV:
		target = "ENV"
	case lexer.KW_SKILLS:
		target = "SKILLS"
	case lexer.KW_CONFIG:
		target = "CONFIG"
	case lexer.KW_ARG:
		target = "arg"
	case lexer.KW_FILE:
		target = "file"
	default:
		p.errorf("REMOVE: expected ENV, SKILLS, CONFIG, arg, or file, got %s", tok.Literal)
		p.advance()
		return &RemoveDecl{Position: pos}
	}
	p.advance()
	name := p.expectIdentOrString()
	p.expectNewline()
	return &RemoveDecl{Target: target, Name: name, Position: pos}
}

// parseModeDecl: MODE pty | MODE acp | MODE acp "arg1" "arg2" ...
// Without args → ModeDecl (sets active mode).
// With string args → ModeArgsDecl (declares per-mode launch args).
func (p *Parser) parseModeDecl(pos Position) Declaration {
	p.advance()
	mode := p.expectIdentOrString()
	if mode != "pty" && mode != "acp" {
		p.errorf("MODE: expected pty or acp, got %s", mode)
	}

	// Check for per-mode args (string tokens on the same line)
	if !p.isNewlineOrEnd() && p.currentIs(lexer.STRING) {
		var args []string
		for p.currentIs(lexer.STRING) {
			args = append(args, p.expectString())
		}
		p.expectNewline()
		return &ModeArgsDecl{Mode: mode, Args: args, Position: pos}
	}

	p.expectNewline()
	return &ModeDecl{Mode: mode, Position: pos}
}

// parseUseEnvBundleDecl: USE_ENV_BUNDLE "bundle-name"
//
// The name is the bundle's primary identifier within the current owner scope
// (user-private bundles take precedence over org-shared ones with the same
// name; that resolution is done by the backend, not the parser).
func (p *Parser) parseUseEnvBundleDecl(pos Position) *UseEnvBundleDecl {
	p.advance()
	name := p.expectIdentOrString()
	p.expectNewline()
	return &UseEnvBundleDecl{Name: name, Position: pos}
}

func (p *Parser) parseUseConfigBundleDecl(pos Position) *UseConfigBundleDecl {
	p.advance()
	name := p.expectIdentOrString()
	p.expectNewline()
	return &UseConfigBundleDecl{Name: name, Position: pos}
}

// parsePromptDecl: PROMPT "prompt content"
func (p *Parser) parsePromptDecl(pos Position) *PromptDecl {
	p.advance()
	content := p.expectString()
	p.expectNewline()
	return &PromptDecl{Content: content, Position: pos}
}

// parsePromptPositionDecl: PROMPT_POSITION prepend | append | after_first | none
func (p *Parser) parsePromptPositionDecl(pos Position) *PromptPositionDecl {
	p.advance()
	tok := p.current()
	mode := tok.Literal
	if tok.Type != lexer.KW_PREPEND && tok.Type != lexer.KW_APPEND && tok.Type != lexer.KW_NONE &&
		(tok.Type != lexer.IDENT || mode != "after_first") {
		p.errorf("PROMPT_POSITION: expected prepend/append/after_first/none, got %s", tok.Literal)
	}
	p.advance()
	p.expectNewline()
	return &PromptPositionDecl{Mode: mode, Position: pos}
}

// parseCapabilityDecl: CAPABILITY <axis> <value>
func (p *Parser) parseCapabilityDecl(pos Position) *CapabilityDecl {
	p.advance()
	axis := strings.ToLower(p.expectIdent())
	var valueParts []string
	for !p.isNewlineOrEnd() {
		tok := p.current()
		switch tok.Type {
		case lexer.IDENT, lexer.STRING, lexer.TRUE, lexer.FALSE, lexer.KW_NONE:
			valueParts = append(valueParts, tok.Literal)
			p.advance()
		case lexer.COMMA:
			valueParts = append(valueParts, ",")
			p.advance()
		default:
			p.errorf("CAPABILITY %s: unexpected token %s", axis, tok.Literal)
			p.advance()
		}
	}
	var value string
	if axis == "control" {
		value = strings.Join(valueParts, "")
	} else if len(valueParts) == 1 {
		value = valueParts[0]
	} else {
		value = strings.Join(valueParts, " ")
	}
	if err := capability.Validate(axis, value); err != nil {
		p.errorf("%s", err.Error())
	}
	p.expectNewline()
	return &CapabilityDecl{Axis: axis, Value: value, Position: pos}
}
