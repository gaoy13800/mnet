package main

import (
	"fmt"
	"mnet/run"
)

func main() {

	fmt.Println("程序启动！")

	defer fmt.Println("程序结束！")

	//开始跑程序

	run.Run()

	//堵塞主程

	select {}
}
