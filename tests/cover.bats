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
    mkdir -p test-temp
    cp samples/simple_project/main.go test-temp
    cp samples/simple_project/go.mod test-temp

    # run centered server
    goc server 3>&- &
    GOC_PID=$!
    sleep 2
    goc init

    info "goc server started"
}

teardown_file() {
    cp test-temp/filtered* .
    rm -rf test-temp
    kill -9 $GOC_PID
}

@test "test basic goc cover command" {
    cd test-temp
    wait_profile_backend "cover1"
    run gocc cover --debug --debugcisyncfile ci-sync.bak;
    info cover1 output: $output
    [ "$status" -eq 0 ]

    run ls http_cover_apis_auto_generated.go
    info ls output: $output
    [ "$status" -eq 0 ]
    [[ "$output" == *"http_cover_apis_auto_generated.go"* ]]
}
