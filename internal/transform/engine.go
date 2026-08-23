package transform

import (
	"errors"
	"fmt"
	"github.com/example/cdc-replication/internal/cdc_domain"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenString
	TokenNumber
	TokenEq
	TokenAnd
	TokenOr
	TokenNot
	TokenLParen
	TokenRParen
)

type Token struct {
	Type TokenType
	Text string
}
type Lexer struct {
	s   string
	pos int
}

func NewLexer(s string) *Lexer { return &Lexer{s: s} }
func (l *Lexer) Next() Token {
	for l.pos < len(l.s) && unicode.IsSpace(rune(l.s[l.pos])) {
		l.pos++
	}
	if l.pos >= len(l.s) {
		return Token{Type: TokenEOF}
	}
	c := l.s[l.pos]
	if c == '(' {
		l.pos++
		return Token{Type: TokenLParen, Text: "("}
	}
	if c == ')' {
		l.pos++
		return Token{Type: TokenRParen, Text: ")"}
	}
	if c == '=' {
		l.pos++
		return Token{Type: TokenEq, Text: "="}
	}
	if c == '\'' || c == '"' {
		q := c
		l.pos++
		start := l.pos
		for l.pos < len(l.s) && l.s[l.pos] != q {
			l.pos++
		}
		text := l.s[start:l.pos]
		if l.pos < len(l.s) {
			l.pos++
		}
		return Token{Type: TokenString, Text: text}
	}
	start := l.pos
	for l.pos < len(l.s) && (unicode.IsLetter(rune(l.s[l.pos])) || unicode.IsDigit(rune(l.s[l.pos])) || l.s[l.pos] == '_' || l.s[l.pos] == '.' || l.s[l.pos] == '-') {
		l.pos++
	}
	text := l.s[start:l.pos]
	low := strings.ToLower(text)
	typ := TokenIdent
	if _, e := strconv.ParseFloat(text, 64); e == nil {
		typ = TokenNumber
	}
	if low == "and" {
		typ = TokenAnd
	}
	if low == "or" {
		typ = TokenOr
	}
	if low == "not" {
		typ = TokenNot
	}
	return Token{Type: typ, Text: text}
}

type Expr interface {
	Eval(map[string]any) bool
	String() string
}
type Compare struct {
	Field, Value string
	Negate       bool
}

func (c Compare) Eval(row map[string]any) bool {
	ok := fmt.Sprint(row[c.Field]) == c.Value
	if c.Negate {
		return !ok
	}
	return ok
}
func (c Compare) String() string {
	if c.Negate {
		return c.Field + " != '" + c.Value + "'"
	}
	return c.Field + " = '" + c.Value + "'"
}

type Binary struct {
	Left, Right Expr
	Or          bool
}

func (b Binary) Eval(r map[string]any) bool {
	if b.Or {
		return b.Left.Eval(r) || b.Right.Eval(r)
	}
	return b.Left.Eval(r) && b.Right.Eval(r)
}
func (b Binary) String() string {
	op := " AND "
	if b.Or {
		op = " OR "
	}
	return "(" + b.Left.String() + op + b.Right.String() + ")"
}

type Parser struct {
	lex *Lexer
	cur Token
}

func NewParser(s string) *Parser { p := &Parser{lex: NewLexer(s)}; p.cur = p.lex.Next(); return p }
func (p *Parser) advance()       { p.cur = p.lex.Next() }
func (p *Parser) Parse() (Expr, error) {
	if p.cur.Type == TokenEOF {
		return nil, nil
	}
	return p.parseOr()
}
func (p *Parser) parseOr() (Expr, error) {
	left, e := p.parseAnd()
	if e != nil {
		return nil, e
	}
	for p.cur.Type == TokenOr {
		p.advance()
		right, e := p.parseAnd()
		if e != nil {
			return nil, e
		}
		left = Binary{Left: left, Right: right, Or: true}
	}
	return left, nil
}
func (p *Parser) parseAnd() (Expr, error) {
	left, e := p.parseAtom()
	if e != nil {
		return nil, e
	}
	for p.cur.Type == TokenAnd {
		p.advance()
		right, e := p.parseAtom()
		if e != nil {
			return nil, e
		}
		left = Binary{Left: left, Right: right}
	}
	return left, nil
}
func (p *Parser) parseAtom() (Expr, error) {
	if p.cur.Type == TokenLParen {
		p.advance()
		e, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.cur.Type != TokenRParen {
			return nil, errors.New("missing )")
		}
		p.advance()
		return e, nil
	}
	if p.cur.Type != TokenIdent {
		return nil, fmt.Errorf("expected field, got %s", p.cur.Text)
	}
	field := p.cur.Text
	p.advance()
	neg := false
	if p.cur.Type == TokenNot {
		neg = true
		p.advance()
	}
	if p.cur.Type != TokenEq {
		return nil, errors.New("expected =")
	}
	p.advance()
	if p.cur.Type != TokenString && p.cur.Type != TokenNumber && p.cur.Type != TokenIdent {
		return nil, errors.New("expected value")
	}
	value := p.cur.Text
	p.advance()
	return Compare{Field: field, Value: value, Negate: neg}, nil
}

type Mapping struct {
	Version int
	Rename  map[string]string
	Cast    map[string]string
	Mask    []string
	Filter  Expr
}
type Result struct {
	Event   cdc_domain.ChangeEvent
	Changed []string
	Dropped bool
	Errors  []string
}

func (m Mapping) Apply(e cdc_domain.ChangeEvent) Result {
	r := Result{Event: e}
	if m.Filter != nil && !m.Filter.Eval(e.After) {
		r.Dropped = true
		return r
	}
	out := map[string]any{}
	for k, v := range e.After {
		nk := k
		if x, ok := m.Rename[k]; ok {
			nk = x
		}
		if contains(m.Mask, k) {
			out[nk] = Mask(fmt.Sprint(v))
		} else if typ := m.Cast[k]; typ != "" {
			x, err := Cast(v, typ)
			if err != nil {
				r.Errors = append(r.Errors, err.Error())
				continue
			}
			out[nk] = x
		} else {
			out[nk] = v
		}
		r.Changed = append(r.Changed, nk)
	}
	r.Event.After = out
	r.Event.SchemaVersion = m.Version
	return r
}
func contains(a []string, s string) bool {
	for _, x := range a {
		if x == s {
			return true
		}
	}
	return false
}
func Mask(s string) string {
	if s == "" {
		return ""
	}
	return strings.Repeat("*", len([]rune(s)))
}
func Cast(v any, typ string) (any, error) {
	s := fmt.Sprint(v)
	switch strings.ToLower(typ) {
	case "string":
		return s, nil
	case "int", "int64":
		x, e := strconv.ParseInt(s, 10, 64)
		return x, e
	case "float64":
		x, e := strconv.ParseFloat(s, 64)
		return x, e
	case "bool":
		x, e := strconv.ParseBool(s)
		return x, e
	}
	return nil, fmt.Errorf("unsupported cast %s", typ)
}

var unsafeScript = regexp.MustCompile(`[;{}]|\b(exec|eval|system)\b`)

func ValidateFilter(s string) error {
	if unsafeScript.MatchString(strings.ToLower(s)) {
		return errors.New("filter contains prohibited expression")
	}
	_, e := NewParser(s).Parse()
	return e
}
