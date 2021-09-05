package store

type Store interface {
	// 返回 key 对应的数据
	Get(key string) (string, error)
	// 设置 key 对应的数据
	Set(key string, value string) error
	// 删除 key 和 key 对应的数据
	Remove(key string) error
	// 遍历返回 pattern 开头的所有 key 数据
	Range(pattern string) ([]string, error)
	// 遍历删除 pattern 开头的所有 key 数据
	RangeRemove(pattern string) error
}
