# protoc-gen-hertz

这是一个基于CloudWeGo Hertz的protoc插件，用于从protobuf文件生成Hertz项目的HTTP代码、模型代码和项目布局。

## 功能特性

- **项目模板生成**: 支持生成标准的Hertz项目布局
- **模型代码生成**: 从protobuf消息生成Go结构体
- **HTTP代码生成**: 生成handler、router和client代码
- **可配置选项**: 支持多种自定义配置选项

## 安装

1. 构建插件：
```bash
go build -o protoc-gen-hertz .
```

2. 将插件添加到PATH：
```bash
export PATH=$PATH:$PWD
```

## 使用方法

### 基本用法

```bash
protoc --hertz_out=. --hertz_opt=command=new,module=github.com/example/project example.proto
```

### 参数选项

- `command`: 命令类型 (`new`, `update`, `model`, `client`)
- `module`: Go模块名
- `out_dir`: 输出目录 (默认: ".")
- `handler_dir`: handler目录 (默认: "biz/handler")
- `model_dir`: model目录 (默认: "biz/model")
- `router_dir`: router目录 (默认: "biz/router")
- `client_dir`: client目录
- `verbose`: 启用详细输出

### 示例

#### 生成新项目
```bash
protoc --hertz_out=. --hertz_opt=command=new,module=github.com/example/project,verbose=true example.proto
```

#### 只生成模型代码
```bash
protoc --hertz_out=. --hertz_opt=command=model,module=github.com/example/project example.proto
```

#### 生成客户端代码
```bash
protoc --hertz_out=. --hertz_opt=command=client,module=github.com/example/project,client_dir=biz/client example.proto
```

## 生成的文件结构

```
.
├── biz/
│   ├── handler/
│   │   ├── SayHello.go
│   │   └── SayGoodbye.go
│   ├── model/
│   │   └── example.pb.go
│   └── router/
│       └── router.go
└── go.mod (如果指定了module)
```

## 示例protobuf文件

```protobuf
syntax = "proto3";

package example;

option go_package = "github.com/example/project/biz/model";

service Greeter {
  rpc SayHello (HelloRequest) returns (HelloReply) {}
  rpc SayGoodbye (HelloRequest) returns (HelloReply) {}
}

message HelloRequest {
  string name = 1;
  int32 age = 2;
}

message HelloReply {
  string message = 1;
  int32 code = 2;
}
```

## 与原始hz工具的区别

1. **作为protoc插件运行**: 直接集成到protoc工作流中
2. **简化的配置**: 通过命令行参数配置，不需要.hz文件
3. **标准protoc接口**: 符合protoc插件标准
4. **模块化设计**: 清晰分离不同的生成功能

## 开发

### 构建项目

```bash
go build -o protoc-gen-hertz .
```

### 测试插件

```bash
protoc --hertz_out=. --hertz_opt=command=new,module=github.com/example/project example.proto
```

## 依赖

- CloudWeGo Hertz
- Protocol Buffers
- Go 1.21+

## 许可证

Apache License 2.0