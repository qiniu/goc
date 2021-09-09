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
