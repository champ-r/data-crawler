#!/usr/bin/env bash

args="${@}"

go build .
./data-crawler -debug $args
