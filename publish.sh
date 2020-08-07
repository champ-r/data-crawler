#!/usr/bin/env bash

set -eo pipefail

go build .
./data-crawler
cp output/index.json output/op.gg/
cd output/op.gg
npm publish --access public
