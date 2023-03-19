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

    info "goc server started"
}

teardown_file() {
    rm *_profile_listen_addr
    kill -9 $GOC_PID
}

setup() {
    goc init
}

@test "test goc merge with same binary" {
    cd samples/merge_profile_samples

    wait_profile_backend "merge1" &
    profile_pid=$!

    # merge two profiles with same binary
    run gocc merge a.voc b.voc --output mergeprofile.voc1 --debug --debugcisyncfile ci-sync.bak;
    info merge1 output: $output
    [ "$status" -eq 0 ]
    run cat mergeprofile.voc1
    [[ "$output" == *"qiniu.com/kodo/apiserver/server/main.go:32.49,33.13 1 60"* ]]
    [[ "$output" == *"qiniu.com/kodo/apiserver/server/main.go:42.49,43.13 1 2"* ]]
}

@test "test goc merge with two binaries, but has some source code in common" {
    cd samples/merge_profile_samples

    wait_profile_backend "merge2" &
    profile_pid=$!

    # merge two profiles from two binaries, but has some source code in common
    run gocc merge a.voc c.voc --output mergeprofile.voc2 --debug --debugcisyncfile ci-sync.bak;
    info merge2 output: $output
    [ "$status" -eq 0 ]
    run cat mergeprofile.voc2
    [[ "$output" == *"qiniu.com/kodo/apiserver/server/main.go:32.49,33.13 1 60"* ]]
    [[ "$output" == *"qiniu.com/kodo/apiserver/server/main.go:42.49,43.13 1 0"* ]]
    [[ "$output" == *"qiniu.com/kodo/apiserver/server/wala.go:42.49,43.13 1 0"* ]]

    wait $profile_pid
}
