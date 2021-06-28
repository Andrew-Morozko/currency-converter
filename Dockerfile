# Code and grammar generation
FROM python:3.9-alpine AS codegen
WORKDIR /codegen
COPY ./codegen/requirements.txt /codegen/requirements.txt
RUN pip install -r /codegen/requirements.txt

COPY ./codegen /codegen
ARG VER
ARG API_KEY
RUN echo "Running codegen for version" ${VER}; python3 codegen.py ${API_KEY}

FROM openjdk:17-jdk-alpine AS grammar
ARG ANTLR_VER=4.9.2
RUN apk add wget &&\
    wget -O /antlr-${ANTLR_VER}-complete.jar http://www.antlr.org/download/antlr-${ANTLR_VER}-complete.jar

COPY --from=codegen /codegen/CurrencyConverterSymbols.g4 /grammar/
COPY ./grammar/CurrencyConverter.g4 /grammar/

RUN export CLASSPATH=".:/antlr-${ANTLR_VER}-complete.jar:$CLASSPATH" &&\
    java -jar /antlr-${ANTLR_VER}-complete.jar -Dlanguage=Go /grammar/CurrencyConverter.g4 -o /parser


FROM golang:1.16.5-buster
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
