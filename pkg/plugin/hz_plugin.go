/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugin

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/ca-x/protoc-gen-hz/pkg/config"
    "github.com/ca-x/protoc-gen-hz/pkg/generator"
    "github.com/cloudwego/hertz/cmd/hz/generator/model"
    "github.com/cloudwego/hertz/cmd/hz/meta"
    "github.com/cloudwego/hertz/cmd/hz/util/logs"
    "github.com/sirupsen/logrus"
    "google.golang.org/protobuf/compiler/protogen"
    "google.golang.org/protobuf/reflect/protoreflect"
    "google.golang.org/protobuf/runtime/protoimpl"
)

// HZPlugin 是HZ protoc插件的主体
type HZPlugin struct {
    gen   *protogen.Plugin
    args  *config.Argument
    logger *logrus.Logger
}

// NewHZPlugin 创建新的HZ插件实例
func NewHZPlugin(gen *protogen.Plugin) *HZPlugin {
    return &HZPlugin{
        gen:    gen,
        logger: logrus.New(),
    }
}

// Run 运行插件
func (p *HZPlugin) Run() error {
    // 解析插件参数
    if err := p.parseArgs(); err != nil {
        return fmt.Errorf("parse args failed: %w", err)
    }

    // 设置日志级别
    if p.args.Verbose {
        p.logger.SetLevel(logrus.DebugLevel)
        logs.SetLevel(logs.LevelDebug)
    }

    p.logger.Debug("HZ protoc plugin started")

    // 根据命令类型执行不同的操作
    switch p.args.CmdType {
    case meta.CmdNew:
        return p.handleNewCommand()
    case meta.CmdUpdate:
        return p.handleUpdateCommand()
    case meta.CmdModel:
        return p.handleModelCommand()
    case meta.CmdClient:
        return p.handleClientCommand()
    default:
        // 默认行为：生成所有代码
        return p.handleDefault()
    }
}

// parseArgs 解析插件参数
func (p *HZPlugin) parseArgs() error {
    // 从protogen插件获取参数
    param := p.gen.Request.GetParameter()
    if param == "" {
        param = "model=true" // 默认生成模型
    }

    p.args = &config.Argument{}
    params := strings.Split(param, ",")
    if err := p.args.Unpack(params); err != nil {
        return err
    }

    // 设置默认值
    if p.args.CmdType == "" {
        p.args.CmdType = meta.CmdNew
    }

    return nil
}

// handleNewCommand 处理new命令，生成项目布局和代码
func (p *HZPlugin) handleNewCommand() error {
    p.logger.Info("Handling new command")

    // 1. 生成项目布局
    if err := p.generateLayout(); err != nil {
        return fmt.Errorf("generate layout failed: %w", err)
    }

    // 2. 生成模型代码
    if err := p.generateModels(); err != nil {
        return fmt.Errorf("generate models failed: %w", err)
    }

    // 3. 生成HTTP代码
    if err := p.generateHTTPCode(); err != nil {
        return fmt.Errorf("generate http code failed: %w", err)
    }

    return nil
}

// handleUpdateCommand 处理update命令，更新现有项目
func (p *HZPlugin) handleUpdateCommand() error {
    p.logger.Info("Handling update command")

    // 1. 生成模型代码
    if err := p.generateModels(); err != nil {
        return fmt.Errorf("generate models failed: %w", err)
    }

    // 2. 生成HTTP代码
    if err := p.generateHTTPCode(); err != nil {
        return fmt.Errorf("generate http code failed: %w", err)
    }

    return nil
}

// handleModelCommand 处理model命令，只生成模型代码
func (p *HZPlugin) handleModelCommand() error {
    p.logger.Info("Handling model command")

    return p.generateModels()
}

// handleClientCommand 处理client命令，生成客户端代码
func (p *HZPlugin) handleClientCommand() error {
    p.logger.Info("Handling client command")

    // 1. 生成模型代码
    if err := p.generateModels(); err != nil {
        return fmt.Errorf("generate models failed: %w", err)
    }

    // 2. 生成客户端代码
    if err := p.generateClientCode(); err != nil {
        return fmt.Errorf("generate client code failed: %w", err)
    }

    return nil
}

// handleDefault 默认处理，生成所有代码
func (p *HZPlugin) handleDefault() error {
    p.logger.Info("Handling default generation")

    // 生成所有类型的代码
    return p.handleNewCommand()
}

// generateLayout 生成项目布局
func (p *HZPlugin) generateLayout() error {
    if p.args.OutDir == "" {
        p.args.OutDir = "."
    }

    layoutGen := &generator.LayoutGenerator{
        TemplateGenerator: generator.TemplateGenerator{
            OutputDir: p.args.OutDir,
            Excludes:  p.args.Excludes,
        },
    }

    layout := generator.Layout{
        GoModule:        p.args.Gomod,
        ServiceName:     p.args.ServiceName,
        UseApacheThrift: false, // protobuf项目不使用thrift
        HasIdl:          true,
        ModelDir:        p.args.ModelDir,
        HandlerDir:      p.args.HandlerDir,
        RouterDir:       p.args.RouterDir,
        NeedGoMod:       p.args.NeedGoMod,
    }

    if err := layoutGen.GenerateByService(layout); err != nil {
        return err
    }

    return layoutGen.Persist()
}

// generateModels 生成模型代码
func (p *HZPlugin) generateModels() error {
    for _, file := range p.gen.Files {
        if file.Generate {
            if err := p.generateFile(file); err != nil {
                return fmt.Errorf("generate file %s failed: %w", file.Proto.GetName(), err)
            }
        }
    }
    return nil
}

// generateFile 生成单个protobuf文件的模型代码
func (p *HZPlugin) generateFile(file *protogen.File) error {
    // 使用hz的protobuf生成逻辑
    filename := file.GeneratedFilenamePrefix + ".pb.go"
    g := p.gen.NewGeneratedFile(filename, file.GoImportPath)
    
    // 生成基本的文件头
    g.P("// Code generated by protoc-gen-hertz. DO NOT EDIT.")
    g.P("// versions:")
    g.P("// \tprotoc-gen-go v1.31.0")
    g.P("// \tprotoc        v3.21.0")
    g.P("// source: ", file.Proto.GetName())
    g.P()
    g.P("package ", file.GoPackageName)
    g.P()

    // 生成基本的导入
    g.P("import (")
    g.P(`protoreflect "google.golang.org/protobuf/reflect/protoreflect"`)
    g.P(`protoimpl "google.golang.org/protobuf/runtime/protoimpl"`)
    g.P(`reflect "reflect"`)
    g.P(`sync "sync"`)
    g.P(")")
    g.P()

    // 生成版本检查
    g.P("const (")
    g.P("// Verify that this generated code is sufficiently up-to-date.")
    g.P("_ = protoimpl.EnforceVersion(", protoimpl.GenVersion, " - ", protoimpl.MinVersion, ")")
    g.P("// Verify that runtime/protoimpl is sufficiently up-to-date.")
    g.P("_ = protoimpl.EnforceVersion(", protoimpl.MaxVersion, " - ", protoimpl.GenVersion, ")")
    g.P(")")
    g.P()

    // 生成消息和枚举
    for _, enum := range file.Enums {
        p.generateEnum(g, enum)
    }
    
    for _, message := range file.Messages {
        p.generateMessage(g, message)
    }

    return nil
}

// generateEnum 生成枚举
func (p *HZPlugin) generateEnum(g *protogen.GeneratedFile, enum *protogen.Enum) {
    g.P("type ", enum.GoIdent, " int32")
    g.P("const (")
    for _, value := range enum.Values {
        g.P(value.GoIdent, " ", enum.GoIdent, " = ", value.Desc.Number())
    }
    g.P(")")
    g.P()

    // 生成枚举的变量和方法
    g.P("var (")
    g.P(enum.GoIdent, "_name = map[int32]string{")
    for _, value := range enum.Values {
        g.P(value.Desc.Number(), ": ", `"` + string(value.Desc.Name()) + `",`)
    }
    g.P("}")
    g.P(enum.GoIdent, "_value = map[string]int32{")
    for _, value := range enum.Values {
        g.P(`"` + string(value.Desc.Name()) + `": `, value.Desc.Number(), ",")
    }
    g.P("}")
    g.P(")")
    g.P()

    g.P("func (x ", enum.GoIdent, ") Enum() *", enum.GoIdent, " {")
    g.P("p := new(", enum.GoIdent, ")")
    g.P("*p = x")
    g.P("return p")
    g.P("}")
    g.P()

    g.P("func (x ", enum.GoIdent, ") String() string {")
    g.P("return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))")
    g.P("}")
    g.P()

    g.P("func (*", enum.GoIdent, ") Descriptor() protoreflect.EnumDescriptor {")
    g.P("return file_", enum.GoIdent.GoName, "_proto.EnumTypes()[0].Descriptor()")
    g.P("}")
    g.P()
}

// generateMessage 生成消息
func (p *HZPlugin) generateMessage(g *protogen.GeneratedFile, message *protogen.Message) {
    if message.Desc.IsMapEntry() {
        return
    }

    g.P("type ", message.GoIdent, " struct {")
    for _, field := range message.Fields {
        // 获取字段类型
        var fieldType string
        switch field.Desc.Kind() {
        case protoreflect.StringKind:
            fieldType = "string"
        case protoreflect.Int32Kind:
            fieldType = "int32"
        case protoreflect.Int64Kind:
            fieldType = "int64"
        case protoreflect.BoolKind:
            fieldType = "bool"
        default:
            fieldType = "string" // 默认为string
        }
        
        g.P(field.GoName, " ", fieldType, " `", string(field.Desc.Name()), ":", string(field.Desc.Number()), "`")
    }
    g.P("}")
    g.P()

    // 生成基本方法
    g.P("func (x *", message.GoIdent, ") Reset() {")
    g.P("*x = ", message.GoIdent, "{}")
    g.P("}")
    g.P()

    g.P("func (x *", message.GoIdent, ") String() string {")
    g.P("return protoimpl.X.MessageStringOf(x)")
    g.P("}")
    g.P()

    g.P("func (*", message.GoIdent, ") ProtoMessage() {}")
    g.P()

    g.P("func (x *", message.GoIdent, ") Descriptor() protoreflect.MessageDescriptor {")
    g.P("return file_", message.GoIdent.GoName, "_proto.MsgTypes()[0].Descriptor()")
    g.P("}")
    g.P()
}

// generateHTTPCode 生成HTTP相关代码
func (p *HZPlugin) generateHTTPCode() error {
    p.logger.Debugf("Generating HTTP code with args: %+v", p.args)

    // 创建HTTP包生成器
    pkgGen := &generator.HTTPPackageGenerator{
        CmdType:        p.args.CmdType,
        ProjPackage:    p.args.Gomod,
        HandlerDir:     p.args.HandlerDir,
        RouterDir:      p.args.RouterDir,
        ModelDir:       p.args.ModelDir,
        ClientDir:      p.args.ClientDir,
        BaseDomain:     p.args.BaseDomain,
        HandlerByMethod: p.args.HandlerByMethod,
        SortRouter:     p.args.SortRouter,
    }

    p.logger.Debugf("Created HTTP package generator: %+v", pkgGen)

    // 初始化生成器
    if err := pkgGen.Init(); err != nil {
        return fmt.Errorf("init http package generator failed: %w", err)
    }

    // 构建HTTP包数据
    httpPkg := p.buildHTTPPackage()
    p.logger.Debugf("Built HTTP package: %+v", httpPkg)
    
    // 生成代码
    files, err := pkgGen.Generate(httpPkg)
    if err != nil {
        return err
    }

    p.logger.Debugf("Generated %d files", len(files))

    // 将生成的文件添加到protogen响应
    for _, file := range files {
        p.logger.Debugf("Adding file: %s", file.Path)
        g := p.gen.NewGeneratedFile(file.Path, p.gen.Files[0].GoImportPath)
        g.P(file.Content)
    }

    return nil
}

// generateClientCode 生成客户端代码
func (p *HZPlugin) generateClientCode() error {
    // 客户端代码生成逻辑
    p.logger.Info("Generating client code")
    return nil
}

// buildHTTPPackage 构建HTTP包数据结构
func (p *HZPlugin) buildHTTPPackage() *generator.HTTPPackage {
    httpPkg := &generator.HTTPPackage{
        IdlName: p.getMainIDLName(),
        Package: p.args.Gomod,
        Services: []*generator.Service{},
        Models:   []*model.Model{},
        RouterInfo: &generator.Router{},
    }

    // 解析protobuf文件，提取服务信息
    for _, file := range p.gen.Files {
        if file.Generate {
            for _, service := range file.Services {
                svc := &generator.Service{
                    Name:   string(service.GoName),
                    Methods: []*generator.HTTPMethod{},
                    ClientMethods: []*generator.ClientMethod{},
                    Models: []*model.Model{},
                    BaseDomain: p.args.BaseDomain,
                }

                // 提取方法信息
                for _, method := range service.Methods {
                    httpMethod := &generator.HTTPMethod{
                        Name:         string(method.GoName),
                        HTTPMethod:   "POST", // 默认POST，可以通过注释配置
                        Path:         "/" + string(service.GoName) + "/" + string(method.GoName),
                        RequestType:  string(method.Input.GoIdent.GoName),
                        ResponseType: string(method.Output.GoIdent.GoName),
                    }
                    svc.Methods = append(svc.Methods, httpMethod)
                }

                httpPkg.Services = append(httpPkg.Services, svc)
            }
        }
    }

    return httpPkg
}

// getMainIDLName 获取主IDL文件名
func (p *HZPlugin) getMainIDLName() string {
    if len(p.gen.Request.FileToGenerate) > 0 {
        return filepath.Base(p.gen.Request.FileToGenerate[0])
    }
    return "unknown.proto"
}