package loopscript

type tokenKind string

const (
	tokenEOF      tokenKind = "EOF"
	tokenIllegal  tokenKind = "illegal"
	tokenIdent    tokenKind = "identifier"
	tokenNumber   tokenKind = "number"
	tokenDuration tokenKind = "duration"
	tokenString   tokenKind = "string"
	tokenPrompt   tokenKind = "prompt-string"

	tokenAt     tokenKind = "@"
	tokenEqual  tokenKind = "="
	tokenDot    tokenKind = "."
	tokenComma  tokenKind = ","
	tokenColon  tokenKind = ":"
	tokenLParen tokenKind = "("
	tokenRParen tokenKind = ")"
	tokenLBrace tokenKind = "{"
	tokenRBrace tokenKind = "}"

	tokenID         tokenKind = "id"
	tokenLoop       tokenKind = "loop"
	tokenLimits     tokenKind = "limits"
	tokenIterations tokenKind = "iterations"
	tokenTokens     tokenKind = "tokens"
	tokenTimeout    tokenKind = "timeout"
	tokenNoProgress tokenKind = "no_progress"
	tokenSameError  tokenKind = "same_error"
	tokenRepeat     tokenKind = "repeat"
	tokenMax        tokenKind = "max"
	tokenUntil      tokenKind = "until"
	tokenAgent      tokenKind = "agent"
	tokenPromptKey  tokenKind = "prompt"
	tokenVerify     tokenKind = "verify"
	tokenCommand    tokenKind = "command"
	tokenAccept     tokenKind = "accept"
	tokenOnFailure  tokenKind = "on_failure"
	tokenPause      tokenKind = "pause"
	tokenFail       tokenKind = "fail"
	tokenSecret     tokenKind = "secret"
)

var keywords = map[string]tokenKind{
	"id":          tokenID,
	"loop":        tokenLoop,
	"limits":      tokenLimits,
	"iterations":  tokenIterations,
	"tokens":      tokenTokens,
	"timeout":     tokenTimeout,
	"no_progress": tokenNoProgress,
	"same_error":  tokenSameError,
	"repeat":      tokenRepeat,
	"max":         tokenMax,
	"until":       tokenUntil,
	"agent":       tokenAgent,
	"prompt":      tokenPromptKey,
	"verify":      tokenVerify,
	"command":     tokenCommand,
	"accept":      tokenAccept,
	"on_failure":  tokenOnFailure,
	"pause":       tokenPause,
	"fail":        tokenFail,
	"secret":      tokenSecret,
}

type token struct {
	kind    tokenKind
	literal string
	line    int
	column  int
}
