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
    sleep 1

    info "goc server started"
}

teardown_file() {
    kill -9 $GOC_PID
    kill -9 $GOCC_PID
}

# we need to catch gocc server, so no init
# setup() {
#     goc init
# }

@test "test basic goc register command" {
    wait_profile_backend "register1" &
    profile_pid=$!

    run gocc register --center=http://127.0.0.1:60001 --name=xyz --address=http://137.0.0.1:666 --debug --debugcisyncfile ci-sync.bak;
    info register1 output: $output
    [ "$status" -eq 0 ]
    [[ "$output" == *"success"* ]]

    wait $profile_pid
}

@test "test goc register without port" {
    wait_profile_backend "register2" &
    profile_pid=$!

    run gocc register --center=http://127.0.0.1:60001 --name=xyz --address=http://137.0.0.1 --debug --debugcisyncfile ci-sync.bak;
    info register2 output: $output
    [ "$status" -eq 0 ]
    [[ "$output" != *"missing port"* ]]

    wait $profile_pid
}
