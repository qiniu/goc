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
    if [ -e samples/simple_agent/register.port ]; then
        rm samples/simple_agent/register.port
    fi
    if [ -e samples/simple_agent/simple-agent_profile_listen_addr ]; then
        rm samples/simple_agent/simple-agent_profile_listen_addr
    fi
    # run centered server
    goc server > samples/simple_agent/register.port &
    GOC_PID=$!
    sleep 2
    goc init

    WORKDIR=$PWD
    info "goc server started"
}

teardown_file() {
    kill -9 $GOC_PID
}

setup() {
    goc init
}

@test "test cover service listen port" {

    cd samples/simple_agent

    # test1: check cover with agent port
    goc build  --agentport=:7888
    sleep 2

    ./simple-agent 3>&- &
    SAMPLE_PID=$!
    sleep 2

    [ -e './simple-agent_profile_listen_addr' ]
    host=$(cat ./simple-agent_profile_listen_addr)

    check_port=$(cat register.port | grep $host)
    [ "$check_port" != "" ]

    kill -9 $SAMPLE_PID

    # test2: check cover with random port
    goc build
    sleep 2

    ./simple-agent 3>&- &
    SAMPLE_PID=$!
    sleep 2

    [ -e './simple-agent_profile_listen_addr' ]
    host=$(cat ./simple-agent_profile_listen_addr)

    check_port=$(cat register.port | grep $host)
    [ "$check_port" != "" ]

    kill -9 $SAMPLE_PID

    # test3: check cover with agent-port again
    goc build  --agentport=:7888
    sleep 2

    echo "" > register.port
    ./simple-agent 3>&- &
    SAMPLE_PID=$!
    sleep 2

    check_port=$(cat register.port | grep 7888)
    [ "$check_port" != "" ]

    kill -9 $SAMPLE_PID

    # test4: check cover with random port again
    goc build
    sleep 2

    ./simple-agent 3>&- &
    SAMPLE_PID=$!
    sleep 2

    [ -e './simple-agent_profile_listen_addr' ]
    host=$(cat ./simple-agent_profile_listen_addr)

    check_port=$(cat register.port | grep $host)
    [ "$check_port" != "" ]

    kill -9 $SAMPLE_PID
}