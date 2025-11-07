# protoc-gen-go-hz 路径设计说明

## 设计原理

### 1. proto文件的go_package是唯一的模块信息来源

protoc插件**无法**直接读取项目的 `go.mod` 文件，只能从proto文件的 `go_package` 选项获取模块信息。

```protobuf
// example.proto
syntax = "proto3";
package example;

// 这是唯一的模块信息来源
option go_package = "github.com/example/project/biz/model";
```

### 2. 文件输出路径 vs Go Import路径

protoc插件需要处理两种路径：

#### (1) 文件输出路径 (File Path)
- **用途**: 告诉protoc把生成的文件写到哪里
- **格式**: **相对路径**，相对于 `--go-hz_out` 指定的输出目录
- **示例**: `biz/handler/SayHello.go`

#### (2) Go Import路径 (Go Import Path)
- **用途**: 生成代码中的import语句
- **格式**: **完整的Go模块路径**
- **示例**: `github.com/example/project/biz/handler`

### 3. NewGeneratedFile API的正确用法

```go
func (gen *Plugin) NewGeneratedFile(
    filename string,        // 相对路径: "biz/handler/SayHello.go"
    goImportPath GoImportPath  // Go import路径: "github.com/example/project/biz/handler"
) *GeneratedFile
```

## 实现细节

### Step 1: 提取Go模块根路径

从 `go_package` 提取模块根路径（去掉 `/biz/xxx` 部分）:

```go
// go_package: "github.com/example/project/biz/model"
// 提取结果: "github.com/example/project"

if idx := strings.Index(goPackage, "/biz/"); idx != -1 {
    moduleRoot = goPackage[:idx]
}
```

### Step 2: 生成文件使用相对路径

```go
// http_package.go
path := pkgGen.HandlerDir + "/" + method.Name + ".go"
// 结果: "biz/handler/SayHello.go"
```

### Step 3: 构建每个文件的Go Import路径

```go
func (p *HZPlugin) buildGoImportPath(filePath string) string {
    // 从文件路径提取包路径
    dir := filepath.Dir(filePath)  // "biz/handler"

    // 拼接模块根路径
    return p.args.Gomod + "/" + dir
    // 结果: "github.com/example/project/biz/handler"
}
```

### Step 4: 正确传递给protogen

```go
for _, file := range files {
    goImportPath := p.buildGoImportPath(file.Path)
    g := p.gen.NewGeneratedFile(file.Path, protogen.GoImportPath(goImportPath))
    g.P(file.Content)
}
```

## 最终效果

### 输入
```protobuf
option go_package = "github.com/example/project/biz/model";
```

### 输出

**文件路径** (相对):
```
biz/handler/SayHello.go
biz/handler/SayGoodbye.go
biz/router/router.go
```

**生成代码中的import** (完整):
```go
// biz/handler/SayHello.go
import (
    "github.com/cloudwego/hertz/pkg/app"
    "github.com/example/project/biz/model"  // ✅ 正确
)

// biz/router/router.go
import (
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/example/project/biz/handler"  // ✅ 正确
)
```

## 为什么这样设计？

### 1. 符合protoc插件标准
- 输出路径使用相对路径是protoc的约定
- protoc根据 `--go-hz_out` 参数决定最终写入位置

### 2. 灵活的输出位置
用户可以指定任意输出目录：
```bash
# 输出到当前目录
protoc --go-hz_out=. example.proto

# 输出到其他目录
protoc --go-hz_out=/tmp/generated example.proto
```

### 3. import路径准确
- 基于proto的 `go_package` 构建
- 不依赖文件系统状态
- 确保生成的代码可以正确编译

## 与你的观点的对应

> "protoc插件如果不能获取外部gomod的名称，那么我觉得相对路径可能才是对的。
> 如果可以获取，那么我觉得生成的导入路径包含外部gomod的名称是对的"

**答案**: protoc插件可以从 `go_package` 获取模块信息，所以：

✅ **文件路径**: 使用相对路径（符合protoc标准）
✅ **Import路径**: 使用完整模块路径（基于go_package）

两者结合，既符合protoc标准，又保证了import的正确性！
