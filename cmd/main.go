package main

import (
	"fmt"
	"github.com/husanpao/game-mitm/gosysproxy"
	"os"
	"os/signal"
	"syscall"
	"xyzw_study/web"
)

func main() {
	// 设置代理
	err := gosysproxy.SetGlobalProxy("127.0.0.1:12311", "localhost;127.*;10.*;172.16.*;172.17.*;172.18.*;172.19.*;172.20.*;172.21.*;172.22.*;172.23.*;172.24.*;172.25.*;172.26.*;172.27.*;172.28.*;172.29.*;172.30.*;172.31.*;192.168.*")
	if err != nil {
		panic(err)
	}

	// 确保在函数返回时关闭代理
	defer func() {
		fmt.Println("正在关闭系统代理...")
		gosysproxy.Off()
		fmt.Println("系统代理已关闭")
	}()

	// 设置信号处理，捕获中断信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go web.StartWebServer()

	// 等待中断信号
	<-c
	fmt.Println("收到退出信号，程序即将关闭...")
}
