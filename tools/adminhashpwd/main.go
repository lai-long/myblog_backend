package main

import (
	"fmt"
	"os"

	"myblog_backend/pkg/pwd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go <密码>")
		return
	}
	hash, err := pwd.Hash(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println(hash)
}
