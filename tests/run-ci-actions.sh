#!/bin/bash
set -ex

chmod +x /home/runner/tools/goc/goc
export PATH=/home/runner/tools/goc:$PATH

chmod +x /home/runner/tools/e2e.test/e2e.test
export PATH=/home/runner/tools/e2e.test:$PATH

export TESTS_ROOT=$PWD

cd $TESTS_ROOT/e2e
e2e.test ./...