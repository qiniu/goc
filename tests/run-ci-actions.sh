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

set -ex

echo "test start"

bats -t server.bats

bats -t run.bats

bats -t version.bats

bats -t list.bats

bats -t clear.bats

bats -t build.bats

bats -t profile.bats

bats -t install.bats

bats -t register.bats

bats -t init.bats

bats -t diff.bats

bats -t cover.bats

bash <(curl -s https://codecov.io/bash) -f 'filtered*' -F e2e-$GOVERSION