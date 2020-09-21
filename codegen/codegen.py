"""
Generate currency symbols from known currencyconverterapi.com currencies
"""

from collections import Counter
import re
import os
import requests
import json
import sys

assert len(sys.argv) >= 2, "enter api key"

API_KEY = sys.argv[1]

t = requests.get(f"https://free.currconv.com/api/v7/currencies?apiKey={API_KEY}")

cur = t.json()["results"]
cur["RUB"]["currencySymbol"] = "₽"
cur["BTC"]["currencySymbol"] = "₿"

cur_sym = dict(
    map(
        lambda x: (x["currencySymbol"].upper(), x["id"]),
        filter(lambda x: "currencySymbol" in x, cur.values()),
    )
)


for sym, count in Counter(cur_sym.keys()).most_common():
    if count == 1:
        break
    # Removing symbols with multiple meanings
    del cur_sym[sym]

# Overrides for most common symbols
cur_sym["$"] = "USD"
cur_sym["£"] = "GBP"
cur_sym["₩"] = "KRW"
cur_sym["P."] = "BYN"

code_to_sym = dict((v, k) for k, v in cur_sym.items())

cur_sym_re = "|".join(map(lambda x: f"'{x}'", cur_sym.keys()))

with open("CurrencyConverterSymbols.g4", "w") as f:
    f.write(
        f"""lexer grammar CurrencyConverterSymbols;
CURSIGN: {cur_sym_re};
"""
    )

with open("cur_symbols.go", "w") as f:
    sym_code = "\n".join(
        f"\t`{k}`: `{v}`," for k, v in sorted(cur_sym.items(), key=lambda x: x[1])
    )
    code_sym = "\n".join(
        f"\t`{cur_code}`: {{`{cur[cur_code]['currencyName']}`, `{code_to_sym.get(cur_code, '')}`}},"
        for cur_code in sorted(cur.keys())
    )

    codes = ", ".join(f"`{code}`" for code in sorted(cur.keys()))
    f.write(
        f"""package main

var SymToCode = map[string]string {{
{sym_code}
}}

var CodeToSym = map[string]struct {{
	Name string
	Sym  string
}}{{
{code_sym}
}}

var Codes = []string{{{codes}}}
"""
    )


# os.system("go fmt cur_symbols.go")
