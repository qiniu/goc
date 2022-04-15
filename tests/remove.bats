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

    WORKDIR=$PWD
    cd $WORKDIR/samples/$(demo_service_name)
    gocc build --center=http://127.0.0.1:60001
    ./simple-project 3>&- &
    SAMPLE_PID=$!
    sleep 2

    info "goc server started"
}

teardown_file() {
    rm *_profile_listen_addr
    kill -9 $GOC_PID
    kill -9 $GOCC_PID
    kill -9 $SAMPLE_PID
}

@test "test basic goc remove command" {
    wait_profile_backend "remove1" &
    profile_pid=$!

    run goc list --center=http://127.0.0.1:60001;
    info remove1_1 output: $output
    [ "$status" -eq 0 ]
    [[ "$output" =~ "simple-project" ]]

    run gocc remove --center=http://127.0.0.1:60001 --service="simple-project" --debug --debugcisyncfile ci-sync.bak;
    info remove1_2 output: $output
    [ "$status" -eq 0 ]
    [[ "$output" =~ "removed from the center" ]]

    run goc list --center=http://127.0.0.1:60001;
    info remove1_3 output: $output
    [ "$status" -eq 0 ]
    [ "$output" = "{}" ]

    wait $profile_pid
}
