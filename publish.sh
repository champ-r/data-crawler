#!/usr/bin/env bash

args="${@}"

npm=$(command -v npm)

./data-crawler $args
cp output/index.json output/op.gg/
cp output/index.json output/op.gg-aram/
cp output/index.json output/murderbridge/
cp output/index.json output/lolalytics/

cd output/op.gg
$npm publish --access public

cd ../op.gg-aram
$npm publish --access public

cd ../murderbridge
$npm publish --access public

cd ../lolalytics
$npm publish --access public

cd ../lolalytics-aram
$npm publish --access public
