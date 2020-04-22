### 使用说明

使用说明参考工程[iot_gateway_v1]()

### 其他说明

命令：

protoc --go_out=plugins=grpc:. parser.proto

protoc --go_out=. gateway.proto

云端端口映射到本地，方面调试开发：

ssh -gR :8000:localhost:8000 root@aliyun

压缩：

upx ./bin/open