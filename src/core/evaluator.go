//
// Helena script evaluation
//

package core

//
// Helena evaluator
//
type Evaluator interface {
	// Evaluate a script and return result
	EvaluateScript(script Script) Result

	// Evaluate a sentence and return result
	EvaluateSentence(sentence Sentence) Result

	// Evaluate a word and return result
	EvaluateWord(word Word) Result
}

//
// Helena compiling evaluator
//
// This class compiles scripts to programs before executing them in an
// encapsulated {@link Executor}
//
type CompilingEvaluator struct {
	// Compiler used for scripts
	compiler Compiler

	// Executor for compiled script programs
	executor *Executor
}

func NewCompilingEvaluator(
	variableResolver VariableResolver,
	commandResolver CommandResolver,
	selectorResolver SelectorResolver,
	context any,
) Evaluator {
	return &CompilingEvaluator{
		compiler: Compiler{},
		executor: &Executor{
			variableResolver,
			commandResolver,
			selectorResolver,
			context,
		},
	}
}

// Evaluate a script and return result
//
// This will compile then execute the script
func (evaluator *CompilingEvaluator) EvaluateScript(script Script) Result {
	program := evaluator.compiler.CompileScript(script)
	return evaluator.executor.Execute(program, nil)
}

// Evaluate a sentence and return result
//
// This will execute a single-sentence script
func (evaluator *CompilingEvaluator) EvaluateSentence(sentence Sentence) Result {
	script := Script{}
	script.Sentences = append(script.Sentences, sentence)
	program := evaluator.compiler.CompileScript(script)
	return evaluator.executor.Execute(program, nil)
}

// Evaluate a word and return result
//
// This will execute a single-word program
func (evaluator *CompilingEvaluator) EvaluateWord(word Word) (result Result) {
	defer func() {
		if err := recover(); err != nil {
			if e, ok := err.(SyntaxError); ok {
				result = ERROR(e.message)
			} else {
				panic(err)
			}
		}
	}()
	program := evaluator.compiler.CompileWord(word)
	return evaluator.executor.Execute(program, nil)
}
