#!/usr/bin/env bash

set -eo pipefail

go build .
./data-crawler -opgg -mb
cp output/index.json output/op.gg/
cp output/index.json output/murderbridge/

cd output/op.gg
npm publish --access public

cd ../murderbridge
npm publish --access public

