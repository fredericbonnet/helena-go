//
// Helena parsing and AST generation
//

package core

//
// Parsing context
//
// As the parser is non-recursive, it uses a context object to hold parsing
// data along with the generated AST instead of relying on the call stack
//
type context struct {
	// Parent parsing context
	parentContext *context

	// Morpheme the context belongs to
	node morphemeNode

	// Current script
	script *scriptNode

	// Current sentence (if any)
	sentence *sentenceNode

	// Current word (if any)
	word *wordNode

	// Current morphemes (if any)
	morphemes *[]morphemeNode

	// 3-state mode during substitution
	substitutionMode substitutionMode
}
type substitutionMode int8

const (
	substitutionMode_INITIAL substitutionMode = iota
	substitutionMode_EXPECT_SOURCE
	substitutionMode_EXPECT_SELECTOR
)

// Return current morpheme (if any)
func (context *context) currentMorpheme() morphemeNode {
	if context.morphemes == nil || len(*context.morphemes) == 0 {
		return nil
	}
	return (*context.morphemes)[len(*context.morphemes)-1]
}

// Script AST node
type scriptNode struct {
	sentences  []*sentenceNode
	firstToken Token
}

func newScriptNode(firstToken Token) *scriptNode {
	return &scriptNode{firstToken: firstToken}
}

func (node scriptNode) toScript(capturePosition bool) *Script {
	script := &Script{}
	script.Sentences = make([]Sentence, len(node.sentences))
	if capturePosition {
		script.Position = &node.firstToken.Position
	}
	for i, sentence := range node.sentences {
		script.Sentences[i] = sentence.toSentence(capturePosition)
	}
	return script
}

// Sentence AST node
type sentenceNode struct {
	words      []*wordNode
	firstToken Token
}

func newSentenceNode(firstToken Token) *sentenceNode {
	return &sentenceNode{firstToken: firstToken}
}

func (node sentenceNode) toSentence(capturePosition bool) Sentence {
	sentence := Sentence{}
	sentence.Words = make([]WordOrValue, len(node.words))
	if capturePosition {
		sentence.Position = &node.firstToken.Position
	}
	for i, word := range node.words {
		sentence.Words[i].Word = word.toWord(capturePosition)
	}
	return sentence
}

// Word AST node
type wordNode struct {
	morphemes  []morphemeNode
	firstToken Token
}

func newWordNode(firstToken Token) *wordNode {
	return &wordNode{firstToken: firstToken}
}

func (node wordNode) toWord(capturePosition bool) Word {
	word := Word{}
	word.Morphemes = make([]Morpheme, len(node.morphemes))
	if capturePosition {
		word.Position = &node.firstToken.Position
	}
	for i, morpheme := range node.morphemes {
		word.Morphemes[i] = morpheme.toMorpheme(capturePosition)
	}
	return word
}

// Morpheme AST node
type morphemeNode interface {
	// Create morpheme from node
	toMorpheme(capturePosition bool) Morpheme
}

// Literal morpheme AST node
type literalNode struct {
	firstToken Token
	value      string
}

func newLiteralNode(firstToken Token, value string) *literalNode {
	return &literalNode{
		firstToken: firstToken,
		value:      value,
	}
}
func (node *literalNode) toMorpheme(capturePosition bool) Morpheme {
	var position *SourcePosition
	if capturePosition {
		position = &node.firstToken.Position
	}
	return LiteralMorpheme{
		Value:    node.value,
		Position: position,
	}
}

// Tuple morpheme AST node
type tupleNode struct {
	firstToken Token
	subscript  *scriptNode
}

func newTupleNode(firstToken Token) *tupleNode {
	return &tupleNode{
		firstToken: firstToken,
		subscript:  newScriptNode(firstToken),
	}
}
func (node *tupleNode) toMorpheme(capturePosition bool) Morpheme {
	var position *SourcePosition
	if capturePosition {
		position = &node.firstToken.Position
	}
	return TupleMorpheme{
		Subscript: *node.subscript.toScript(capturePosition),
		Position:  position,
	}
}

// Block morpheme AST node
type blockNode struct {
	firstToken Token
	subscript  *scriptNode
	value      string

	// Starting position of block, used to get literal value
	start uint
}

func newBlockNode(firstToken Token, start uint) *blockNode {
	return &blockNode{
		firstToken: firstToken,
		subscript:  newScriptNode(firstToken),
		start:      start,
	}
}
func (node *blockNode) toMorpheme(capturePosition bool) Morpheme {
	var position *SourcePosition
	if capturePosition {
		position = &node.firstToken.Position
	}
	return BlockMorpheme{
		Subscript: *node.subscript.toScript(capturePosition),
		Value:     node.value,
		Position:  position,
	}
}

// Expression morpheme AST node
type expressionNode struct {
	firstToken Token
	subscript  *scriptNode
}

func newExpressionNode(firstToken Token) *expressionNode {
	return &expressionNode{
		firstToken: firstToken,
		subscript:  newScriptNode(firstToken),
	}
}
func (node *expressionNode) toMorpheme(capturePosition bool) Morpheme {
	var position *SourcePosition
	if capturePosition {
		position = &node.firstToken.Position
	}
	return ExpressionMorpheme{
		Subscript: *node.subscript.toScript(capturePosition),
		Position:  position,
	}
}

// String morpheme AST node
type stringNode struct {
	firstToken Token
	morphemes  []morphemeNode
}

func newStringNode(firstToken Token) *stringNode {
	return &stringNode{
		firstToken: firstToken,
	}
}
func (node *stringNode) toMorpheme(capturePosition bool) Morpheme {
	morpheme := StringMorpheme{}
	morpheme.Morphemes = make([]Morpheme, len(node.morphemes))
	if capturePosition {
		morpheme.Position = &node.firstToken.Position
	}
	for i, child := range node.morphemes {
		morpheme.Morphemes[i] = child.toMorpheme(capturePosition)
	}
	return morpheme
}

// Here-string morpheme AST node
type hereStringNode struct {
	firstToken      Token
	value           string
	delimiterLength uint
}

func newHereStringNode(firstToken Token, delimiter string) *hereStringNode {
	return &hereStringNode{
		firstToken:      firstToken,
		delimiterLength: uint(len(delimiter)),
	}
}
func (node *hereStringNode) toMorpheme(capturePosition bool) Morpheme {
	var position *SourcePosition
	if capturePosition {
		position = &node.firstToken.Position
	}
	return HereStringMorpheme{
		Value:           node.value,
		DelimiterLength: node.delimiterLength,
		Position:        position,
	}
}

// Tagged string morpheme AST node
type taggedStringNode struct {
	firstToken Token
	value      string
	tag        string
}

func newTaggedStringNode(firstToken Token, tag string) *taggedStringNode {
	return &taggedStringNode{
		firstToken: firstToken,
		tag:        tag,
	}
}
func (node *taggedStringNode) toMorpheme(capturePosition bool) Morpheme {
	// Shift lines by prefix length
	// - First find prefix length = length of last line
	i := len(node.value)
	for ; i > 0; i -= 1 {
		if node.value[i-1] == '\n' {
			break
		}
	}
	prefixLen := len(node.value) - i
	// - Then append all lines with prefix removed
	value := ""
	start := 0
	for j, c := range node.value {
		if c == '\n' {
			value += node.value[min(start+prefixLen, j) : j+1]
			start = j + 1
		}
	}
	value += node.value[start+prefixLen:]

	var position *SourcePosition
	if capturePosition {
		position = &node.firstToken.Position
	}
	return TaggedStringMorpheme{
		Value:    value,
		Tag:      node.tag,
		Position: position,
	}
}

// Line comment morpheme AST node
type lineCommentNode struct {
	firstToken      Token
	value           string
	delimiterLength uint
}

func newLineCommentNode(firstToken Token, delimiter string) *lineCommentNode {
	return &lineCommentNode{
		firstToken:      firstToken,
		delimiterLength: uint(len(delimiter)),
	}
}
func (node *lineCommentNode) toMorpheme(capturePosition bool) Morpheme {
	var position *SourcePosition
	if capturePosition {
		position = &node.firstToken.Position
	}
	return LineCommentMorpheme{
		Value:           node.value,
		DelimiterLength: node.delimiterLength,
		Position:        position,
	}
}

// Block comment morpheme AST node
type blockCommentNode struct {
	firstToken      Token
	value           string
	delimiterLength uint

	// Nesting level, node is closed when it reaches zero
	nesting uint
}

func newBlockCommentNode(firstToken Token, delimiter string) *blockCommentNode {
	return &blockCommentNode{
		firstToken:      firstToken,
		delimiterLength: uint(len(delimiter)),
		nesting:         1,
	}
}
func (node *blockCommentNode) toMorpheme(capturePosition bool) Morpheme {
	var position *SourcePosition
	if capturePosition {
		position = &node.firstToken.Position
	}
	return BlockCommentMorpheme{
		Value:           node.value,
		DelimiterLength: node.delimiterLength,
		Position:        position,
	}
}

// Substitute Next morpheme AST node
type substituteNextNode struct {
	firstToken Token
	expansion  bool
	value      string
}

func newSubstituteNextNode(firstToken Token, value string) *substituteNextNode {
	return &substituteNextNode{
		firstToken: firstToken,
		expansion:  false,
		value:      value,
	}
}

func (node *substituteNextNode) toMorpheme(capturePosition bool) Morpheme {
	var position *SourcePosition
	if capturePosition {
		position = &node.firstToken.Position
	}
	return SubstituteNextMorpheme{
		Expansion: node.expansion,
		Value:     node.value,
		Position:  position,
	}
}

// Parse result
type ParseResult struct {
	// Success flag
	Success bool

	// Parsed Script on success
	Script *Script

	// Error Message
	Message string
}

// Helpers

func PARSE_OK(script *Script) ParseResult    { return ParseResult{Success: true, Script: script} }
func PARSE_ERROR(message string) ParseResult { return ParseResult{Success: false, Message: message} }

//
// Parser options
//
type ParserOptions struct {
	// Whether to capture morpheme positions
	CapturePositions bool
}

//
// Helena parser
//
// This class transforms a stream of tokens into an abstract syntax tree
//
type Parser struct {
	// Input stream
	stream TokenStream

	// Current context
	context *context

	/** Parser options */
	options ParserOptions
}

func NewParser(options *ParserOptions) *Parser {
	if options == nil {
		return &Parser{options: ParserOptions{false}}
	} else {
		return &Parser{options: *options}
	}
}

// Parse an array of tokens
func (parser *Parser) Parse(tokens []Token) ParseResult {
	stream := NewArrayTokenStream(tokens)
	parser.begin(stream)
	for !parser.end() {
		result := parser.next()
		if !result.Success {
			return result
		}
	}
	return parser.CloseStream()
}

// Parse a token stream till the end
//
// This method is useful when parsing incomplete scripts in interactive mode,
// as getting an error at this stage is unrecoverable even if there is more
// input to parse
func (parser *Parser) ParseStream(stream TokenStream) ParseResult {
	parser.begin(stream)
	for !parser.end() {
		result := parser.next()
		if !result.Success {
			return result
		}
	}
	return PARSE_OK(nil)
}

// Start incremental parsing of a Helena token stream
func (parser *Parser) begin(stream TokenStream) {
	var firstToken Token
	if !stream.end() {
		firstToken = stream.current()
	}
	parser.context = &context{
		script: newScriptNode(firstToken),
	}
	parser.stream = stream
}

// Report whether parsing is done
func (parser *Parser) end() bool {
	return parser.stream.end()
}

// Parse current token and advance to next one
func (parser *Parser) next() ParseResult {
	token := parser.stream.next()
	return parser.parseToken(token)
}

// Close the current token stream and return parse result
//
// This method is useful when testing for script completeness in interactive
// mode and prompt for more input
func (parser *Parser) CloseStream() ParseResult {
	if parser.context.node != nil {
		switch (parser.context.node).(type) {
		case *tupleNode:
			return PARSE_ERROR("unmatched left parenthesis")
		case *blockNode:
			return PARSE_ERROR("unmatched left brace")
		case *expressionNode:
			return PARSE_ERROR("unmatched left bracket")
		case *stringNode:
			return PARSE_ERROR("unmatched string delimiter")
		case *hereStringNode:
			return PARSE_ERROR("unmatched here-string delimiter")
		case *taggedStringNode:
			return PARSE_ERROR("unmatched tagged string delimiter")
		case *lineCommentNode:
			parser.closeLineComment()
		case *blockCommentNode:
			return PARSE_ERROR("unmatched block comment delimiter")
		default:
			return PARSE_ERROR("unterminated script")
		}
	}
	parser.closeSentence()

	return PARSE_OK(parser.context.script.toScript(parser.options.CapturePositions))
}

// Parse a single token
func (parser *Parser) parseToken(token Token) ParseResult {
	if parser.context.node != nil {
		switch (parser.context.node).(type) {
		case *tupleNode:
			return parser.parseTuple(token)

		case *blockNode:
			return parser.parseBlock(token)

		case *expressionNode:
			return parser.parseExpression(token)

		case *stringNode:
			return parser.parseString(token)

		case *hereStringNode:
			return parser.parseHereString(token)

		case *taggedStringNode:
			return parser.parseTaggedString(token)

		case *lineCommentNode:
			return parser.parseLineComment(token)

		case *blockCommentNode:
			return parser.parseBlockComment(token)
		}

	}
	return parser.parseScript(token)
}

//
// Context management
//

// Push a new context for a contextual node
func (parser *Parser) pushContext(node morphemeNode, ctx context) {
	ctx.node = node
	ctx.parentContext = parser.context
	parser.context = &ctx
}

// Pop the existing context and return to its parent
func (parser *Parser) popContext() {
	parser.context = parser.context.parentContext
}

//
// Scripts
//

func (parser *Parser) parseScript(token Token) ParseResult {
	switch token.Type {
	case TokenType_CLOSE_TUPLE:
		return PARSE_ERROR("unmatched right parenthesis")

	case TokenType_CLOSE_BLOCK:
		return PARSE_ERROR("unmatched right brace")

	case TokenType_CLOSE_EXPRESSION:
		return PARSE_ERROR("unmatched right bracket")

	default:
		return parser.parseWord(token)
	}
}

//
// Tuples
//

func (parser *Parser) parseTuple(token Token) ParseResult {
	switch token.Type {
	case TokenType_CLOSE_TUPLE:
		parser.closeTuple()
		if parser.expectSource() {
			parser.continueSubstitution()
		}
		return PARSE_OK(nil)

	default:
		return parser.parseWord(token)
	}
}

// Open a tuple parsing context
func (parser *Parser) openTuple(token Token) {
	node := newTupleNode(token)
	*parser.context.morphemes = append(*parser.context.morphemes, node)
	parser.pushContext(node, context{
		script: node.subscript,
	})
}

// Close the tuple parsing context
func (parser *Parser) closeTuple() {
	parser.popContext()
}

//
// Blocks
//

func (parser *Parser) parseBlock(token Token) ParseResult {
	switch token.Type {
	case TokenType_CLOSE_BLOCK:
		parser.closeBlock()
		if parser.expectSource() {
			parser.continueSubstitution()
		}
		return PARSE_OK(nil)

	default:
		return parser.parseWord(token)
	}
}

// Open a block parsing context
func (parser *Parser) openBlock(token Token) {
	node := newBlockNode(token, parser.stream.currentIndex())
	*parser.context.morphemes = append(*parser.context.morphemes, node)
	parser.pushContext(node, context{
		script: node.subscript,
	})
}

// Close the block parsing context
func (parser *Parser) closeBlock() {
	node := (parser.context.node).(*blockNode)
	range_ := parser.stream.range_(node.start, parser.stream.currentIndex()-1)
	value := ""
	for _, token := range range_ {
		value += token.Literal
	}
	node.value = value
	parser.popContext()
}

//
// Expressions
//

func (parser *Parser) parseExpression(token Token) ParseResult {
	switch token.Type {
	case TokenType_CLOSE_EXPRESSION:
		parser.closeExpression()
		parser.continueSubstitution()
		return PARSE_OK(nil)

	default:
		return parser.parseWord(token)
	}
}

// Open an expression parsing context
func (parser *Parser) openExpression(token Token) {
	node := newExpressionNode(token)
	*parser.context.morphemes = append(*parser.context.morphemes, node)
	parser.pushContext(node, context{
		script: node.subscript,
	})
}

// Close the expression parsing context
func (parser *Parser) closeExpression() {
	parser.popContext()
}

//
// Words
//

func (parser *Parser) parseWord(token Token) ParseResult {
	switch token.Type {
	case TokenType_WHITESPACE,
		TokenType_CONTINUATION:
		parser.closeWord()
		return PARSE_OK(nil)

	case TokenType_NEWLINE,
		TokenType_SEMICOLON:
		parser.closeSentence()
		return PARSE_OK(nil)

	case TokenType_TEXT,
		TokenType_ESCAPE:
		parser.ensureWord(token)
		parser.addLiteral(token, token.Literal)
		return PARSE_OK(nil)

	case TokenType_STRING_DELIMITER:
		if !parser.ensureWord(token) {
			return PARSE_ERROR("unexpected string delimiter")
		}
		if len(token.Literal) == 1 {
			// Regular strings
			parser.openString(token)
		} else if len(token.Literal) == 2 {
			if !parser.stream.end() && parser.stream.current().Type == TokenType_TEXT {
				// Tagged strings
				next := parser.stream.current()
				parser.openTaggedString(token, next.Literal)
			} else {
				// Special case for empty strings
				parser.openString(token)
				parser.closeString()
			}
		} else {
			// Here-strings
			parser.openHereString(token, token.Literal)
		}
		return PARSE_OK(nil)

	case TokenType_OPEN_TUPLE:
		parser.ensureWord(token)
		parser.openTuple(token)
		return PARSE_OK(nil)

	case TokenType_OPEN_BLOCK:
		parser.ensureWord(token)
		parser.openBlock(token)
		return PARSE_OK(nil)

	case TokenType_OPEN_EXPRESSION:
		parser.ensureWord(token)
		parser.openExpression(token)
		return PARSE_OK(nil)

	case TokenType_COMMENT:
		if parser.expectSource() {
			return PARSE_ERROR("unexpected comment delimiter")
		}
		if !parser.ensureWord(token) {
			parser.addLiteral(token, token.Literal)
			return PARSE_OK(nil)
		}
		if !parser.openBlockComment(token, token.Literal, false) {
			parser.openLineComment(token, token.Literal)
		}
		return PARSE_OK(nil)

	case TokenType_DOLLAR:
		parser.ensureWord(token)
		parser.beginSubstitution(token, token.Literal)
		return PARSE_OK(nil)

	case TokenType_ASTERISK:
		parser.ensureWord(token)
		parser.addLiteral(token, token.Literal)
		return PARSE_OK(nil)

	case TokenType_CLOSE_TUPLE:
		return PARSE_ERROR("mismatched right parenthesis")

	case TokenType_CLOSE_BLOCK:
		return PARSE_ERROR("mismatched right brace")

	case TokenType_CLOSE_EXPRESSION:
		return PARSE_ERROR("mismatched right bracket")

	default:
		return PARSE_ERROR("syntax error")
	}
}

// Ensure that word-related context info exists.
//
// Returns false if the word context already exists, true if it has been created
func (parser *Parser) ensureWord(token Token) bool {
	if parser.context.word != nil {
		return false
	}
	if parser.context.sentence == nil {
		parser.context.sentence = newSentenceNode(token)
		parser.context.script.sentences = append(parser.context.script.sentences, parser.context.sentence)
	}
	parser.context.word = newWordNode(token)
	parser.context.sentence.words = append(parser.context.sentence.words, parser.context.word)
	parser.context.morphemes = &parser.context.word.morphemes
	return true
}

// Attempt to merge consecutive, non substituted literals
func (parser *Parser) addLiteral(token Token, value string) {
	current := parser.context.currentMorpheme()
	if current != nil {
		if morpheme, ok := current.(*literalNode); ok && !parser.withinSubstitution() {
			morpheme.value += value
			(*parser.context.morphemes)[len(*parser.context.morphemes)-1] = morpheme
			return
		}
	}
	morpheme := newLiteralNode(token, value)
	*parser.context.morphemes = append(*parser.context.morphemes, morpheme)
	parser.continueSubstitution()
}

// Close the current word
func (parser *Parser) closeWord() {
	parser.endSubstitution()
	parser.context.word = nil
}

// Close the current sentence
func (parser *Parser) closeSentence() {
	parser.closeWord()
	parser.context.sentence = nil
}

//
// Strings
//

func (parser *Parser) parseString(token Token) ParseResult {
	if parser.expectSource() {
		switch token.Type {
		case TokenType_TEXT,
			TokenType_DOLLAR,
			TokenType_OPEN_BLOCK,
			TokenType_OPEN_EXPRESSION:

		default:
			parser.endSubstitution()
		}
	}
	switch token.Type {
	case TokenType_DOLLAR:
		parser.beginSubstitution(token, token.Literal)

	case TokenType_STRING_DELIMITER:
		if len(token.Literal) != 1 {
			return PARSE_ERROR("extra characters after string delimiter")
		}
		parser.closeString()

	case TokenType_OPEN_TUPLE:
		if parser.withinSubstitution() {
			parser.openTuple(token)
		} else {
			parser.addLiteral(token, token.Literal)
		}

	case TokenType_OPEN_BLOCK:
		if parser.withinSubstitution() {
			parser.openBlock(token)
		} else {
			parser.addLiteral(token, token.Literal)
		}

	case TokenType_OPEN_EXPRESSION:
		parser.openExpression(token)

	default:
		parser.addLiteral(token, token.Literal)
	}
	return PARSE_OK(nil)
}

// Open a string parsing context
func (parser *Parser) openString(token Token) {
	node := newStringNode(token)
	*parser.context.morphemes = append(*parser.context.morphemes, node)
	parser.pushContext(node, context{
		morphemes: &node.morphemes,
	})
}

/** Close the string parsing context */
func (parser *Parser) closeString() {
	parser.endSubstitution()
	parser.popContext()
}

//
// Here-strings
//

func (parser *Parser) parseHereString(token Token) ParseResult {
	if token.Type == TokenType_STRING_DELIMITER &&
		parser.closeHereString(token.Literal) {
		return PARSE_OK(nil)
	}
	parser.addHereStringSequence(token.Sequence)
	return PARSE_OK(nil)
}

// Open a here-string parsing context
func (parser *Parser) openHereString(token Token, delimiter string) {
	node := newHereStringNode(token, delimiter)
	*parser.context.morphemes = append(*parser.context.morphemes, node)
	parser.pushContext(node, context{})
}

// Attempt to close the here-string parsing context, report whether the context is closed
func (parser *Parser) closeHereString(delimiter string) bool {
	node := parser.context.node.(*hereStringNode)
	if uint(len(delimiter)) != node.delimiterLength {
		return false
	}
	parser.popContext()
	return true
}

// Append sequence to current here-string
func (parser *Parser) addHereStringSequence(value string) {
	node := parser.context.node.(*hereStringNode)
	node.value += value
}

//
// Tagged strings
//

func (parser *Parser) parseTaggedString(token Token) ParseResult {
	if token.Type == TokenType_TEXT && parser.closeTaggedString(token.Literal) {
		return PARSE_OK(nil)
	}
	parser.addTaggedStringSequence(token.Sequence)
	return PARSE_OK(nil)
}

// Open a tagged string parsing context
func (parser *Parser) openTaggedString(token Token, tag string) {
	node := newTaggedStringNode(token, tag)
	*parser.context.morphemes = append(*parser.context.morphemes, node)
	parser.pushContext(node, context{})

	// Discard everything until the next newline
	for !parser.stream.end() && parser.stream.next().Type != TokenType_NEWLINE {
	}
}

// Attempt to close the tagged string parsing context , report whether the context is closed
//
// The literal must must match open tag
func (parser *Parser) closeTaggedString(literal string) bool {
	node := parser.context.node.(*taggedStringNode)
	if literal != node.tag {
		return false
	}
	next := parser.stream.current()
	if next.Type != TokenType_STRING_DELIMITER {
		return false
	}
	if len(next.Literal) != 2 {
		return false
	}
	parser.stream.next()

	parser.popContext()
	return true
}

// Append sequence to current tagged string
func (parser *Parser) addTaggedStringSequence(value string) {
	node := parser.context.node.(*taggedStringNode)
	node.value += value
}

//
// Line comments
//

func (parser *Parser) parseLineComment(token Token) ParseResult {
	switch token.Type {
	case TokenType_NEWLINE:
		parser.closeLineComment()
		parser.closeSentence()

	default:
		parser.addLineCommentSequence(token.Literal)
	}
	return PARSE_OK(nil)
}

// Open a line comment parsing context
func (parser *Parser) openLineComment(token Token, delimiter string) {
	node := newLineCommentNode(token, delimiter)
	*parser.context.morphemes = append(*parser.context.morphemes, node)
	parser.pushContext(node, context{})
}

// Close the line comment parsing context
func (parser *Parser) closeLineComment() {
	parser.popContext()
}

// Append sequence to current line comment
func (parser *Parser) addLineCommentSequence(value string) {
	node := parser.context.node.(*lineCommentNode)
	node.value += value
}

//
// Block comments
//

func (parser *Parser) parseBlockComment(token Token) ParseResult {
	switch token.Type {
	case TokenType_COMMENT:
		if !parser.openBlockComment(token, token.Literal, true) {
			parser.addBlockCommentSequence(token.Sequence)
		}

	case TokenType_CLOSE_BLOCK:
		if !parser.closeBlockComment() {
			parser.addBlockCommentSequence(token.Sequence)
		}

	default:
		parser.addBlockCommentSequence(token.Sequence)
	}
	return PARSE_OK(nil)
}

// Attempt to open a block comment parsing context, report whether the context was open
func (parser *Parser) openBlockComment(token Token, delimiter string, nested bool) bool {
	if parser.stream.end() || parser.stream.current().Type != TokenType_OPEN_BLOCK {
		return false
	}
	if nested {
		node := parser.context.node.(*blockCommentNode)
		if node.delimiterLength == uint(len(delimiter)) {
			node.nesting++
		}
		return false
	}
	parser.stream.next()
	node := newBlockCommentNode(token, delimiter)
	*parser.context.morphemes = append(*parser.context.morphemes, node)
	parser.pushContext(node, context{})
	return true
}

// Attempt to close the block comment parsing context, report whether the context was closed
func (parser *Parser) closeBlockComment() bool {
	node := parser.context.node.(*blockCommentNode)
	token := parser.stream.current()
	if token.Type != TokenType_COMMENT {
		return false
	}
	if uint(len(token.Literal)) != node.delimiterLength {
		return false
	}
	node.nesting--
	if node.nesting > 0 {
		return false
	}
	parser.stream.next()
	parser.popContext()
	return true
}

// Append sequence to current block comment
func (parser *Parser) addBlockCommentSequence(value string) {
	node := parser.context.node.(*blockCommentNode)
	node.value += value
}

//
// Substitutions
//

func (parser *Parser) beginSubstitution(token Token, value string) {
	morpheme := newSubstituteNextNode(token, value)
	if !parser.stream.end() && parser.stream.current().Type == TokenType_ASTERISK {
		// Only expand the leading substitution
		current := parser.context.currentMorpheme()
		if current == nil {
			morpheme.expansion = true
		} else if _, ok := current.(*substituteNextNode); !ok {
			morpheme.expansion = true
		}
		morpheme.value += parser.stream.next().Literal
	}
	*parser.context.morphemes = append(*parser.context.morphemes, morpheme)
	parser.context.substitutionMode = substitutionMode_EXPECT_SOURCE
}

func (parser *Parser) continueSubstitution() {
	if parser.expectSource() {
		parser.context.substitutionMode = substitutionMode_EXPECT_SELECTOR
	} else {
		parser.context.substitutionMode = substitutionMode_INITIAL
	}
}
func (parser *Parser) endSubstitution() {
	if !parser.expectSource() {
		return
	}

	// Convert stale substitutions to literals
	parser.context.substitutionMode = substitutionMode_INITIAL
	var firstToken Token
	value := ""
	for {
		current := parser.context.currentMorpheme()
		if current == nil {
			break
		} else if morpheme, ok := current.(*substituteNextNode); ok {
			firstToken = morpheme.firstToken
			value = morpheme.value + value
			*parser.context.morphemes = (*parser.context.morphemes)[:len(*parser.context.morphemes)-1]
		} else {
			break
		}
	}
	parser.addLiteral(firstToken, value)
}
func (parser *Parser) withinSubstitution() bool {
	return parser.context.substitutionMode != substitutionMode_INITIAL
}

func (parser *Parser) expectSource() bool {
	return parser.context.substitutionMode == substitutionMode_EXPECT_SOURCE
}
