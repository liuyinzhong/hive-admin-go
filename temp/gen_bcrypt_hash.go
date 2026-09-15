package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 一次性脚本：为存量用户批量重置生成 bcrypt 哈希，输出到迁移 SQL 使用
	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(hash))
}
