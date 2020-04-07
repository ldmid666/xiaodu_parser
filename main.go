package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("start")
	registerServer() // 注册服务
	 // 运行服务
	fmt.Printf("wait port =%d\r\n", PORT)
	sigChan := make(chan os.Signal)
	exitChan := make(chan struct{})
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		err :=  CreateNewParser()
		if err != nil {
			log.Println(err)
		}
		log.Println("stopping ")
		exitChan <- struct{}{}
	}()
	//两种结束程序的方式
	select {
	case <-exitChan:
	case s := <-sigChan:
		log.Println("signal exit", s)
	}
	ConsulDeRegister()

}
