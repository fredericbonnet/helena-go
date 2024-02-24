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
	sentences []*sentenceNode
}

func (node scriptNode) toScript() *Script {
	script := &Script{}
	script.Sentences = make([]Sentence, len(node.sentences))
	for i, sentence := range node.sentences {
		script.Sentences[i] = sentence.toSentence()
	}
	return script
}

// Sentence AST node
type sentenceNode struct {
	words []*wordNode
}

func (node sentenceNode) toSentence() Sentence {
	sentence := Sentence{}
	sentence.Words = make([]Word, len(node.words))
	for i, word := range node.words {
		sentence.Words[i] = word.toWord()
	}
	return sentence
}

// Word AST node
type wordNode struct {
	morphemes []morphemeNode
}

func (node wordNode) toWord() Word {
	word := Word{}
	word.Morphemes = make([]Morpheme, len(node.morphemes))
	for i, morpheme := range node.morphemes {
		word.Morphemes[i] = morpheme.toMorpheme()
	}
	return word
}

// Morpheme AST node
type morphemeNode interface {
	// Create morpheme from node
	toMorpheme() Morpheme
}

// Literal morpheme AST node
type literalNode struct {
	value string
}

func newLiteralNode(value string) *literalNode {
	return &literalNode{value}
}
func (node *literalNode) toMorpheme() Morpheme {
	return LiteralMorpheme{
		Value: node.value,
	}
}

// Tuple morpheme AST node
type tupleNode struct {
	subscript *scriptNode
}

func newTupleNode() *tupleNode {
	return &tupleNode{&scriptNode{}}
}
func (node *tupleNode) toMorpheme() Morpheme {
	return TupleMorpheme{
		Subscript: *node.subscript.toScript(),
	}
}

// Block morpheme AST node
type blockNode struct {
	subscript *scriptNode
	value     string

	// Starting position of block, used to get literal value
	start uint
}

func newBlockNode(start uint) *blockNode {
	return &blockNode{
		subscript: &scriptNode{},
		start:     start,
	}
}
func (node *blockNode) toMorpheme() Morpheme {
	return BlockMorpheme{
		Subscript: *node.subscript.toScript(),
		Value:     node.value,
	}
}

// Expression morpheme AST node
type expressionNode struct {
	subscript *scriptNode
}

func newExpressionNode() *expressionNode {
	return &expressionNode{&scriptNode{}}
}
func (node *expressionNode) toMorpheme() Morpheme {
	return ExpressionMorpheme{
		Subscript: *node.subscript.toScript(),
	}
}

// String morpheme AST node
type stringNode struct {
	morphemes []morphemeNode
}

func newStringNode() *stringNode {
	return &stringNode{}
}
func (node *stringNode) toMorpheme() Morpheme {
	morpheme := StringMorpheme{}
	morpheme.Morphemes = make([]Morpheme, len(node.morphemes))
	for i, child := range node.morphemes {
		morpheme.Morphemes[i] = child.toMorpheme()
	}
	return morpheme
}

// Here-string morpheme AST node
type hereStringNode struct {
	value           string
	delimiterLength uint
}

func newHereStringNode(delimiter string) *hereStringNode {
	return &hereStringNode{
		delimiterLength: uint(len(delimiter)),
	}
}
func (node *hereStringNode) toMorpheme() Morpheme {
	return HereStringMorpheme{
		Value:           node.value,
		DelimiterLength: node.delimiterLength,
	}
}

// Tagged string morpheme AST node
type taggedStringNode struct {
	value string
	tag   string
}

func newTaggedStringNode(tag string) *taggedStringNode {
	return &taggedStringNode{
		tag: tag,
	}
}
func (node *taggedStringNode) toMorpheme() Morpheme {
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

	return TaggedStringMorpheme{
		Value: value,
		Tag:   node.tag,
	}
}

// Line comment morpheme AST node
type lineCommentNode struct {
	value           string
	delimiterLength uint
}

func newLineCommentNode(delimiter string) *lineCommentNode {
	return &lineCommentNode{
		delimiterLength: uint(len(delimiter)),
	}
}
func (node *lineCommentNode) toMorpheme() Morpheme {
	return LineCommentMorpheme{
		Value:           node.value,
		DelimiterLength: node.delimiterLength,
	}
}

// Block comment morpheme AST node
type blockCommentNode struct {
	value           string
	delimiterLength uint

	// Nesting level, node is closed when it reaches zero
	nesting uint
}

func newBlockCommentNode(delimiter string) *blockCommentNode {
	return &blockCommentNode{
		delimiterLength: uint(len(delimiter)),
		nesting:         1,
	}
}
func (node *blockCommentNode) toMorpheme() Morpheme {
	return BlockCommentMorpheme{
		Value:           node.value,
		DelimiterLength: node.delimiterLength,
	}
}

// Substitute Next morpheme AST node
type substituteNextNode struct {
	expansion bool
	levels    uint
	value     string
}

func newSubstituteNextNode(value string) *substituteNextNode {
	return &substituteNextNode{
		expansion: false,
		levels:    1,
		value:     value,
	}
}

func (node *substituteNextNode) toMorpheme() Morpheme {
	return SubstituteNextMorpheme{
		Expansion: node.expansion,
		Levels:    node.levels,
		Value:     node.value,
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
// Helena parser
//
// This class transforms a stream of tokens into an abstract syntax tree
//
type Parser struct {
	// Input stream
	stream TokenStream

	// Current context */
	context *context
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
	return parser.closeStream()
}

//   /**
//    * Parse a token stream till the end
//    *
//    * This method is useful when parsing incomplete scripts in interactive mode,
//    * as getting an error at this stage is unrecoverable even if there is more
//    * input to parse
//    *
//    * @param stream - Stream to parse
//    *
//    * @returns        Empty result on success, else error
//    */
//   parseStream(stream: TokenStream): ParseResult {
//     this.begin(stream);
//     while (!this.end()) {
//       const result = this.next();
//       if (!result.success) return result;
//     }
//     return PARSE_OK();
//   }

// Start incremental parsing of a Helena token stream
func (parser *Parser) begin(stream TokenStream) {
	parser.context = &context{
		script: &scriptNode{},
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
func (parser *Parser) closeStream() ParseResult {
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

	return PARSE_OK(parser.context.script.toScript())
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
func (parser *Parser) openTuple() {
	node := newTupleNode()
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
func (parser *Parser) openBlock() {
	node := newBlockNode(parser.stream.currentIndex())
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
func (parser *Parser) openExpression() {
	node := newExpressionNode()
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
		parser.ensureWord()
		parser.addLiteral(token.Literal)
		return PARSE_OK(nil)

	case TokenType_STRING_DELIMITER:
		if !parser.ensureWord() {
			return PARSE_ERROR("unexpected string delimiter")
		}
		if len(token.Literal) == 1 {
			// Regular strings
			parser.openString()
		} else if len(token.Literal) == 2 {
			if !parser.stream.end() && parser.stream.current().Type == TokenType_TEXT {
				// Tagged strings
				next := parser.stream.current()
				parser.openTaggedString(next.Literal)
			} else {
				// Special case for empty strings
				parser.openString()
				parser.closeString()
			}
		} else {
			// Here-strings
			parser.openHereString(token.Literal)
		}
		return PARSE_OK(nil)

	case TokenType_OPEN_TUPLE:
		parser.ensureWord()
		parser.openTuple()
		return PARSE_OK(nil)

	case TokenType_OPEN_BLOCK:
		parser.ensureWord()
		parser.openBlock()
		return PARSE_OK(nil)

	case TokenType_OPEN_EXPRESSION:
		parser.ensureWord()
		parser.openExpression()
		return PARSE_OK(nil)

	case TokenType_COMMENT:
		if parser.expectSource() {
			return PARSE_ERROR("unexpected comment delimiter")
		}
		if !parser.ensureWord() {
			parser.addLiteral(token.Literal)
			return PARSE_OK(nil)
		}
		if !parser.openBlockComment(token.Literal, false) {
			parser.openLineComment(token.Literal)
		}
		return PARSE_OK(nil)

	case TokenType_DOLLAR:
		parser.ensureWord()
		parser.beginSubstitution(token.Literal)
		return PARSE_OK(nil)

	case TokenType_ASTERISK:
		parser.ensureWord()
		parser.addLiteral(token.Literal)
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
func (parser *Parser) ensureWord() bool {
	if parser.context.word != nil {
		return false
	}
	if parser.context.sentence == nil {
		parser.context.sentence = &sentenceNode{}
		parser.context.script.sentences = append(parser.context.script.sentences, parser.context.sentence)
	}
	parser.context.word = &wordNode{}
	parser.context.sentence.words = append(parser.context.sentence.words, parser.context.word)
	parser.context.morphemes = &parser.context.word.morphemes
	return true
}

// Attempt to merge consecutive, non substituted literals
func (parser *Parser) addLiteral(value string) {
	current := parser.context.currentMorpheme()
	if current != nil {
		if morpheme, ok := current.(*literalNode); ok && !parser.withinSubstitution() {
			morpheme.value += value
			(*parser.context.morphemes)[len(*parser.context.morphemes)-1] = morpheme
			return
		}
	}
	morpheme := newLiteralNode(value)
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
		parser.beginSubstitution(token.Literal)

	case TokenType_STRING_DELIMITER:
		if len(token.Literal) != 1 {
			return PARSE_ERROR("extra characters after string delimiter")
		}
		parser.closeString()

	case TokenType_OPEN_TUPLE:
		if parser.withinSubstitution() {
			parser.openTuple()
		} else {
			parser.addLiteral(token.Literal)
		}

	case TokenType_OPEN_BLOCK:
		if parser.withinSubstitution() {
			parser.openBlock()
		} else {
			parser.addLiteral(token.Literal)
		}

	case TokenType_OPEN_EXPRESSION:
		parser.openExpression()

	default:
		parser.addLiteral(token.Literal)
	}
	return PARSE_OK(nil)
}

// Open a string parsing context
func (parser *Parser) openString() {
	node := newStringNode()
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
func (parser *Parser) openHereString(delimiter string) {
	node := newHereStringNode(delimiter)
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
func (parser *Parser) openTaggedString(tag string) {
	node := newTaggedStringNode(tag)
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
func (parser *Parser) openLineComment(delimiter string) {
	node := newLineCommentNode(delimiter)
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
		if !parser.openBlockComment(token.Literal, true) {
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
func (parser *Parser) openBlockComment(delimiter string, nested bool) bool {
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
	node := newBlockCommentNode(delimiter)
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

func (parser *Parser) beginSubstitution(value string) {
	current := parser.context.currentMorpheme()
	if current != nil {
		if morpheme, ok := current.(*substituteNextNode); ok {
			morpheme.value += value
			morpheme.levels++
			if !parser.stream.end() && parser.stream.current().Type == TokenType_ASTERISK {
				// Ignore expansion on inner substitutions
				morpheme.value += parser.stream.next().Literal
			}
			return
		}
	}
	morpheme := newSubstituteNextNode(value)
	if !parser.stream.end() && parser.stream.current().Type == TokenType_ASTERISK {
		morpheme.expansion = true
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
	value := parser.context.currentMorpheme().(*substituteNextNode).value
	*parser.context.morphemes = (*parser.context.morphemes)[:len(*parser.context.morphemes)-1]
	parser.addLiteral(value)
}
func (parser *Parser) withinSubstitution() bool {
	return parser.context.substitutionMode != substitutionMode_INITIAL
}

func (parser *Parser) expectSource() bool {
	return parser.context.substitutionMode == substitutionMode_EXPECT_SOURCE
}
func (parser *Parser) expectSelector() bool {
	return parser.context.substitutionMode == substitutionMode_EXPECT_SELECTOR
}
