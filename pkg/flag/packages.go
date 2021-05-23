package flag

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/qiniu/goc/v2/pkg/config"
)

// GetPackagesDir parse [pacakges] part of args, it will fatal if error encountered
//
// 函数获取 1： [packages] 所在的目录位置，供后续插桩使用。
//
// 函数获取 2： 如果参数是 *.go，第一个 .go 文件的文件名。go build 中，二进制名字既可能是目录名也可能是文件名，和参数类型有关。
//
// 如果 [packages] 非法（即不符合 go 原生的定义），则返回对应错误
// 这里只考虑 go mod 的方式
func GetPackagesDir(patterns []string) {
	for _, p := range patterns {
		// patterns 只支持两种格式
		// 1. 要么是直接指向某些 .go 文件的相对/绝对路径
		if strings.HasSuffix(p, ".go") {
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
				// check if valid
				if err := goFilesPackage(patterns); err != nil {
					log.Fatalf("%v", err)
				}

				// 获取绝对路径
				absp, err := filepath.Abs(p)
				if err != nil {
					log.Fatalf("%v", err)
				}

				// 获取当前 [packages] 所在的目录位置，供后续插桩使用。
				config.GocConfig.CurPkgDir = filepath.Dir(absp)
				// 获取二进制名字
				config.GocConfig.BinaryName = filepath.Base(absp)
				return
			}
		}
	}

	// 2. 要么是 import path
	coverWd, err := getDirFromImportPaths(patterns)
	if err != nil {
		log.Fatalf("%v", err)
	}

	config.GocConfig.CurPkgDir = coverWd
	config.GocConfig.BinaryName = ""
	return
}

// goFilesPackage 对一组 go 文件解析，判断是否合法
// go 本身还判断语法上是否是同一个 package，goc 这里不做解析
// 1. 都是 *.go 文件？
// 2. *.go 文件都在同一个目录？
// 3. *.go 文件存在？
func goFilesPackage(gofiles []string) error {
	// 1. 必须都是 *.go 结尾
	for _, f := range gofiles {
		if !strings.HasSuffix(f, ".go") {
			return fmt.Errorf("named files must be .go files: %s", f)
		}
	}

	var dir string
	for _, file := range gofiles {
		// 3. 文件都存在？
		fi, err := os.Stat(file)
		if err != nil {
			return err
		}

		// 2.1 有可能以 *.go 结尾的目录
		if fi.IsDir() {
			return fmt.Errorf("%s is a directory, should be a Go file", file)
		}

		// 2.2 所有 *.go 必须在同一个目录内
		dir1, _ := filepath.Split(file)
		if dir1 == "" {
			dir1 = "./"
		}

		if dir == "" {
			dir = dir1
		} else if dir != dir1 {
			return fmt.Errorf("named files must all be in one directory: have %s and %s", dir, dir1)
		}
	}

	return nil
}

// getDirFromImportPaths return the import path's real directory
//
// 该函数接收到的只有 dir 或 import path，file 在上一步已被排除
// 只考虑 go modules 的情况
func getDirFromImportPaths(patterns []string) (string, error) {
	// no import path, pattern = current wd
	if len(patterns) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("fail to parse import path: %w", err)
		}
		return wd, nil
	}

	// 为了简化插桩的逻辑，goc 对 import path 要求必须都在同一个目录
	// 所以干脆只允许一个 pattern 得了 -_-
	// 对于 goc build/run 来说本身就是只能在一个目录内
	// 对于 goc install 来讲，这个行为就和 go install 不同，不过多 import path 较少见 >_<，先忽略
	if len(patterns) > 1 {
		return "", fmt.Errorf("goc only support one import path now")
	}

	pattern := patterns[0]
	switch {
	// case isLocalImport(pattern) || filepath.IsAbs(pattern):
	// 	dir1, err := filepath.Abs(pattern)
	// 	if err != nil {
	// 		return "", fmt.Errorf("error (%w) get directory from the import path: %v", err, pattern)
	// 	}
	// 	if _, err := os.Stat(dir1); err != nil {
	// 		return "", fmt.Errorf("error (%w) get directory from the import path: %v", err, pattern)
	// 	}
	// 	return dir1, nil

	case strings.Contains(pattern, "..."):
		i := strings.Index(pattern, "...")
		dir, _ := filepath.Split(pattern[:i])
		if _, err := os.Stat(dir); err != nil {
			return "", fmt.Errorf("error (%w) get directory from the import path: %v", err, pattern)
		}
		return dir, nil

	case strings.IndexByte(pattern, '@') > 0:
		return "", fmt.Errorf("import path with @ version query is not supported in goc")

	case isMetaPackage(pattern):
		return "", fmt.Errorf("`std`, `cmd`, `all` import path is not supported by goc")

	default: // 到这一步认为 pattern 是相对路径或者绝对路径
		dir1, err := filepath.Abs(pattern)
		if err != nil {
			return "", fmt.Errorf("error (%w) get directory from the import path: %v", err, pattern)
		}
		if _, err := os.Stat(dir1); err != nil {
			return "", fmt.Errorf("error (%w) get directory from the import path: %v", err, pattern)
		}

		return dir1, nil
	}
}

// isLocalImport reports whether the import path is
// a local import path, like ".", "..", "./foo", or "../foo"
func isLocalImport(path string) bool {
	return path == "." || path == ".." ||
		strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../")
}

// isMetaPackage checks if the name is a reserved package name
func isMetaPackage(name string) bool {
	return name == "std" || name == "cmd" || name == "all"
}

// find direct path of current project which contains go.mod
func findModuleRoot(dir string) string {
	dir = filepath.Clean(dir)

	// look for enclosing go.mod
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}

	return ""
}
