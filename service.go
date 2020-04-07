package main

import (
	"fmt"
	"log"
	"net/http"
	consulapi "github.com/hashicorp/consul/api"
	"time"
)


//参考 ：https://github.com/changjixiong/goNotes/blob/master/consulnotes/client/main.go
// 参考 ： https://www.cnblogs.com/hcy-fly/p/10826607.html
// 这两个demo相似
var PORT = 9500
var CONSUL_ADDR = "127.0.0.1:8500"
var DRIVER_NAME = "smartLamp"

func consulCheck(w http.ResponseWriter, r *http.Request) {
	log.Println("[consulCheck]")
	fmt.Fprintln(w, "status ok")
}

// 取消consul注册的服务
func ConsulDeRegister() {
	// 创建连接consul服务配置
	config := consulapi.DefaultConfig()
	config.Address = CONSUL_ADDR
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul client error : ", err)
	}

	client.Agent().ServiceDeregister(DRIVER_NAME)
}
// 健康检查参考：https://vizee.org/2017/12/24/grpc-service-discovery-with-consul/
func registerServer() {
	const ttl = 30 * time.Second
	config := consulapi.DefaultConfig()
	config.Address = CONSUL_ADDR
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul client error : ", err)
	}
	agent := client.Agent()

	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = DRIVER_NAME                    // 保证唯一性，驱动id。一种设备一个解析器。
	registration.Name = DRIVER_NAME  //这个名字可以重复.也是厂商id。针对不同公司可以有不同的解析规则. 这个参数应该在回调的时候传递过来
	registration.Port = PORT
	registration.Tags = []string{"support version v1,v2,v3,v5"}
	registration.Address = "127.0.0.1" // 自己的服务所在地址.这里用阿里云
	registration.Check= &consulapi.AgentServiceCheck{
		TTL:     (ttl + time.Second).String(),
		Timeout: time.Minute.String(),
	}

	go func() {
		//log.Println("checkid")
		checkid := "service:" + DRIVER_NAME
		for range time.Tick(ttl) {
			//log.Println("checkid")
			err := client.Agent().PassTTL(checkid, "")
			if err != nil {
				log.Fatalln(err)
			}
		}
	}()
	err = agent.ServiceRegister(registration)
	if err != nil {
		log.Fatal("register server error : ", err)
	}

}
