#!/usr/bin/env bash

args="${@}"

npm=$(command -v npm)

./data-crawler $args
cp output/index.json output/op.gg/
cp output/index.json output/murderbridge/

cd output/op.gg
$npm publish --access public

cd ../op.gg-aram
$npm publish --access public

cd ../murderbridge
$npm publish --access public
