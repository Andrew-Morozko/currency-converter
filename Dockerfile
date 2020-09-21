# Code and grammar generation
FROM python:3.8-alpine AS codegen
ARG API_KEY
WORKDIR /codegen
COPY ./codegen/requirements.txt /codegen/requirements.txt
RUN pip install -r /codegen/requirements.txt

COPY ./codegen /codegen
RUN python3 codegen.py ${API_KEY}


FROM openjdk:16-jdk-alpine AS grammar
RUN apk add wget &&\
    wget -O /antlr-4.8-complete.jar http://www.antlr.org/download/antlr-4.8-complete.jar

COPY --from=codegen /codegen/CurrencyConverterSymbols.g4 /grammar/
COPY ./grammar/CurrencyConverter.g4 /grammar/

RUN export CLASSPATH=".:/antlr-4.8-complete.jar:$CLASSPATH" &&\
    java -jar /antlr-4.8-complete.jar -Dlanguage=Go /grammar/CurrencyConverter.g4 -o /parser


FROM golang:1.15.0-buster
RUN apt-get update &&\
    apt-get install -y zip &&\
    rm -rf /var/lib/apt/lists/*

WORKDIR /data/
COPY go.* ./
RUN go mod download

COPY --from=codegen /codegen/cur_symbols.go ./cmd/
COPY --from=grammar /parser ./parser/

COPY ./cmd ./cmd/

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
