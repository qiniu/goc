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
    kill -9 $GOC_PID
}

@test "test basic goc diff command" {
    cd samples/diff_samples

    wait_profile_backend "diff1"


    run gocc diff --new-profile=./new.voc --base-profile=./base.voc --debug --debugcisyncfile ci-sync.bak;
    info list output: $output
    [ "$status" -eq 0 ]
    [[ "$output" == *"qiniu.com/kodo/apiserver/server/main.go |     50.0%     |    100.0%    | 50.0%"* ]]
    [[ "$output" == *"Total                                   |     50.0%     |    100.0%    | 50.0%"* ]]
}

@test "test diff in prow environment with periodic job" {
    cd samples/diff_samples

    wait_profile_backend "diff2"

    export JOB_TYPE=periodic

    run gocc diff --new-profile=./new.voc --prow-postsubmit-job=base --debug --debugcisyncfile ci-sync.bak;
    info diff1 output: $output
    [ "$status" -eq 0 ]
    [[ "$output" == *"do nothing"* ]]
}

@test "test diff in prow environment with postsubmit job" {
    cd samples/diff_samples

    wait_profile_backend "diff3"

    export JOB_TYPE=postsubmit

    run gocc diff --new-profile=./new.voc --prow-postsubmit-job=base --debug --debugcisyncfile ci-sync.bak;
    info diff2 output: $output
    [ "$status" -eq 0 ]
    [[ "$output" == *"do nothing"* ]]
}