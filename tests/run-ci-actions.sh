#!/bin/bash
set -ex

chmod +x /home/runner/tools/goc/goc
export PATH=/home/runner/tools/goc:$PATH

chmod +x /home/runner/tools/e2e.test/e2e.test
export PATH=/home/runner/tools/e2e.test:$PATH

cd e2e
e2e.test ./...