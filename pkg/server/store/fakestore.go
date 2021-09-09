/*
 Copyright 2021 Qiniu Cloud (qiniu.com)
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
     http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package store

type FakeStore struct {
}

func NewFakeStore() *FakeStore {
	return &FakeStore{}
}

// 返回 key 对应的数据
func (f *FakeStore) Get(key string) (string, error) {
	return "", nil
}

// 设置 key 对应的数据
func (f *FakeStore) Set(key string, value string) error {
	return nil
}

// 删除 key 和 key 对应的数据
func (f *FakeStore) Remove(key string) error {
	return nil
}

// 遍历返回 pattern 开头的所有 key 数据
func (f *FakeStore) Range(pattern string) ([]string, error) {
	return nil, nil
}

// 遍历删除 pattern 开头的所有 key 数据
func (f *FakeStore) RangeRemove(pattern string) error {
	return nil
}
