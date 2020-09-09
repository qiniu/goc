#!/usr/bin/env bats
# Copyright 2020 Qiniu Cloud (七牛云)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

load util.sh

setup_file() {
    # run centered server
    goc server 3>&- &
    GOC_PID=$!
    sleep 2
    goc init
    # run covered goc
    gocc server --port=:60001 --debug 3>&- &
    GOCC_PID=$!
    sleep 2
    info "goc gocc server started"
}

teardown_file() {
    # collect from center
    goc profile --debug -o filtered-server.cov
    kill -9 $GOC_PID
    kill -9 $GOCC_PID
}

@test "test basic goc server" {
    # connect to covered goc
    run goc clear --center=http://127.0.0.1:60001
    [ "$status" -eq 0 ]
    # connect to covered goc
    run goc profile --center=http://127.0.0.1:60001
    # no profiles
    [ "$status" -eq 1 ]
}

@test "register a covered service" {
    WORKDIR=$PWD
    cd $WORKDIR/samples/run_for_several_seconds

    run goc build --debug --center=http://127.0.0.1:60001
    [ "$status" -eq 0 ]
    ./simple-project 3>&- &
    sleep 1
    # connect to covered goc
    run goc profile --center=http://127.0.0.1:60001
    [ "$status" -eq 0 ]
}