# protoc-gen-hertz

[English](#english) | [中文](#chinese)

---

## Chinese

### 中文文档

这是一个基于 CloudWeGo Hertz 的 protoc 插件，用于从 protobuf 文件生成 Hertz 项目的 HTTP 代码、模型代码和项目布局。

#### 功能特性

- **项目模板生成**: 支持生成标准的 Hertz 项目布局
- **模型代码生成**: 从 protobuf 消息生成 Go 结构体
- **HTTP 代码生成**: 生成 handler、router 和 client 代码
- **灵活的配置选项**: 支持多种自定义配置选项
- **自动命令检测**: 根据现有项目结构自动判断生成类型
- **模块信息自动提取**: 从 proto 的 go_package 选项自动提取模块信息

#### 快速开始

##### 安装

使用 go install 方式安装最新版本：

```bash
go install github.com/ca-x/protoc-gen-hz@latest
```

##### 基本用法

```bash
protoc --hertz_out=. example.proto
```

该命令会根据现有项目结构自动判断操作类型：
- 如果项目不存在，会创建新项目（检测到 handler/ 和 router/ 目录）
- 如果项目已存在，会在现有项目中生成代码

##### 常见使用场景

###### 生成新项目

```bash
protoc --hertz_out=out_dir=. example.proto
```

生成的项目结构如下：

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
├── go.mod
├── go.sum
└── main.go
```

###### 只生成模型代码

```bash
protoc --hertz_out=. --hertz_opt=model=true example.proto
```

###### 生成客户端代码

```bash
protoc --hertz_out=. --hertz_opt=client_dir=biz/client example.proto
```

##### 参数选项

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `out_dir` | string | "." | 输出目录 |
| `handler_dir` | string | "biz/handler" | handler 代码输出目录 |
| `model_dir` | string | "biz/model" | 模型代码输出目录 |
| `router_dir` | string | "biz/router" | 路由代码输出目录 |
| `client_dir` | string | "biz/client" | 客户端代码输出目录 |
| `model` | bool | false | 仅生成模型代码（对应 OnlyModel 标志） |
| `verbose` | bool | false | 启用详细输出 |
| `base_domain` | string | "" | 基础域名 |
| `service` | string | "" | 服务名称 |
| `use` | string | "" | 自定义使用选项 |
| `need_go_mod` | bool | false | 是否需要生成 go.mod 文件 |
| `json_enumstr` | bool | false | JSON 枚举使用字符串 |
| `query_enumint` | bool | false | Query 参数枚举使用整数 |
| `unset_omitempty` | bool | false | 取消 omitempty 标签 |
| `pb_camel_json_tag` | bool | false | protobuf 字段使用驼峰命名的 JSON 标签 |
| `snake_tag` | bool | false | 使用蛇形命名的标签 |
| `no_recurse` | bool | false | 不递归处理导入的 proto 文件 |
| `handler_by_method` | bool | false | 每个方法生成单独的 handler 文件 |
| `sort_router` | bool | false | 对路由进行排序 |
| `force_client` | bool | false | 强制生成客户端代码 |
| `customize_layout` | string | "" | 自定义项目布局 |
| `customize_package` | string | "" | 自定义包名 |

##### 示例 Protobuf 文件

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

##### 与原始 hz 工具的区别

1. **作为 protoc 插件运行**: 直接集成到 protoc 工作流中
2. **简化的配置**: 通过命令行参数配置，不需要 .hz 文件
3. **标准 protoc 接口**: 符合 protoc 插件标准
4. **模块化设计**: 清晰分离不同的生成功能
5. **自动检测**: 自动从 proto 文件提取模块信息和命令类型

#### 开发

##### 构建项目

```bash
go build -o protoc-gen-hertz .
```

##### 测试插件

```bash
protoc --hertz_out=. example.proto
```

#### 依赖

- CloudWeGo Hertz (v0.9.7+)
- Protocol Buffers
- Go 1.21+

#### 许可证

Apache License 2.0

---

## English

### English Documentation

This is a protoc plugin based on CloudWeGo Hertz for generating HTTP code, model code, and project layouts for Hertz projects from protobuf files.

#### Features

- **Project Template Generation**: Support for generating standard Hertz project layouts
- **Model Code Generation**: Generate Go structs from protobuf messages
- **HTTP Code Generation**: Generate handler, router, and client code
- **Flexible Configuration**: Support for multiple customization options
- **Automatic Command Detection**: Automatically determine the operation type based on existing project structure
- **Automatic Module Extraction**: Automatically extract module information from proto's go_package option

#### Quick Start

##### Installation

Install the latest version using go install:

```bash
go install github.com/ca-x/protoc-gen-hz@latest
```

##### Basic Usage

```bash
protoc --hertz_out=. example.proto
```

This command will automatically determine the operation type based on the existing project structure:
- If the project doesn't exist, it will create a new project (by detecting handler/ and router/ directories)
- If the project exists, it will generate code in the existing project

##### Common Use Cases

###### Generate a New Project

```bash
protoc --hertz_out=out_dir=. example.proto
```

The generated project structure will be:

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
├── go.mod
├── go.sum
└── main.go
```

###### Generate Model Code Only

```bash
protoc --hertz_out=. --hertz_opt=model=true example.proto
```

###### Generate Client Code

```bash
protoc --hertz_out=. --hertz_opt=client_dir=biz/client example.proto
```

##### Parameter Options

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `out_dir` | string | "." | Output directory |
| `handler_dir` | string | "biz/handler" | Handler code output directory |
| `model_dir` | string | "biz/model" | Model code output directory |
| `router_dir` | string | "biz/router" | Router code output directory |
| `client_dir` | string | "biz/client" | Client code output directory |
| `model` | bool | false | Generate model code only (OnlyModel flag) |
| `verbose` | bool | false | Enable verbose output |
| `base_domain` | string | "" | Base domain |
| `service` | string | "" | Service name |
| `use` | string | "" | Custom use option |
| `need_go_mod` | bool | false | Whether to generate go.mod file |
| `json_enumstr` | bool | false | Use string for JSON enum |
| `query_enumint` | bool | false | Use integer for query enum |
| `unset_omitempty` | bool | false | Remove omitempty tag |
| `pb_camel_json_tag` | bool | false | Use camelCase JSON tag for protobuf fields |
| `snake_tag` | bool | false | Use snake_case tag |
| `no_recurse` | bool | false | Don't recursively process imported proto files |
| `handler_by_method` | bool | false | Generate separate handler file for each method |
| `sort_router` | bool | false | Sort routes |
| `force_client` | bool | false | Force client code generation |
| `customize_layout` | string | "" | Customize project layout |
| `customize_package` | string | "" | Customize package name |

##### Example Protobuf File

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

##### Differences from Original hz Tool

1. **Runs as a protoc Plugin**: Directly integrated into the protoc workflow
2. **Simplified Configuration**: Configure via command-line parameters, no .hz file needed
3. **Standard protoc Interface**: Conforms to the protoc plugin standard
4. **Modular Design**: Clear separation of different generation functions
5. **Automatic Detection**: Automatically extract module information and command type from proto files

#### Development

##### Build the Project

```bash
go build -o protoc-gen-hertz .
```

##### Test the Plugin

```bash
protoc --hertz_out=. example.proto
```

#### Dependencies

- CloudWeGo Hertz (v0.9.7+)
- Protocol Buffers
- Go 1.21+

#### License

Apache License 2.0
