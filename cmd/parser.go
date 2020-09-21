package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Andrew-Morozko/currency-converter/parser"
	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type ErrorListener struct {
	antlr.ErrorListener
}

func (c *ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	panic(fmt.Sprintf("character %d: %s", column, msg))
}

var errorListenerInst = &ErrorListener{
	ErrorListener: antlr.NewDefaultErrorListener(),
}

type calcListener struct {
	*parser.BaseCurrencyConverterListener
	SrcCurrencyCode string
	TgtCurrencyCode string
	curCurrencyTgt  *string

	numReplacer *strings.Replacer

	Result float64
	stack  []float64
}

func (l *calcListener) push(i float64) {
	l.stack = append(l.stack, i)
}

func (l *calcListener) pop() float64 {
	if len(l.stack) == 0 {
		panic("stack is empty unable to pop")
	}

	// Get the last value from the stack.
	result := l.stack[len(l.stack)-1]

	// Remove the last element from the stack.
	l.stack = l.stack[:len(l.stack)-1]

	return result
}

func (l *calcListener) EnterExplicit_src(c *parser.Explicit_srcContext) {
	l.curCurrencyTgt = &l.SrcCurrencyCode
}

func (l *calcListener) EnterDst(c *parser.DstContext) {
	l.curCurrencyTgt = &l.TgtCurrencyCode
}

func (l *calcListener) ExitCurrency(c *parser.CurrencyContext) {
	var ok bool
	sym := c.GetSym()
	if sym != nil {
		*l.curCurrencyTgt, ok = SymToCode[sym.GetText()]
		if !ok {
			panic(UIError{fmt.Sprintf(`unknown currency symbol "%s"`, sym.GetText())})
		}
	} else {
		curCode := c.GetName().GetText()
		_, ok = CodeToSym[curCode]
		if !ok {
			panic(UIError{fmt.Sprintf(`unknown currency code "%s"`, curCode)})
		}
		*l.curCurrencyTgt = curCode
	}
}

func (l *calcListener) ExitExpr(c *parser.ExprContext) {
	op := c.GetOp()
	if op == nil {
		return
	}
	operation := op.GetTokenType()
	var left_val, right_val float64

	right_val = l.pop()
	if c.GetRight_pct() != nil {
		// Convert % to mul
		right_val /= 100
		switch op.GetTokenType() {
		case parser.CurrencyConverterParserADD:
			right_val = 1 + right_val
		case parser.CurrencyConverterLexerSUB:
			right_val = 1 - right_val
		default:
			panic(UIError{"invalid operation with %"})
		}
		operation = parser.CurrencyConverterParserMUL
	}

	if c.GetLeft() != nil {
		left_val = l.pop()
	}

	switch operation {
	case parser.CurrencyConverterParserPOW:
		l.push(math.Pow(left_val, right_val))
	case parser.CurrencyConverterParserMUL:
		l.push(left_val * right_val)
	case parser.CurrencyConverterParserDIV:
		l.push(left_val / right_val)
	case parser.CurrencyConverterParserADD:
		l.push(left_val + right_val)
	case parser.CurrencyConverterParserSUB:
		l.push(left_val - right_val)
	default:
		panic(UIError{"unknown operation"})
	}

}

func (l *calcListener) ExitNum(c *parser.NumContext) {
	text := c.GetText()
	numstr := l.numReplacer.Replace(text)
	num, err := strconv.ParseFloat(numstr, 64)
	if err != nil {
		if strings.Count(numstr, ".") > 1 {
			panic(UIError{fmt.Sprintf(`multiple decimal points in "%s"`, text)})
		} else {
			panic(UIError{fmt.Sprintf(`can't convert "%s" to number`, text)})
		}
	}
	l.push(num)
}

func (l *calcListener) ExitRoot(c *parser.RootContext) {
	if len(l.stack) != 1 {
		panic("Incorrect expression!")
	}
	l.Result = l.pop()
}

type ParseResult struct {
	Src         float64
	SrcCurrency string
	Tgt         float64
	TgtCurrency string
}

func createReplacer() *strings.Replacer {
	var replacerArgs []string
	noopChars := []rune("., '`")
	for _, decSepChar := range args.DecSep {
		for i, noopChar := range noopChars {
			if decSepChar == noopChar {
				noopChars[i] = noopChars[len(noopChars)-1]
				noopChars = noopChars[:len(noopChars)-1]
				break
			}
		}
		// Replacing every decSepChar with "."
		if decSepChar != '.' {
			replacerArgs = append(replacerArgs, string(decSepChar), ".")
		}
	}
	// Replacing every non-decSepChar with ""
	for _, noopChar := range noopChars {
		replacerArgs = append(replacerArgs, string(noopChar), "")
	}
	return strings.NewReplacer(replacerArgs...)
}

func RunCalc() (res *ParseResult) {
	is := antlr.NewInputStream(strings.ToUpper(args.Expression))
	lexer := parser.NewCurrencyConverterLexer(is)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(errorListenerInst)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewCurrencyConverterParser(stream)
	p.RemoveErrorListeners()
	p.AddErrorListener(errorListenerInst)

	listener := &calcListener{
		SrcCurrencyCode: args.DefaultSrcCurrency,
		TgtCurrencyCode: args.DefaultTgtCurrency,
		numReplacer:     createReplacer(),
	}

	antlr.ParseTreeWalkerDefault.Walk(listener, p.Root())

	if listener.SrcCurrencyCode == "" {
		panic(UIError{"Can't determine source currency for conversion"})
	}
	if listener.TgtCurrencyCode == "" {
		panic(UIError{"Can't determine target currency for conversion"})
	}

	return &ParseResult{
		Src:         listener.Result,
		SrcCurrency: listener.SrcCurrencyCode,
		TgtCurrency: listener.TgtCurrencyCode,
	}
}
