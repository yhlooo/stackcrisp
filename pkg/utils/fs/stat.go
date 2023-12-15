package fs

import "os"

// IsExists 返回路径是否存在
func IsExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir 返回路径是否目录
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		// 获取文件信息错误
		return false
	}
	return info.IsDir()
}

// IsEmptyDir 返回路径是否空目录
func IsEmptyDir(path string) bool {
	if !IsDir(path) {
		// 非目录自然也不是空目录
		return false
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		// 列出目录错误
		return false
	}
	return len(entries) == 0
}
