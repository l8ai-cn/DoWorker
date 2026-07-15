package loopscript

import (
	"strconv"
	"unicode"
)

type lexer struct {
	input  []rune
	pos    int
	line   int
	column int
	tokens []token
}

func lex(source string) []token {
	l := &lexer{input: []rune(source), line: 1, column: 1}
	for l.pos < len(l.input) {
		if unicode.IsSpace(l.current()) {
			l.advance()
			continue
		}
		line, column := l.line, l.column
		switch ch := l.current(); {
		case ch == '"' && l.peek(1) == '"' && l.peek(2) == '"':
			l.readPrompt(line, column)
		case ch == '"':
			l.readString(line, column)
		case isASCIIDigit(ch):
			l.readNumber(line, column)
		case isWordStart(ch):
			l.readWord(line, column)
		default:
			kind, ok := symbolKind(ch)
			if !ok {
				l.emit(tokenIllegal, "unexpected character "+string(ch), line, column)
				l.advance()
				continue
			}
			l.emit(kind, string(ch), line, column)
			l.advance()
		}
	}
	l.emit(tokenEOF, "", l.line, l.column)
	return l.tokens
}

func (l *lexer) readWord(line, column int) {
	start := l.pos
	for l.pos < len(l.input) && isWordPart(l.current()) {
		l.advance()
	}
	literal := string(l.input[start:l.pos])
	kind := tokenIdent
	if keyword, ok := keywords[literal]; ok {
		kind = keyword
	}
	l.emit(kind, literal, line, column)
}

func (l *lexer) readNumber(line, column int) {
	start := l.pos
	for l.pos < len(l.input) && isASCIIDigit(l.current()) {
		l.advance()
	}
	kind := tokenNumber
	if l.current() == 'm' && !isWordPart(l.peek(1)) {
		kind = tokenDuration
		l.advance()
	} else if isWordPart(l.current()) {
		for l.pos < len(l.input) && isWordPart(l.current()) {
			l.advance()
		}
		kind = tokenIdent
	}
	l.emit(kind, string(l.input[start:l.pos]), line, column)
}

func (l *lexer) readString(line, column int) {
	start := l.pos
	l.advance()
	for l.pos < len(l.input) {
		switch l.current() {
		case '\n':
			l.emit(tokenIllegal, "unterminated string", line, column)
			return
		case '\\':
			l.advance()
			if l.pos < len(l.input) {
				l.advance()
			}
		case '"':
			l.advance()
			value, err := strconv.Unquote(string(l.input[start:l.pos]))
			if err != nil {
				l.emit(tokenIllegal, "invalid string escape", line, column)
				return
			}
			l.emit(tokenString, value, line, column)
			return
		default:
			l.advance()
		}
	}
	l.emit(tokenIllegal, "unterminated string", line, column)
}

func (l *lexer) readPrompt(line, column int) {
	l.advance()
	l.advance()
	l.advance()
	start := l.pos
	for l.pos < len(l.input) {
		if l.current() == '"' && l.peek(1) == '"' && l.peek(2) == '"' {
			value := string(l.input[start:l.pos])
			l.advance()
			l.advance()
			l.advance()
			l.emit(tokenPrompt, value, line, column)
			return
		}
		l.advance()
	}
	l.emit(tokenIllegal, "unterminated prompt string", line, column)
}

func (l *lexer) current() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *lexer) peek(offset int) rune {
	index := l.pos + offset
	if index >= len(l.input) {
		return 0
	}
	return l.input[index]
}

func (l *lexer) advance() {
	if l.pos >= len(l.input) {
		return
	}
	if l.input[l.pos] == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	l.pos++
}

func (l *lexer) emit(kind tokenKind, literal string, line, column int) {
	l.tokens = append(l.tokens, token{kind: kind, literal: literal, line: line, column: column})
}

func isASCIIDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isWordStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '-'
}

func isWordPart(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '-'
}

func symbolKind(ch rune) (tokenKind, bool) {
	symbols := map[rune]tokenKind{
		'@': tokenAt, '=': tokenEqual, '.': tokenDot, ',': tokenComma,
		':': tokenColon, '(': tokenLParen, ')': tokenRParen,
		'{': tokenLBrace, '}': tokenRBrace,
	}
	kind, ok := symbols[ch]
	return kind, ok
}
