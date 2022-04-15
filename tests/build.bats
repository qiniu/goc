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

setup() {
    goc init
}

@test "test basic goc build command" {
    cd samples/run_for_several_seconds

    wait_profile_backend "build1" &
    profile_pid=$!

    run gocc build --debug --debugcisyncfile ci-sync.bak;
    info build1 output: $output
    [ "$status" -eq 0 ]

    wait $profile_pid
}

@test "test goc build command without debug" {
    cd samples/run_for_several_seconds

    wait_profile_backend "build2" &
    profile_pid=$!

    run gocc build --debugcisyncfile ci-sync.bak;
    info build2 output: $output
    [ "$status" -eq 0 ]

    wait $profile_pid
}

@test "test goc build in GOPATH project" {
    info $PWD
    export GOPATH=$PWD/samples/simple_gopath_project
    export GO111MODULE=off
    cd samples/simple_gopath_project/src/qiniu.com/simple_gopath_project

    wait_profile_backend "build3" &
    profile_pid=$!

    run gocc build --buildflags="-v" --debug --debugcisyncfile ci-sync.bak;
    info build3 output: $output
    [ "$status" -eq 0 ]

    wait $profile_pid
}

@test "test goc build with go.mod project which contains replace directive" {
    cd samples/gomod_replace_project

    wait_profile_backend "build4" &
    profile_pid=$!

    run gocc build --debug --debugcisyncfile ci-sync.bak;
    info build4 output: $output
    [ "$status" -eq 0 ]

    wait $profile_pid
}

@test "test goc build on complex project" {
    cd samples/complex_project

    wait_profile_backend "build5" &
    profile_pid=$!

    run gocc build --debug --debugcisyncfile ci-sync.bak;
    info build5 output: $output
    [ "$status" -eq 0 ]

    wait $profile_pid
}

@test "test goc build on reference other package project" {
    cd samples/reference_other_package_project/app

    wait_profile_backend "build6" &
    profile_pid=$!

    run gocc build --debug --debugcisyncfile ci-sync.bak;
    info build5 output: $output
    [ "$status" -eq 0 ]

    wait $profile_pid
}

@test "test basic goc build command with singleton" {
    cd samples/run_for_several_seconds

    wait_profile_backend "build7" &
    profile_pid=$!

    run gocc build --debug --singleton --debugcisyncfile ci-sync.bak;
    info build7 output: $output
    [ "$status" -eq 0 ]

    wait $profile_pid
}

@test "test goc build command on project using Go generics" {
    if ! go_version_at_least 1 18; then
        info skipped on old Go versions
        return 0
    fi

    cd samples/simple_project_with_generics

    wait_profile_backend "build8" &
    profile_pid=$!

    run gocc build --debug --debugcisyncfile ci-sync.bak;
    info build8 output: $output
    [ "$status" -eq 0 ]

    wait $profile_pid
}
