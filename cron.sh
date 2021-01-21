#!/usr/bin/env bash

args="${@}"

go=$(which go)
npm=$(which npm)

$go build .
./data-crawler $args
cp output/index.json output/op.gg/
cp output/index.json output/murderbridge/

cd output/op.gg
$npm publish --access public

cd ../murderbridge
$npm publish --access public
