# currency-converter
Extremely overengineered currency converter.

Rates are acquired from currencyconverterapi.com (you should get their free API key to use this tool). Rates are cached locally for the next hour.

## Currency converter

The following format is supported
* `1` - convert 1 unit of default source currency to default target currency.
* `1 USD` | `$1` - convert 1 USD to default target currency.
* `1 in EUR` | `1 to €` - convert 1 unit of default source currency to euro.
* `1 usd eur` - convert 1 USD to EUR (you can omit "to"/"in" if both currencies are present)

You can use standard 3-letter currency code or some currency symbols. See `currency-converter --list` for the full list.


## Math
This currency converter also supports mathematical notation (in order of precedence):
```
()
a^b (power)
* /
+ -
+X% -X% (percent addition/subtraction)
```
Numbers could use custom `--dec-sep` characters from the list ```., `'``` as decimal separators. All other characters from the list are ignored and could be used to aid with readability.

Example: `$ (1' 000' 000/2 000 + 25% - 5*2)^1.01 in eur`

## Output
The tool has 3 output formats:
* `text` - human readable: `655.79$ is 557.02€`
* `num` - just to get the result: `557.02`
* `alfred` - [alfred](https://www.alfredapp.com)-compatible json result



## CLI Usage

```
Usage: currency-converter [--api-key API-KEY] [--src SRC] [--tgt TGT]
 [--dec-sep DEC-SEP] [--format FMT] [--clear-cache]
 [--no-cache] [--list] [--version] EXPR

Positional arguments:
 EXPR expression to evaluate

Options:
 --api-key API-KEY currencyconverterapi.com API key
 --src SRC default source currency
 --tgt TGT default target currency
 --dec-sep DEC-SEP decimal separator character[s] [default: .]
 --format FMT, -f FMT output format (alfred/text/num) [default: text]
 --clear-cache reset all cached data
 --no-cache request rates every time
 --list show the list of supported currencies
 --version, -v print version
 --help, -h display this help and exit
```

In general all parameters could be supplied via environment variables.

## Credits

Icon made by [Pixel perfect](https://www.flaticon.com/authors/pixel-perfect) from [www.flaticon.com](https://www.flaticon.com/)

