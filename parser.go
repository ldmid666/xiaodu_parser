package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"net"
	"strconv"
	pb "xiaodu_parser/grpc_proto"
	devProto "xiaodu_parser/dev_proto"
	"google.golang.org/grpc"
	"github.com/golang/protobuf/proto"

)
const (
	Category_CMD    uint32 = 0
	Category_CONFIG uint32 = 1
)
const (
	TYPE_CMD = "cmd"
	CMD_LAMP = "lamp"
	CMD_ON = "on"
	CMD_OFF = "off"
	TYPE_CONFIG = "config"
	CONFIG_VOLUME = "VOLUME"
)
/**
复制的 .pd.go文件的
type HelloserverClient interface {
   //一个打招呼的函数
   Sayhello(ctx context.Context, in *HelloReq, opts ...grpc.CallOption) (*HelloRsp, error)
   //一个说名字的服务
   Sayname(ctx context.Context, in *NameReq, opts ...grpc.CallOption) (*NameRsq, error)
}
*/
type server struct{}

//rpc
//函数关键字（对象）函数名（客户端发送过来的内容 ， 返回给客户端的内容） 错误返回值

//grpc
//函数关键字 （对象）函数名 （cotext，客户端发过来的参数 ）（发送给客户端的参数，错误）

//进行改下,绑定到你自己的结构体,参数带上包名
func (this *server) Marshal(ctx context.Context, in *pb.DownReq) (out *pb.DownRsp, err error) {
	log.Println("[Marshal]", in.Name)
	if ctx.Err() == context.Canceled {
		log.Println("timeout")
		return nil,errors.New ("[Marshal]timeout")
	}

	//应用层下发的指令可能有多种组合
	req := make(map[string]interface{})
	err = json.Unmarshal(in.Payload,&req)
	if err != nil {
		log.Println("[Marshal]req error",in.Name,string(in.Payload),err)
		err =  errors.Wrap(err,"[Marshal]Unmarshal error")
		return nil,err
	}
	downMsg := &devProto.Payload{}
	kind := mustString(req["kind"])
	filed := mustString(req["field"])
	if kind == TYPE_CMD{
		downMsg.Kind = uint32(devProto.Category_CMD)
		switch filed {
		case CMD_LAMP:
			// 下发给灯的值，是保留几位小数，要不要扩大10倍数。要不要把下发的值范围由100~1 量化成中控的范围10~1等逻辑处理等等
			// 应该交给解析器完成。既不要交给上层应用，因为他们没必要了解终端逻辑和数据类型。也不要交给终端完成这些，因为能在解
			// 析器完成，就尽量不要在终端。因为涉及日后更新，bug修复等，终端改改起来成本都很高。
			downMsg.Key = uint32(devProto.Device_LAMP)
			if mustString(req["val"]) == CMD_ON{
				downMsg.Val = []byte{uint8(devProto.Operation_ON)}
			}else {
				downMsg.Val=[]byte{uint8(devProto.Operation_OFF)}
			}
			payload , err :=CreateMarshalDown(downMsg)
			if err != nil {
				log.Println("[Marshal]CreateMarshalDown error",err)
				err = errors.Wrap(err,"CreateMarshalDown error")
				return nil ,err
			}
			out = &pb.DownRsp{}
			out.ID = in.ID
			out.Name = in.Name
			out.Payload = payload

			return  out,nil
		default:
			return nil,errors.New ("[Marshal]cmd filed error")
		}
	}else if kind == TYPE_CONFIG {
			//TODO:配置参数的修改
	}else{
		return nil,errors.New ("[Marshal]kind error")
	}

	err = errors.New("Unknown error type")
	return nil,err
}

// Name 字段表示不同设备定制厂商名称。同样一款产品，定制厂商要求配置可能不一样的。解析结果也不一样
// 对于不同版本解析方式的不同，通过payload的首字段来表示版本号
func (this *server) UnMarshal(ctx context.Context, in *pb.UpReq) (out *pb.UpRsp, err error) {
	log.Println("[Marshal]", in.Name)
	if ctx.Err() == context.Canceled {
		return nil,errors.New ("[Marshal]timeout")
	}
	payDecrypt,err := decrypt(in.Payload)
	if err != nil {
		log.Println("[controlHandle]Decrypt error",err)
		return nil,err
	}
	upPayload := &devProto.Payload{}
	err = proto.Unmarshal(payDecrypt, upPayload)
	if err != nil {
		log.Println("[controlHandle] proto Unmarshal error",err)
		return  nil ,err
	}
	resp := make(map[string]interface{})
	resp["kind"] = TYPE_CMD
	resp["field"] = CMD_LAMP
	if 	upPayload.Val[0] == byte(devProto.Operation_ON){
		resp["val"] = "on"
	}else {
		resp["val"] = "off"
	}
	data ,err := json.Marshal(resp)
	out = &pb.UpRsp{}
	out.ID = in.ID
	out.Name = in.Name
	out.Payload = data
	return out, nil
}


//health
type HealthImpl struct{}

// Check 实现健康检查接口，这里直接返回健康状态，这里也可以有更复杂的健康检查策略，比如根据服务器负载来返回
func (h *HealthImpl) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	fmt.Print("health checking\n")
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h *HealthImpl) Watch(req *grpc_health_v1.HealthCheckRequest, w grpc_health_v1.Health_WatchServer) error {
	return nil
}

func CreateNewParser() error{
	ln, err := net.Listen("tcp", ":"+ strconv.Itoa(PORT))
	if err != nil {
		log.Println("网络错误: ", err)
		ConsulDeRegister()

		return errors.Wrap(err,"network error")
	}

	//创建grpc的服务
	srv := grpc.NewServer()
	//注册服务
	//pd就是protobuf那边的包,Register后面跟服务名
	pb.RegisterParserServer(srv, &server{})
	grpc_health_v1.RegisterHealthServer(srv, &HealthImpl{})

	//等待网络连接
	err = srv.Serve(ln)
	if err != nil {
		ConsulDeRegister()
		log.Println("网络错误: ", err)
		return errors.Wrap(err,"connect error")
	}
	// TODO:如果服务建立出现问题，应该讲注册的服务主动停掉
	return nil
}

func mustString(data interface{}) string {
	if data != nil && fmt.Sprintf("%T", data) == "string" {
		return data.(string)
	}
	return ""
}