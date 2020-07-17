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

    info "goc gocc server started"
}

teardown_file() {
    kill -9 $GOC_PID
}

@test "test basic goc run" {
    info $PWD
    export GOPATH=$PWD/samples/simple_gopath_project
    export GO111MODULE=off
    cd samples/simple_gopath_project/src/qiniu.com/simple_gopath_project

    wait_profile_backend "run1" &
    profile_pid=$!

    run gocc run . --debug --debugcisyncfile ci-sync.bak;
    info run output: $output
    [ "$status" -eq 0 ]
    [[ "$output" == *"hello, world."* ]]

    wait $profile_pid
}