#!/usr/bin/env bash
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

info() {
  echo -e "[$(date +'%Y-%m-%dT%H:%M:%S.%N%z')] INFO: $@" >&3
}

wait_profile() {
    local n=0
    local timeout=10
    until [[ ${n} -ge ${timeout} ]]
    do
        LS=`ls`
        info $1, $LS
        if [[ -f ci-sync.bak ]]; then
            break
        fi
        n=$[${n}+1]
        sleep 1
    done
    # collect from center
    goc profile -o filtered-$1.cov
    info "done $1 collect"
}

wait_profile_backend() {
    rm ci-sync.bak || true
    wait_profile $1
}