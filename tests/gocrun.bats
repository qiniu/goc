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

setup_file() {
    # run centered server
    goc server 3>&- &
    GOC_PID=$!
    sleep 2
    goc init

    # run covered goc run
    WORKDIR=$PWD
    cd samples/run_for_several_seconds
    ls -al
    gocc run --debug . 3>&- &
    GOCC_PID=$!
    sleep 2
    echo "goc gocc server started"
}

teardown_file() {
    cd $WORKDIR
    # collect from center
    goc profile --debug -o filtered2.cov
    kill -9 $GOC_PID
    kill -9 $GOCC_PID
}

@test "test basic goc run" {

}