package gendsl

// Code generated by peg -switch -inline -strict -output ./grammar.go grammar.peg DO NOT EDIT.

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleScript
	ruleExpression
	ruleOperator
	ruleValue
	ruleOptionList
	ruleOption
	ruleSpacing
	ruleIdentifier
	ruleLiteral
	ruleNilLiteral
	ruleBoolLiteral
	ruleFloatLiteral
	ruleExponent
	ruleIntegerLiteral
	ruleHexNumeral
	ruleDecimalNumeral
	ruleLongStringLiteral
	ruleLongStringChar
	ruleStringLiteral
	ruleStringChar
	ruleHexByte
	ruleUChar
	ruleLetterOrDigit
	ruleLetter
	ruleDigits
	ruleEscape
	ruleHexDigit
	ruleLPAR
	ruleRPAR
	ruleEOT
)

var rul3s = [...]string{
	"Unknown",
	"Script",
	"Expression",
	"Operator",
	"Value",
	"OptionList",
	"Option",
	"Spacing",
	"Identifier",
	"Literal",
	"NilLiteral",
	"BoolLiteral",
	"FloatLiteral",
	"Exponent",
	"IntegerLiteral",
	"HexNumeral",
	"DecimalNumeral",
	"LongStringLiteral",
	"LongStringChar",
	"StringLiteral",
	"StringChar",
	"HexByte",
	"UChar",
	"LetterOrDigit",
	"Letter",
	"Digits",
	"Escape",
	"HexDigit",
	"LPAR",
	"RPAR",
	"EOT",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(w io.Writer, pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Fprintf(w, " ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Fprintf(w, "%v %v\n", rule, quote)
			} else {
				fmt.Fprintf(w, "\x1B[36m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(w io.Writer, buffer string) {
	node.print(w, false, buffer)
}

func (node *node32) PrettyPrint(w io.Writer, buffer string) {
	node.print(w, true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(os.Stdout, buffer)
}

func (t *tokens32) WriteSyntaxTree(w io.Writer, buffer string) {
	t.AST().Print(w, buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(os.Stdout, buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	tree, i := t.tree, int(index)
	if i >= len(tree) {
		t.tree = append(tree, token32{pegRule: rule, begin: begin, end: end})
		return
	}
	tree[i] = token32{pegRule: rule, begin: begin, end: end}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type parser struct {
	Buffer string
	buffer []rune
	rules  [31]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *parser) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *parser) Reset() {
	p.reset()
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *parser
	max token32
}

func (e *parseError) Error() string {
	tokens, err := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		err += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return err
}

func (p *parser) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *parser) WriteSyntaxTree(w io.Writer) {
	p.tokens32.WriteSyntaxTree(w, p.Buffer)
}

func (p *parser) SprintSyntaxTree() string {
	var bldr strings.Builder
	p.WriteSyntaxTree(&bldr)
	return bldr.String()
}

func Pretty(pretty bool) func(*parser) error {
	return func(p *parser) error {
		p.Pretty = pretty
		return nil
	}
}

func Size(size int) func(*parser) error {
	return func(p *parser) error {
		p.tokens32 = tokens32{tree: make([]token32, 0, size)}
		return nil
	}
}
func (p *parser) Init(options ...func(*parser) error) error {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	for _, option := range options {
		err := option(p)
		if err != nil {
			return err
		}
	}
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := p.tokens32
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Script <- <(Value EOT)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[ruleValue]() {
					goto l0
				}
				{
					position2 := position
					{
						position3, tokenIndex3 := position, tokenIndex
						if !matchDot() {
							goto l3
						}
						goto l0
					l3:
						position, tokenIndex = position3, tokenIndex3
					}
					add(ruleEOT, position2)
				}
				add(ruleScript, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 Expression <- <(LPAR Operator OptionList? Value* RPAR)> */
		nil,
		/* 2 Operator <- <Identifier> */
		nil,
		/* 3 Value <- <((Expression / Literal / Identifier) Spacing)> */
		func() bool {
			position6, tokenIndex6 := position, tokenIndex
			{
				position7 := position
				{
					position8, tokenIndex8 := position, tokenIndex
					{
						position10 := position
						{
							position11 := position
							if !_rules[ruleSpacing]() {
								goto l9
							}
							if buffer[position] != rune('(') {
								goto l9
							}
							position++
							if !_rules[ruleSpacing]() {
								goto l9
							}
							add(ruleLPAR, position11)
						}
						{
							position12 := position
							if !_rules[ruleIdentifier]() {
								goto l9
							}
							add(ruleOperator, position12)
						}
						{
							position13, tokenIndex13 := position, tokenIndex
							{
								position15 := position
								{
									position18 := position
									if buffer[position] != rune('#') {
										goto l13
									}
									position++
									if buffer[position] != rune(':') {
										goto l13
									}
									position++
									if !_rules[ruleIdentifier]() {
										goto l13
									}
									{
										position19, tokenIndex19 := position, tokenIndex
										if !_rules[ruleLiteral]() {
											goto l20
										}
										goto l19
									l20:
										position, tokenIndex = position19, tokenIndex19
										if !_rules[ruleIdentifier]() {
											goto l13
										}
									}
								l19:
									if !_rules[ruleSpacing]() {
										goto l13
									}
									add(ruleOption, position18)
								}
							l16:
								{
									position17, tokenIndex17 := position, tokenIndex
									{
										position21 := position
										if buffer[position] != rune('#') {
											goto l17
										}
										position++
										if buffer[position] != rune(':') {
											goto l17
										}
										position++
										if !_rules[ruleIdentifier]() {
											goto l17
										}
										{
											position22, tokenIndex22 := position, tokenIndex
											if !_rules[ruleLiteral]() {
												goto l23
											}
											goto l22
										l23:
											position, tokenIndex = position22, tokenIndex22
											if !_rules[ruleIdentifier]() {
												goto l17
											}
										}
									l22:
										if !_rules[ruleSpacing]() {
											goto l17
										}
										add(ruleOption, position21)
									}
									goto l16
								l17:
									position, tokenIndex = position17, tokenIndex17
								}
								add(ruleOptionList, position15)
							}
							goto l14
						l13:
							position, tokenIndex = position13, tokenIndex13
						}
					l14:
					l24:
						{
							position25, tokenIndex25 := position, tokenIndex
							if !_rules[ruleValue]() {
								goto l25
							}
							goto l24
						l25:
							position, tokenIndex = position25, tokenIndex25
						}
						{
							position26 := position
							if !_rules[ruleSpacing]() {
								goto l9
							}
							if buffer[position] != rune(')') {
								goto l9
							}
							position++
							if !_rules[ruleSpacing]() {
								goto l9
							}
							add(ruleRPAR, position26)
						}
						add(ruleExpression, position10)
					}
					goto l8
				l9:
					position, tokenIndex = position8, tokenIndex8
					if !_rules[ruleLiteral]() {
						goto l27
					}
					goto l8
				l27:
					position, tokenIndex = position8, tokenIndex8
					if !_rules[ruleIdentifier]() {
						goto l6
					}
				}
			l8:
				if !_rules[ruleSpacing]() {
					goto l6
				}
				add(ruleValue, position7)
			}
			return true
		l6:
			position, tokenIndex = position6, tokenIndex6
			return false
		},
		/* 4 OptionList <- <Option+> */
		nil,
		/* 5 Option <- <('#' ':' Identifier (Literal / Identifier) Spacing)> */
		nil,
		/* 6 Spacing <- <(((&('\n') '\n') | (&('\r') '\r') | (&('\t') '\t') | (&(' ') ' '))+ / (';' (!('\r' / '\n') .)* ('\r' / '\n')))*> */
		func() bool {
			{
				position31 := position
			l32:
				{
					position33, tokenIndex33 := position, tokenIndex
					{
						position34, tokenIndex34 := position, tokenIndex
						{
							switch buffer[position] {
							case '\n':
								if buffer[position] != rune('\n') {
									goto l35
								}
								position++
							case '\r':
								if buffer[position] != rune('\r') {
									goto l35
								}
								position++
							case '\t':
								if buffer[position] != rune('\t') {
									goto l35
								}
								position++
							default:
								if buffer[position] != rune(' ') {
									goto l35
								}
								position++
							}
						}

					l36:
						{
							position37, tokenIndex37 := position, tokenIndex
							{
								switch buffer[position] {
								case '\n':
									if buffer[position] != rune('\n') {
										goto l37
									}
									position++
								case '\r':
									if buffer[position] != rune('\r') {
										goto l37
									}
									position++
								case '\t':
									if buffer[position] != rune('\t') {
										goto l37
									}
									position++
								default:
									if buffer[position] != rune(' ') {
										goto l37
									}
									position++
								}
							}

							goto l36
						l37:
							position, tokenIndex = position37, tokenIndex37
						}
						goto l34
					l35:
						position, tokenIndex = position34, tokenIndex34
						if buffer[position] != rune(';') {
							goto l33
						}
						position++
					l40:
						{
							position41, tokenIndex41 := position, tokenIndex
							{
								position42, tokenIndex42 := position, tokenIndex
								{
									position43, tokenIndex43 := position, tokenIndex
									if buffer[position] != rune('\r') {
										goto l44
									}
									position++
									goto l43
								l44:
									position, tokenIndex = position43, tokenIndex43
									if buffer[position] != rune('\n') {
										goto l42
									}
									position++
								}
							l43:
								goto l41
							l42:
								position, tokenIndex = position42, tokenIndex42
							}
							if !matchDot() {
								goto l41
							}
							goto l40
						l41:
							position, tokenIndex = position41, tokenIndex41
						}
						{
							position45, tokenIndex45 := position, tokenIndex
							if buffer[position] != rune('\r') {
								goto l46
							}
							position++
							goto l45
						l46:
							position, tokenIndex = position45, tokenIndex45
							if buffer[position] != rune('\n') {
								goto l33
							}
							position++
						}
					l45:
					}
				l34:
					goto l32
				l33:
					position, tokenIndex = position33, tokenIndex33
				}
				add(ruleSpacing, position31)
			}
			return true
		},
		/* 7 Identifier <- <(((&('@') '@') | (&('$') '$') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') Letter)) (LetterOrDigit / '-')* Spacing)> */
		func() bool {
			position47, tokenIndex47 := position, tokenIndex
			{
				position48 := position
				{
					switch buffer[position] {
					case '@':
						if buffer[position] != rune('@') {
							goto l47
						}
						position++
					case '$':
						if buffer[position] != rune('$') {
							goto l47
						}
						position++
					default:
						{
							position50 := position
							{
								switch buffer[position] {
								case '_':
									if buffer[position] != rune('_') {
										goto l47
									}
									position++
								case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
									if c := buffer[position]; c < rune('A') || c > rune('Z') {
										goto l47
									}
									position++
								default:
									if c := buffer[position]; c < rune('a') || c > rune('z') {
										goto l47
									}
									position++
								}
							}

							add(ruleLetter, position50)
						}
					}
				}

			l52:
				{
					position53, tokenIndex53 := position, tokenIndex
					{
						position54, tokenIndex54 := position, tokenIndex
						if !_rules[ruleLetterOrDigit]() {
							goto l55
						}
						goto l54
					l55:
						position, tokenIndex = position54, tokenIndex54
						if buffer[position] != rune('-') {
							goto l53
						}
						position++
					}
				l54:
					goto l52
				l53:
					position, tokenIndex = position53, tokenIndex53
				}
				if !_rules[ruleSpacing]() {
					goto l47
				}
				add(ruleIdentifier, position48)
			}
			return true
		l47:
			position, tokenIndex = position47, tokenIndex47
			return false
		},
		/* 8 Literal <- <((FloatLiteral / LongStringLiteral / ((&('#') BoolLiteral) | (&('"') StringLiteral) | (&('n') NilLiteral) | (&('+' | '-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') IntegerLiteral))) Spacing)> */
		func() bool {
			position56, tokenIndex56 := position, tokenIndex
			{
				position57 := position
				{
					position58, tokenIndex58 := position, tokenIndex
					{
						position60 := position
						{
							position61, tokenIndex61 := position, tokenIndex
							{
								position63, tokenIndex63 := position, tokenIndex
								if buffer[position] != rune('+') {
									goto l64
								}
								position++
								goto l63
							l64:
								position, tokenIndex = position63, tokenIndex63
								if buffer[position] != rune('-') {
									goto l61
								}
								position++
							}
						l63:
							goto l62
						l61:
							position, tokenIndex = position61, tokenIndex61
						}
					l62:
						{
							position65, tokenIndex65 := position, tokenIndex
							if !_rules[ruleDigits]() {
								goto l66
							}
							if buffer[position] != rune('.') {
								goto l66
							}
							position++
							{
								position67, tokenIndex67 := position, tokenIndex
								if !_rules[ruleDigits]() {
									goto l67
								}
								goto l68
							l67:
								position, tokenIndex = position67, tokenIndex67
							}
						l68:
							{
								position69, tokenIndex69 := position, tokenIndex
								if !_rules[ruleExponent]() {
									goto l69
								}
								goto l70
							l69:
								position, tokenIndex = position69, tokenIndex69
							}
						l70:
							goto l65
						l66:
							position, tokenIndex = position65, tokenIndex65
							if !_rules[ruleDigits]() {
								goto l71
							}
							if !_rules[ruleExponent]() {
								goto l71
							}
							goto l65
						l71:
							position, tokenIndex = position65, tokenIndex65
							if buffer[position] != rune('.') {
								goto l59
							}
							position++
							if !_rules[ruleDigits]() {
								goto l59
							}
							{
								position72, tokenIndex72 := position, tokenIndex
								if !_rules[ruleExponent]() {
									goto l72
								}
								goto l73
							l72:
								position, tokenIndex = position72, tokenIndex72
							}
						l73:
						}
					l65:
						add(ruleFloatLiteral, position60)
					}
					goto l58
				l59:
					position, tokenIndex = position58, tokenIndex58
					{
						position75 := position
						if buffer[position] != rune('"') {
							goto l74
						}
						position++
						if buffer[position] != rune('"') {
							goto l74
						}
						position++
						if buffer[position] != rune('"') {
							goto l74
						}
						position++
					l76:
						{
							position77, tokenIndex77 := position, tokenIndex
							{
								position78 := position
								{
									position79, tokenIndex79 := position, tokenIndex
									if buffer[position] != rune('"') {
										goto l79
									}
									position++
									goto l77
								l79:
									position, tokenIndex = position79, tokenIndex79
								}
								if !matchDot() {
									goto l77
								}
								add(ruleLongStringChar, position78)
							}
							goto l76
						l77:
							position, tokenIndex = position77, tokenIndex77
						}
						if buffer[position] != rune('"') {
							goto l74
						}
						position++
						if buffer[position] != rune('"') {
							goto l74
						}
						position++
						if buffer[position] != rune('"') {
							goto l74
						}
						position++
						add(ruleLongStringLiteral, position75)
					}
					goto l58
				l74:
					position, tokenIndex = position58, tokenIndex58
					{
						switch buffer[position] {
						case '#':
							{
								position81 := position
								{
									position82, tokenIndex82 := position, tokenIndex
									if buffer[position] != rune('#') {
										goto l83
									}
									position++
									if buffer[position] != rune('f') {
										goto l83
									}
									position++
									goto l82
								l83:
									position, tokenIndex = position82, tokenIndex82
									if buffer[position] != rune('#') {
										goto l56
									}
									position++
									if buffer[position] != rune('t') {
										goto l56
									}
									position++
								}
							l82:
								{
									position84, tokenIndex84 := position, tokenIndex
									if !_rules[ruleLetterOrDigit]() {
										goto l84
									}
									goto l56
								l84:
									position, tokenIndex = position84, tokenIndex84
								}
								add(ruleBoolLiteral, position81)
							}
						case '"':
							{
								position85 := position
								if buffer[position] != rune('"') {
									goto l56
								}
								position++
							l86:
								{
									position87, tokenIndex87 := position, tokenIndex
									{
										position88 := position
										{
											position89, tokenIndex89 := position, tokenIndex
											{
												position91 := position
												{
													position92, tokenIndex92 := position, tokenIndex
													if buffer[position] != rune('\\') {
														goto l93
													}
													position++
													if buffer[position] != rune('u') {
														goto l93
													}
													position++
													if !_rules[ruleHexDigit]() {
														goto l93
													}
													if !_rules[ruleHexDigit]() {
														goto l93
													}
													if !_rules[ruleHexDigit]() {
														goto l93
													}
													if !_rules[ruleHexDigit]() {
														goto l93
													}
													goto l92
												l93:
													position, tokenIndex = position92, tokenIndex92
													if buffer[position] != rune('\\') {
														goto l90
													}
													position++
													if buffer[position] != rune('U') {
														goto l90
													}
													position++
													if !_rules[ruleHexDigit]() {
														goto l90
													}
													if !_rules[ruleHexDigit]() {
														goto l90
													}
													if !_rules[ruleHexDigit]() {
														goto l90
													}
													if !_rules[ruleHexDigit]() {
														goto l90
													}
													if !_rules[ruleHexDigit]() {
														goto l90
													}
													if !_rules[ruleHexDigit]() {
														goto l90
													}
													if !_rules[ruleHexDigit]() {
														goto l90
													}
													if !_rules[ruleHexDigit]() {
														goto l90
													}
												}
											l92:
												add(ruleUChar, position91)
											}
											goto l89
										l90:
											position, tokenIndex = position89, tokenIndex89
											{
												position95 := position
												if buffer[position] != rune('\\') {
													goto l94
												}
												position++
												{
													switch buffer[position] {
													case '\'':
														if buffer[position] != rune('\'') {
															goto l94
														}
														position++
													case '"':
														if buffer[position] != rune('"') {
															goto l94
														}
														position++
													case '\\':
														if buffer[position] != rune('\\') {
															goto l94
														}
														position++
													case 'v':
														if buffer[position] != rune('v') {
															goto l94
														}
														position++
													case 't':
														if buffer[position] != rune('t') {
															goto l94
														}
														position++
													case 'r':
														if buffer[position] != rune('r') {
															goto l94
														}
														position++
													case 'n':
														if buffer[position] != rune('n') {
															goto l94
														}
														position++
													case 'f':
														if buffer[position] != rune('f') {
															goto l94
														}
														position++
													case 'b':
														if buffer[position] != rune('b') {
															goto l94
														}
														position++
													default:
														if buffer[position] != rune('a') {
															goto l94
														}
														position++
													}
												}

												add(ruleEscape, position95)
											}
											goto l89
										l94:
											position, tokenIndex = position89, tokenIndex89
											{
												position98 := position
												if buffer[position] != rune('\\') {
													goto l97
												}
												position++
												if buffer[position] != rune('x') {
													goto l97
												}
												position++
												if !_rules[ruleHexDigit]() {
													goto l97
												}
												if !_rules[ruleHexDigit]() {
													goto l97
												}
												add(ruleHexByte, position98)
											}
											goto l89
										l97:
											position, tokenIndex = position89, tokenIndex89
											{
												position99, tokenIndex99 := position, tokenIndex
												{
													switch buffer[position] {
													case '\\':
														if buffer[position] != rune('\\') {
															goto l99
														}
														position++
													case '\n':
														if buffer[position] != rune('\n') {
															goto l99
														}
														position++
													default:
														if buffer[position] != rune('"') {
															goto l99
														}
														position++
													}
												}

												goto l87
											l99:
												position, tokenIndex = position99, tokenIndex99
											}
											if !matchDot() {
												goto l87
											}
										}
									l89:
										add(ruleStringChar, position88)
									}
									goto l86
								l87:
									position, tokenIndex = position87, tokenIndex87
								}
								if buffer[position] != rune('"') {
									goto l56
								}
								position++
								add(ruleStringLiteral, position85)
							}
						case 'n':
							{
								position101 := position
								if buffer[position] != rune('n') {
									goto l56
								}
								position++
								if buffer[position] != rune('i') {
									goto l56
								}
								position++
								if buffer[position] != rune('l') {
									goto l56
								}
								position++
								add(ruleNilLiteral, position101)
							}
						default:
							{
								position102 := position
								{
									position103, tokenIndex103 := position, tokenIndex
									{
										position105, tokenIndex105 := position, tokenIndex
										if buffer[position] != rune('+') {
											goto l106
										}
										position++
										goto l105
									l106:
										position, tokenIndex = position105, tokenIndex105
										if buffer[position] != rune('-') {
											goto l103
										}
										position++
									}
								l105:
									goto l104
								l103:
									position, tokenIndex = position103, tokenIndex103
								}
							l104:
								{
									position107, tokenIndex107 := position, tokenIndex
									if buffer[position] != rune('0') {
										goto l108
									}
									position++
									{
										position109, tokenIndex109 := position, tokenIndex
										if buffer[position] != rune('x') {
											goto l110
										}
										position++
										goto l109
									l110:
										position, tokenIndex = position109, tokenIndex109
										if buffer[position] != rune('X') {
											goto l108
										}
										position++
									}
								l109:
									{
										position111 := position
										{
											position112, tokenIndex112 := position, tokenIndex
											if !_rules[ruleHexDigit]() {
												goto l113
											}
										l114:
											{
												position115, tokenIndex115 := position, tokenIndex
											l116:
												{
													position117, tokenIndex117 := position, tokenIndex
													if buffer[position] != rune('_') {
														goto l117
													}
													position++
													goto l116
												l117:
													position, tokenIndex = position117, tokenIndex117
												}
												if !_rules[ruleHexDigit]() {
													goto l115
												}
												goto l114
											l115:
												position, tokenIndex = position115, tokenIndex115
											}
											goto l112
										l113:
											position, tokenIndex = position112, tokenIndex112
											if buffer[position] != rune('0') {
												goto l108
											}
											position++
										}
									l112:
										add(ruleHexNumeral, position111)
									}
									goto l107
								l108:
									position, tokenIndex = position107, tokenIndex107
									{
										position118 := position
										{
											position119, tokenIndex119 := position, tokenIndex
											if c := buffer[position]; c < rune('1') || c > rune('9') {
												goto l120
											}
											position++
										l121:
											{
												position122, tokenIndex122 := position, tokenIndex
											l123:
												{
													position124, tokenIndex124 := position, tokenIndex
													if buffer[position] != rune('_') {
														goto l124
													}
													position++
													goto l123
												l124:
													position, tokenIndex = position124, tokenIndex124
												}
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l122
												}
												position++
												goto l121
											l122:
												position, tokenIndex = position122, tokenIndex122
											}
											goto l119
										l120:
											position, tokenIndex = position119, tokenIndex119
											if buffer[position] != rune('0') {
												goto l56
											}
											position++
										}
									l119:
										add(ruleDecimalNumeral, position118)
									}
								}
							l107:
								{
									position125, tokenIndex125 := position, tokenIndex
									{
										position127, tokenIndex127 := position, tokenIndex
										if buffer[position] != rune('u') {
											goto l128
										}
										position++
										goto l127
									l128:
										position, tokenIndex = position127, tokenIndex127
										if buffer[position] != rune('U') {
											goto l125
										}
										position++
									}
								l127:
									goto l126
								l125:
									position, tokenIndex = position125, tokenIndex125
								}
							l126:
								add(ruleIntegerLiteral, position102)
							}
						}
					}

				}
			l58:
				if !_rules[ruleSpacing]() {
					goto l56
				}
				add(ruleLiteral, position57)
			}
			return true
		l56:
			position, tokenIndex = position56, tokenIndex56
			return false
		},
		/* 9 NilLiteral <- <('n' 'i' 'l')> */
		nil,
		/* 10 BoolLiteral <- <((('#' 'f') / ('#' 't')) !LetterOrDigit)> */
		nil,
		/* 11 FloatLiteral <- <(('+' / '-')? ((Digits '.' Digits? Exponent?) / (Digits Exponent) / ('.' Digits Exponent?)))> */
		nil,
		/* 12 Exponent <- <(('e' / 'E') ('+' / '-')? Digits)> */
		func() bool {
			position132, tokenIndex132 := position, tokenIndex
			{
				position133 := position
				{
					position134, tokenIndex134 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l135
					}
					position++
					goto l134
				l135:
					position, tokenIndex = position134, tokenIndex134
					if buffer[position] != rune('E') {
						goto l132
					}
					position++
				}
			l134:
				{
					position136, tokenIndex136 := position, tokenIndex
					{
						position138, tokenIndex138 := position, tokenIndex
						if buffer[position] != rune('+') {
							goto l139
						}
						position++
						goto l138
					l139:
						position, tokenIndex = position138, tokenIndex138
						if buffer[position] != rune('-') {
							goto l136
						}
						position++
					}
				l138:
					goto l137
				l136:
					position, tokenIndex = position136, tokenIndex136
				}
			l137:
				if !_rules[ruleDigits]() {
					goto l132
				}
				add(ruleExponent, position133)
			}
			return true
		l132:
			position, tokenIndex = position132, tokenIndex132
			return false
		},
		/* 13 IntegerLiteral <- <(('+' / '-')? (('0' ('x' / 'X') HexNumeral) / DecimalNumeral) ('u' / 'U')?)> */
		nil,
		/* 14 HexNumeral <- <((HexDigit ('_'* HexDigit)*) / '0')> */
		nil,
		/* 15 DecimalNumeral <- <(([1-9] ('_'* [0-9])*) / '0')> */
		nil,
		/* 16 LongStringLiteral <- <('"' '"' '"' LongStringChar* ('"' '"' '"'))> */
		nil,
		/* 17 LongStringChar <- <(!'"' .)> */
		nil,
		/* 18 StringLiteral <- <('"' StringChar* '"')> */
		nil,
		/* 19 StringChar <- <(UChar / Escape / HexByte / (!((&('\\') '\\') | (&('\n') '\n') | (&('"') '"')) .))> */
		nil,
		/* 20 HexByte <- <('\\' 'x' HexDigit HexDigit)> */
		nil,
		/* 21 UChar <- <(('\\' 'u' HexDigit HexDigit HexDigit HexDigit) / ('\\' 'U' HexDigit HexDigit HexDigit HexDigit HexDigit HexDigit HexDigit HexDigit))> */
		nil,
		/* 22 LetterOrDigit <- <((&('_') '_') | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position149, tokenIndex149 := position, tokenIndex
			{
				position150 := position
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l149
						}
						position++
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l149
						}
						position++
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l149
						}
						position++
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l149
						}
						position++
					}
				}

				add(ruleLetterOrDigit, position150)
			}
			return true
		l149:
			position, tokenIndex = position149, tokenIndex149
			return false
		},
		/* 23 Letter <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		nil,
		/* 24 Digits <- <([0-9] ('_'* [0-9])*)> */
		func() bool {
			position153, tokenIndex153 := position, tokenIndex
			{
				position154 := position
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l153
				}
				position++
			l155:
				{
					position156, tokenIndex156 := position, tokenIndex
				l157:
					{
						position158, tokenIndex158 := position, tokenIndex
						if buffer[position] != rune('_') {
							goto l158
						}
						position++
						goto l157
					l158:
						position, tokenIndex = position158, tokenIndex158
					}
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l156
					}
					position++
					goto l155
				l156:
					position, tokenIndex = position156, tokenIndex156
				}
				add(ruleDigits, position154)
			}
			return true
		l153:
			position, tokenIndex = position153, tokenIndex153
			return false
		},
		/* 25 Escape <- <('\\' ((&('\'') '\'') | (&('"') '"') | (&('\\') '\\') | (&('v') 'v') | (&('t') 't') | (&('r') 'r') | (&('n') 'n') | (&('f') 'f') | (&('b') 'b') | (&('a') 'a')))> */
		nil,
		/* 26 HexDigit <- <((&('a' | 'b' | 'c' | 'd' | 'e' | 'f') [a-f]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F') [A-F]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]))> */
		func() bool {
			position160, tokenIndex160 := position, tokenIndex
			{
				position161 := position
				{
					switch buffer[position] {
					case 'a', 'b', 'c', 'd', 'e', 'f':
						if c := buffer[position]; c < rune('a') || c > rune('f') {
							goto l160
						}
						position++
					case 'A', 'B', 'C', 'D', 'E', 'F':
						if c := buffer[position]; c < rune('A') || c > rune('F') {
							goto l160
						}
						position++
					default:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l160
						}
						position++
					}
				}

				add(ruleHexDigit, position161)
			}
			return true
		l160:
			position, tokenIndex = position160, tokenIndex160
			return false
		},
		/* 27 LPAR <- <(Spacing '(' Spacing)> */
		nil,
		/* 28 RPAR <- <(Spacing ')' Spacing)> */
		nil,
		/* 29 EOT <- <!.> */
		nil,
	}
	p.rules = _rules
	return nil
}
