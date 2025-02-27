package utils

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
)

// PathExists 判断文件是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// VolumeExtract 通过:分割volume目录, -v /tmp:/tmp
// 返回源路径sourcePath，目标路径destPath
func VolumeExtract(volume string) (string, string, error) {
	parts := strings.Split(volume, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid volume [%s], must split by `:`", volume)
	}

	sourcePath, destPath := parts[0], parts[1]
	if sourcePath == "" || destPath == "" {
		return "", "", fmt.Errorf("invalid volume [%s], path can't be empty", volume)
	}

	return sourcePath, destPath, nil
}

// RandStringBytes 随机生成数字字符串
func RandStringBytes(n int) string {
	letterBytes := "123456789"
	res := make([]byte, n)
	for i := range res {
		res[i] = letterBytes[rand.Intn(9)]
	}
	return string(res)
}
