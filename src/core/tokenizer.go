//
// Helena tokenization
//

package core

//
// Helena token type for each special character or sequence
//
type TokenType int8

const (
	TokenType_WHITESPACE TokenType = iota
	TokenType_NEWLINE
	TokenType_CONTINUATION
	TokenType_TEXT
	TokenType_ESCAPE
	TokenType_COMMENT
	TokenType_OPEN_TUPLE
	TokenType_CLOSE_TUPLE
	TokenType_OPEN_BLOCK
	TokenType_CLOSE_BLOCK
	TokenType_OPEN_EXPRESSION
	TokenType_CLOSE_EXPRESSION
	TokenType_STRING_DELIMITER
	TokenType_DOLLAR
	TokenType_SEMICOLON
	TokenType_ASTERISK
)

//
// Current position in source stream
//
type sourceCursor struct {
	// Character Index (zero-indexed)
	index uint

	// Line number (zero-indexed)
	line uint

	// Column number (zero-indexed)
	column uint
}

// Return current position
func (cursor *sourceCursor) current() SourcePosition {
	return SourcePosition{
		Index:  cursor.index,
		Line:   cursor.line,
		Column: cursor.column,
	}
}

// Advance to next character.
//
// If newline is true, increment line number as well.
//
// Returns the previous index.
func (cursor *sourceCursor) Next(newline bool) uint {
	if newline {
		cursor.line++
		cursor.column = 0
	} else {
		cursor.column++
	}
	cursor.index += 1
	return cursor.index - 1
}

//
// Helena token
//
type Token struct {
	// Token type
	Type TokenType

	// Position in source stream
	Position SourcePosition

	// Raw Sequence of characters from stream
	Sequence string

	// String Literal
	Literal string
}

//
// Helena tokenizer
//
// This class transforms a stream of characters into a stream of tokens
//
type Tokenizer struct {
	// Input stream
	input SourceStream

	// Current token
	currentToken *Token
}

// Tokenize and return a Helena source string into a token array
func (tokenizer Tokenizer) Tokenize(source string) []Token {
	input := NewStringStream(source)
	output := NewArrayTokenStream([]Token{})
	tokenizer.TokenizeStream(input, output)
	return output.tokens
}

// Tokenize the Helena input source stream into the output token stream
func (tokenizer *Tokenizer) TokenizeStream(input SourceStream, output TokenStream) {
	tokenizer.Begin(input)
	for !tokenizer.End() {
		emittedToken := tokenizer.Next()
		if emittedToken != nil {
			output.emit(*emittedToken)
		}
	}
}

// Start incremental tokenization of a Helena input source stream
func (tokenizer *Tokenizer) Begin(input SourceStream) {
	tokenizer.input = input
	tokenizer.currentToken = nil
}

// Report whether tokenization is done
func (tokenizer *Tokenizer) End() bool {
	return tokenizer.input.end() && (tokenizer.currentToken == nil)
}

// Return current token and advance to next one
func (tokenizer *Tokenizer) Next() *Token {
	for !tokenizer.input.end() {
		position := tokenizer.input.currentPosition()
		c := tokenizer.input.next()
		var emittedToken *Token
		switch c {
		// Whitespaces
		case ' ',
			'\t',
			'\r',
			'\f':
			for !tokenizer.input.end() && isWhitespace(tokenizer.input.current()) {
				tokenizer.input.next()
			}
			emittedToken = tokenizer.addToken(TokenType_WHITESPACE, position, nil)

		// Newline
		case '\n':
			emittedToken = tokenizer.addToken(TokenType_NEWLINE, position, nil)

		// Escape sequence
		case '\\':
			{
				if tokenizer.input.end() {
					emittedToken = tokenizer.addToken(TokenType_TEXT, position, nil)
					break
				}
				e := tokenizer.input.next()
				if e == '\n' {
					// Continuation, eat up all subsequent whitespaces
					for !tokenizer.input.end() && isWhitespace(tokenizer.input.current()) {
						tokenizer.input.next()
					}
					literal := " "
					emittedToken = tokenizer.addToken(TokenType_CONTINUATION, position, &literal)
					break
				}
				escape := string(e) // Default value for unrecognized sequences
				if isEscape(e) {
					escape = string(getEscape(e))
				} else if isOctal(e) {
					codepoint := rune(e - '0')
					for i := 1; !tokenizer.input.end() &&
						isOctal(tokenizer.input.current()) &&
						i < 3; i += 1 {
						codepoint *= 8
						codepoint += rune(digitValue(tokenizer.input.current()))
						tokenizer.input.next()
					}
					escape = string(rune(codepoint))
				} else if e == 'x' {
					codepoint := rune(0)
					i := 0
					for ; !tokenizer.input.end() &&
						isHexadecimal(tokenizer.input.current()) &&
						i < 2; i += 1 {
						codepoint *= 16
						codepoint += rune(digitValue(tokenizer.input.current()))
						tokenizer.input.next()
					}
					if i > 0 {
						escape = string(rune(codepoint))
					}
				} else if e == 'u' {
					codepoint := rune(0)
					i := 0
					for ; !tokenizer.input.end() &&
						isHexadecimal(tokenizer.input.current()) &&
						i < 4; i += 1 {
						codepoint *= 16
						codepoint += rune(digitValue(tokenizer.input.current()))
						tokenizer.input.next()
					}
					if i > 0 {
						escape = string(rune(codepoint))
					}
				} else if e == 'U' {
					codepoint := rune(0)
					i := 0
					for ; !tokenizer.input.end() &&
						isHexadecimal(tokenizer.input.current()) &&
						i < 8; i += 1 {
						codepoint *= 16
						codepoint += rune(digitValue(tokenizer.input.current()))
						tokenizer.input.next()
					}
					if i > 0 {
						escape = string(rune(codepoint))
					}
				}
				emittedToken = tokenizer.addToken(TokenType_ESCAPE, position, &escape)
			}

		// Comment
		case '#':
			for !tokenizer.input.end() && tokenizer.input.current() == '#' {
				tokenizer.input.next()
			}
			emittedToken = tokenizer.addToken(TokenType_COMMENT, position, nil)

		// Tuple delimiters
		case '(':
			emittedToken = tokenizer.addToken(TokenType_OPEN_TUPLE, position, nil)

		case ')':
			emittedToken = tokenizer.addToken(TokenType_CLOSE_TUPLE, position, nil)

		// Block delimiters
		case '{':
			emittedToken = tokenizer.addToken(TokenType_OPEN_BLOCK, position, nil)

		case '}':
			emittedToken = tokenizer.addToken(TokenType_CLOSE_BLOCK, position, nil)

		// Expression delimiters
		case '[':
			emittedToken = tokenizer.addToken(TokenType_OPEN_EXPRESSION, position, nil)

		case ']':
			emittedToken = tokenizer.addToken(TokenType_CLOSE_EXPRESSION, position, nil)

		// String delimiter
		case '"':
			for !tokenizer.input.end() && tokenizer.input.current() == '"' {
				tokenizer.input.next()
			}
			emittedToken = tokenizer.addToken(TokenType_STRING_DELIMITER, position, nil)

		// Dollar
		case '$':
			emittedToken = tokenizer.addToken(TokenType_DOLLAR, position, nil)

		// Semicolon
		case ';':
			emittedToken = tokenizer.addToken(TokenType_SEMICOLON, position, nil)

		// Asterisk
		case '*':
			emittedToken = tokenizer.addToken(TokenType_ASTERISK, position, nil)

		default:
			emittedToken = tokenizer.addText(position)
		}
		if emittedToken != nil {
			return emittedToken
		}
	}
	return tokenizer.emitToken()
}

// Emit and return current token or nil
func (tokenizer *Tokenizer) emitToken() *Token {
	emitted := tokenizer.currentToken
	tokenizer.currentToken = nil
	return emitted
}

// Add and return token of given type and starting at position to result.
//
// The token can have an optional literal value.
func (tokenizer *Tokenizer) addToken(
	type_ TokenType,
	position SourcePosition,
	literal *string,
) *Token {
	emitted := tokenizer.emitToken()
	sequence := tokenizer.input.range_(
		position.Index,
		tokenizer.input.currentIndex(),
	)
	var lit string
	if literal != nil {
		lit = *literal
	} else {
		lit = sequence
	}
	tokenizer.currentToken = &Token{
		type_,
		position,
		sequence,
		lit,
	}
	return emitted
}

// Add character sequence to new or existing text token
//
// Added character sequence is between given position and current stream
// position
func (tokenizer *Tokenizer) addText(position SourcePosition) *Token {
	literal := tokenizer.input.range_(position.Index, tokenizer.input.currentIndex())
	if tokenizer.currentToken == nil || tokenizer.currentToken.Type != TokenType_TEXT {
		return tokenizer.addToken(TokenType_TEXT, position, &literal)
	} else {
		tokenizer.currentToken.Literal += literal
		tokenizer.currentToken.Sequence = tokenizer.input.range_(
			tokenizer.currentToken.Position.Index,
			tokenizer.input.currentIndex(),
		)
		return nil
	}
}

// Report whether character c is a whitespace (excluding newlines)
func isWhitespace(c byte) bool {
	switch c {
	case ' ', '\t', '\r', '\x0C':
		return true
	default:
		return false
	}
}

// Report whether character c is a known escape
func isEscape(c byte) bool {
	switch c {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\':
		return true
	default:
		return false
	}
}

// Return escaped rune from character c
func getEscape(c byte) rune {
	switch c {
	case 'a':
		return '\x07'
	case 'b':
		return '\b'
	case 'f':
		return '\f'
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	case 'v':
		return '\v'
	case '\\':
		return '\\'
	default:
		panic("unreachable")
	}
}

// Report whether character c is octal
func isOctal(c byte) bool {
	return (c >= '0' && c <= '7')
}

// Report whether character c is hexadecimal
func isHexadecimal(c byte) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'f') ||
		(c >= 'A' && c <= 'F')
}

/// Get digit value from character c
func digitValue(c byte) rune {
	switch {
	case c >= '0' && c <= '9':
		return rune(c - '0')
	case c >= 'a' && c <= 'f':
		return rune(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return rune(c - 'A' + 10)
	default:
		panic("unreachable")
	}
}

//
// Source stream (input)
//
type SourceStream interface {
	// Report whether stream is at end
	end() bool

	// Advance to next character and return character at previous position
	next() byte

	// Get current character
	current() byte

	// Get range of characters between start (inclusive) and end (exclusive)
	range_(start uint, end uint) string

	// Get current character index
	currentIndex() uint

	// Get current character position
	currentPosition() SourcePosition
}

//
// String-based character stream
//
type StringStream struct {
	// Source string
	source string

	// Current input cursor in stream
	cursor sourceCursor
}

// Create a new stream from the source string
func NewStringStream(source string) *StringStream {
	return &StringStream{source, sourceCursor{}}
}

// Report whether stream is at end
func (stream *StringStream) end() bool {
	return stream.cursor.index >= uint(len(stream.source))
}

// Advance to next character and return character at previous position
func (stream *StringStream) next() byte {
	return stream.source[stream.cursor.Next(stream.current() == '\n')]
}

// Get current character
func (stream *StringStream) current() byte {
	return stream.source[stream.cursor.index]
}

// Get range of characters between start (inclusive) and end (exclusive)
func (stream *StringStream) range_(start uint, end uint) string {

	return stream.source[start:end]
}

// Get current character index
func (stream *StringStream) currentIndex() uint {
	return stream.cursor.index
}

// Get current character position
func (stream *StringStream) currentPosition() SourcePosition {
	return stream.cursor.current()
}

//
// Token stream (input/output)
//
type TokenStream interface {
	// Emit (add) token to end of stream
	emit(token Token)

	// Reports whether stream is at end
	end() bool

	// Advance to next token and return token at previous position
	next() Token

	// Get current token
	current() Token

	// Get range of tokens between start (inclusive) and end (exclusive)
	range_(start uint, end uint) []Token

	// Get current token index
	currentIndex() uint
}

//
// Array-based token stream
//
type ArrayTokenStream struct {
	// Emitted tokens
	tokens []Token

	// Current input position in stream
	index uint
}

// Create a new stream from from the tokens array
func NewArrayTokenStream(tokens []Token) *ArrayTokenStream {
	return &ArrayTokenStream{tokens, 0}
}

// Emit (add) token to end of stream
func (stream *ArrayTokenStream) emit(token Token) {
	stream.tokens = append(stream.tokens, token)
}

// Reports whether stream is at end
func (stream ArrayTokenStream) end() bool {
	return stream.index >= uint(len(stream.tokens))
}

// Advance to next token and return token at previous position
func (stream *ArrayTokenStream) next() Token {
	stream.index += 1
	return stream.tokens[stream.index-1]
}

// Get current token
func (stream ArrayTokenStream) current() Token {
	return stream.tokens[stream.index]
}

// Get range of tokens between start (inclusive) and end (exclusive)
func (stream ArrayTokenStream) range_(start uint, end uint) []Token {
	return stream.tokens[start:end]
}

// Get current token index
func (stream ArrayTokenStream) currentIndex() uint {
	return stream.index
}
