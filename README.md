# protoc-gen-go-hz

[English](#english-documentation) | [中文](./README_CN.md)

---

## English Documentation

This is a protoc plugin based on [CloudWeGo Hertz](https://github.com/cloudwego/hertz) for generating HTTP code, model code, and project layouts for Hertz projects from protobuf files.

### Features

- **Project Template Generation**: Support for generating standard Hertz project layouts
- **Model Code Generation**: Generate Go structs from protobuf messages
- **HTTP Code Generation**: Generate handler, router, and client code
- **Flexible Configuration**: Support for multiple customization options
- **Automatic Command Detection**: Automatically determine the operation type based on existing project structure
- **Automatic Module Extraction**: Automatically extract module information from proto's go_package option

### Quick Start

#### Installation

Install the latest version using go install:

```bash
go install github.com/ca-x/protoc-gen-go-hz@latest
```

#### Basic Usage

```bash
protoc --go-hz_out=. example.proto
```

This command will automatically determine the operation type based on the existing project structure:
- If the project doesn't exist, it will create a new project (new command)
- If the project exists, it will update code in the existing project (update command)

If needed, you can also override the automatic detection by explicitly specifying the command type using `cmd_type` parameter:

```bash
# Force new project creation (override auto-detection)
protoc --go-hz_out=. --go-hz_opt=cmd_type=new example.proto

# Force update (override auto-detection)
protoc --go-hz_out=. --go-hz_opt=cmd_type=update example.proto
```

#### Common Use Cases

##### Generate a New Project

```bash
protoc --go-hz_out=out_dir=. example.proto
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

##### Generate Model Code Only

```bash
protoc --go-hz_out=. --go-hz_opt=model=true example.proto
```

##### Generate Client Code

```bash
protoc --go-hz_out=. --go-hz_opt=client_dir=biz/client example.proto
```

#### Parameter Options

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `out_dir` | string | "." | Output directory |
| `handler_dir` | string | "biz/handler" | Handler code output directory |
| `model_dir` | string | "biz/model" | Model code output directory |
| `router_dir` | string | "biz/router" | Router code output directory |
| `client_dir` | string | "biz/client" | Client code output directory |
| `cmd_type` | string | "" | Command type: "new", "update", "model", "client" (optional, auto-detected by default) |
| `model` | bool | false | Generate model code only (OnlyModel flag) - **Note: use protoc-gen-go instead** |
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

#### Example Protobuf File

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

#### Differences from Original hz Tool

1. **Runs as a protoc Plugin**: Directly integrated into the protoc workflow
2. **Simplified Configuration**: Configure via command-line parameters, no .hz file needed
3. **Standard protoc Interface**: Conforms to the protoc plugin standard
4. **Modular Design**: Clear separation of different generation functions
5. **Automatic Detection**: Automatically extract module information and detect command type from project structure
6. **Follows protoc Plugin Best Practices**:
   - Model code generation delegated to `protoc-gen-go` (single responsibility)
   - Uses relative paths for output files
   - Graceful handling of unknown parameters
   - Smart auto-detection with manual override support

### Best Practices

#### Recommended Usage Pattern

For best results, use this plugin together with `protoc-gen-go`:

```bash
# Generate both protobuf models and Hertz HTTP code
protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-hz_out=. --go-hz_opt=handler_dir=biz/handler,router_dir=biz/router \
  example.proto
```

The plugin will automatically detect whether to create a new project or update an existing one.

#### When to Specify `cmd_type`

The `cmd_type` parameter is optional. Specify it only when you need to override the auto-detection:

```bash
# Force new project creation even if directories exist
protoc --go-hz_out=. --go-hz_opt=cmd_type=new example.proto

# Force update even in an empty directory
protoc --go-hz_out=. --go-hz_opt=cmd_type=update example.proto
```

#### Use protoc-gen-go for Models

Don't use the `model=true` option. Instead, let `protoc-gen-go` handle model generation:

```bash
# Correct way
protoc --go_out=. --go_opt=paths=source_relative example.proto

# Not recommended
protoc --go-hz_out=. --go-hz_opt=model=true example.proto
```

### Usage with Buf

You can also use protoc-gen-go-hz with [Buf](https://buf.build/), a modern tool for Protocol Buffers.

#### Installation

First, install the plugin:

```bash
go install github.com/ca-x/protoc-gen-go-hz@latest
```

#### Configure buf.yaml

Create a `buf.yaml` file in your project root:

```yaml
version: v1
deps:
  - buf.build/googleapis/googleapis
lint:
  use:
    - DEFAULT
breaking:
  use:
    - FILE
```

#### Configure buf.gen.yaml

Create a `buf.gen.yaml` file to configure code generation:

```yaml
version: v1
plugins:
  - plugin: go
    out: gen/go
    opt:
      - paths=source_relative
  - plugin: go-hz
    out: .
    opt:
      - model=true  # Generate model code only
      # - handler_dir=biz/handler
      # - router_dir=biz/router
      # - client_dir=biz/client
```

#### Generate Code

Run the following command to generate code:

```bash
buf generate
```

This will generate both the standard protobuf Go code and the Hertz HTTP code in the specified directories.

#### Example buf.gen.yaml for Full Project Generation

```yaml
version: v1
plugins:
  - plugin: go
    out: gen/go
    opt:
      - paths=source_relative
  - plugin: go-hz
    out: .
    opt:
      - handler_dir=biz/handler
      - router_dir=biz/router
      - model_dir=biz/model
      - client_dir=biz/client
      - need_go_mod=true
```

### Development

#### Build the Project

```bash
go build -o protoc-gen-go-hz .
```

#### Test the Plugin

```bash
protoc --go-hz_out=. example.proto
```

### Dependencies

- [CloudWeGo Hertz](https://github.com/cloudwego/hertz) (v0.9.7+)
- Protocol Buffers
- Go 1.21+

### License

Apache License 2.0
