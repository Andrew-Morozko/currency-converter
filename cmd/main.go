package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/olekukonko/tablewriter"
)

type Args struct {
	Expression         string `arg:"--expr,positional" help:"expression to evaluate"`
	ApiKey             string `arg:"--api-key, env:API_KEY" help:"currencyconverterapi.com API key"`
	DefaultSrcCurrency string `arg:"--src, env:DEFAULT_SRC_CURRENCY"  help:"default source currency"`
	DefaultTgtCurrency string `arg:"--tgt, env:DEFAULT_TGT_CURRENCY"  help:"default target currency"`
	DecSep             string `arg:"--dec-sep, env:DEC_SEP"  help:"decimal separator character[s]" default:"."`
	Format             string `arg:"-f,--format, env:OUTPUT_FORMAT" help:"output format (alfred/text/num)" default:"text" placeholder:"FMT"`
	ClearCache         bool   `arg:"--clear-cache" help:"reset all cached data"`
	NoCache            bool   `arg:"--no-cache" help:"request rates every time"`
	ListCurrencies     bool   `arg:"--list" help:"show the list of supported currencies"`
	Version            bool   `arg:"-v,--version" help:"print version"`
}

func (Args) Description() string {
	return `Currency converter with built-in calculator. Version ` + version + `.

Mathematical operations:
+ - * /
+X% -X%
a^b (power)

Numbers could include spaces to aid with readability and use "," or "." for decimal point.

Examples:
"1" - convert 1 unit of default source currency to default target currency.
"1 USD" | "$1" - convert 1 USD to default target currency.
"1 in EUR" | "1 to €" - convert 1 unit of default source currency to euro.
"1 usd eur" - convert 1 USD to EUR (you can omit "to"/"in" if both currency symbols are present)

You can use standard 3-letter currency code or currency symbols.`

}

func SpaceFormatFloat(fl float64) (res string) {
	txt := fmt.Sprintf("%.2f", fl)
	data := strings.Split(txt, ".")
	s := 0
	e := (len(data[0]) % 3)
	if e == 0 {
		e = 3
	}
	for s < len(data[0]) {
		res += data[0][s:e] + " "
		s = e
		e += 3
	}
	res = res[0 : len(res)-1]
	return fmt.Sprintf("%s.%s", res, data[1])
}

var args = &Args{}

var CurConvCacheFile = filepath.Join(os.TempDir(), "cur_conf_cache.json")

var _DEBUG_STR = "true"
var version = "in-dev" // will be replaced by build script

var DEBUG = func() bool {
	val, err := strconv.ParseBool(_DEBUG_STR)
	if err != nil {
		panic(err)
	}
	return val
}()

func reportError(msg string, isUI bool) {
	switch strings.ToLower(strings.TrimSpace(args.Format)) {
	case "alfred":
		if !(isUI || DEBUG) {
			msg = "please enter a valid expression"
		}
		fmt.Printf(`
{"items": [
 	{
 		"title": "...",
 		"subtitle": "%s",
 		"valid": false,
 	}
]}
`,
			strings.ReplaceAll(msg, `"`, `\"`),
		)
		os.Exit(0)
	case "text", "num":
		if isUI || DEBUG {
			fmt.Fprintf(os.Stderr, "%s\n", msg)
		} else {
			fmt.Fprintf(os.Stderr, "Unknown error\n")
		}
		os.Exit(1)
	}
}
func formatCurrency(code string) string {
	if t := CodeToSym[code].Sym; t != "" {
		return t
	} else {
		return " " + code
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case UIError:
				reportError(v.Error(), true)
			case string:
				reportError(v, false)
			default:
				reportError(fmt.Sprintf("Unknown error: %+v", v), false)
			}
		}
	}()

	p := arg.MustParse(args)

	if args.Version {
		fmt.Println(version)
		os.Exit(0)
	}

	if args.ListCurrencies {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Code", "Symbol", "Name"})
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		// table.SetTablePadding("\t") // pad with tabs
		table.SetNoWhiteSpace(true)

		for _, code := range Codes {
			table.Append([]string{code, CodeToSym[code].Sym, CodeToSym[code].Name})
		}

		table.Render()
		os.Exit(0)
	}
	if args.ClearCache {
		_ = os.Remove(CurConvCacheFile)
		os.Exit(0)
	}
	args.ApiKey = strings.TrimSpace(args.ApiKey)
	if args.ApiKey == "" {
		p.Fail("No API key present")
	}
	args.Expression = strings.TrimSpace(args.Expression)
	if args.Expression == "" {
		p.Fail("No expression present")
	}
	args.Format = strings.ToLower(strings.TrimSpace(args.Format))
	switch args.Format {
	case "text", "num", "alfred":
	default:
		p.Fail(`Format could be "text", "num" or "alfred"`)
	}

	args.DecSep = strings.ToLower(strings.TrimSpace(args.DecSep))
	if args.DecSep == "" {
		p.Fail("Empty decimal separators list")
	}

	var cache ExchangeRateCache
	if args.NoCache {
		cache = EmptyExchangRatesCache()
	} else {
		cache = LoadExchangRatesCache()
	}

	res := RunCalc()
	cache.Convert(res)

	switch args.Format {
	case "text":
		fmt.Printf(
			"%s%s is %s%s\n",
			SpaceFormatFloat(res.Src), formatCurrency(res.SrcCurrency),
			SpaceFormatFloat(res.Tgt), formatCurrency(res.TgtCurrency),
		)
	case "num":
		fmt.Printf(
			"%.2f\n",
			res.Tgt,
		)
	case "alfred":
		text := fmt.Sprintf(
			"%s%s is %s%s",
			SpaceFormatFloat(res.Src), formatCurrency(res.SrcCurrency),
			SpaceFormatFloat(res.Tgt), formatCurrency(res.TgtCurrency),
		)
		fmt.Printf(`{"items": [
	{
		"title": "%s",
		"subtitle": "Action this item to copy this number to the clipboard",
		"arg": "%.2f",
		"valid": true,
		"text": {
			"copy": "%.2f",
			"largetype": "%s"
		},
	}
]}
`,
			text, res.Tgt, res.Tgt, text,
		)
	}
}
