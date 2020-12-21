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
    cd samples/run_for_several_seconds
    goc build --center=http://127.0.0.1:60001

    info "goc server started"
}

teardown_file() {
    kill -9 $GOC_PID
    kill -9 $GOCC_PID
}

setup() {
    goc init --center=http://127.0.0.1:60001
    goc init
}

@test "test goc profile to stdout" {
    ./simple-project 3>&- &
    SAMPLE_PID=$!
    sleep 2

    wait_profile_backend "profile1" &
    profile_pid=$!

    run gocc profile --center=http://127.0.0.1:60001 --debug --debugcisyncfile ci-sync.bak
    info $output
    [ "$status" -eq 0 ]
    [[ "$output" == *"mode: count"* ]]

    wait $profile_pid
    kill -9 $SAMPLE_PID
}

@test "test goc profile to file" {
    ./simple-project 3>&- &
    SAMPLE_PID=$!
    sleep 2

    wait_profile_backend "profile2" &
    profile_pid=$!

    run gocc profile --center=http://127.0.0.1:60001 -o test-profile.bak --debug --debugcisyncfile ci-sync.bak;
    [ "$status" -eq 0 ]
    run cat test-profile.bak
    [[ "$output" == *"mode: count"* ]]

    wait $profile_pid
    kill -9 $SAMPLE_PID
}

@test "test goc profile with coverfile flag" {
    ./simple-project 3>&- &
    SAMPLE_PID=$!
    sleep 2

    wait_profile_backend "profile3" &
    profile_pid=$!

    run gocc profile --center=http://127.0.0.1:60001 --coverfile="a.go$,b.go$" --debug --debugcisyncfile ci-sync.bak;
    info $output
    [ "$status" -eq 0 ]
    [[ "$output" == *"mode: count"* ]]
    [[ "$output" == *"a.go"* ]] # contains a.go file
    [[ "$output" == *"b.go"* ]] # contains b.go file
    [[ "$output" != *"main.go"* ]] # not contains main.go file

    wait $profile_pid
    kill -9 $SAMPLE_PID
}

@test "test goc profile with service flag" {
    ./simple-project 3>&- &
    SAMPLE_PID=$!
    sleep 2

    wait_profile_backend "profile4" &
    profile_pid=$!

    run gocc profile --center=http://127.0.0.1:60001 --service="simple-project" --debug --debugcisyncfile ci-sync.bak;
    info $output
    [ "$status" -eq 0 ]
    [[ "$output" == *"mode: count"* ]]

    wait $profile_pid
    kill -9 $SAMPLE_PID
}

@test "test goc profile with force flag" {
    ./simple-project 3>&- &
    SAMPLE_PID=$!
    sleep 2

    wait_profile_backend "profile5" &
    profile_pid=$!

    run gocc profile --center=http://127.0.0.1:60001 --service="simple-project,unknown" --force --debug --debugcisyncfile ci-sync.bak;
    info $output
    [ "$status" -eq 0 ]
    [[ "$output" == *"mode: count"* ]]

    wait $profile_pid
    kill -9 $SAMPLE_PID
}

@test "test goc profile with coverfile and skipfile flags" {
    ./simple-project 3>&- &
    SAMPLE_PID=$!
    sleep 2

    wait_profile_backend "profile6" &
    profile_pid=$!

    run gocc profile --center=http://127.0.0.1:60001 --coverfile="a.go$,b.go$" --skipfile="b.go$" --debug --debugcisyncfile ci-sync.bak;
    info $output
    [ "$status" -eq 0 ]
    [[ "$output" == *"mode: count"* ]]
    [[ "$output" == *"a.go"* ]] # contains a.go file
    [[ "$output" != *"b.go"* ]] # not contains b.go file
    [[ "$output" != *"main.go"* ]] # not contains main.go file

    wait $profile_pid
    kill -9 $SAMPLE_PID
}