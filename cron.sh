#!/usr/bin/env bash

go build .
./data-crawler
cp output/index.json output/op.gg/
cp output/index.json output/murderbridge/

cd output/op.gg
npm publish --access public

cd ../murderbridge
npm publish --access public

