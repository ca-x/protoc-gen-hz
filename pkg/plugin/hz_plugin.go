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
	"os"
	"path/filepath"
	"strings"

	"github.com/ca-x/protoc-gen-go-hz/pkg/config"
	"github.com/ca-x/protoc-gen-go-hz/pkg/generator"
	"github.com/cloudwego/hertz/cmd/hz/generator/model"
	"github.com/cloudwego/hertz/cmd/hz/meta"
	"github.com/cloudwego/hertz/cmd/hz/util/logs"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/compiler/protogen"
)

// HZPlugin 是HZ protoc插件的主体
type HZPlugin struct {
	gen    *protogen.Plugin
	args   *config.Argument
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

	// 如果只生成模型代码
	if p.args.OnlyModel {
		return p.handleModelCommand()
	}

	// 确定命令类型：优先使用显式指定，否则自动检测
	// 自动检测基于项目目录结构，这对于脚手架工具是合理的
	cmdType := p.args.CmdType
	if cmdType == "" {
		cmdType = p.autoDetectCommand()
		p.args.CmdType = cmdType // 保存检测结果，避免后续重复检测
		p.logger.Infof("Auto-detected command type: %s", cmdType)
	} else {
		p.logger.Infof("Using explicitly specified command type: %s", cmdType)
	}

	switch cmdType {
	case meta.CmdNew:
		return p.handleNewCommand()
	case meta.CmdUpdate:
		return p.handleUpdateCommand()
	default:
		return p.handleNewCommand()
	}
}

// parseArgs 解析插件参数
func (p *HZPlugin) parseArgs() error {
	// 从protogen插件获取参数
	param := p.gen.Request.GetParameter()

	p.args = &config.Argument{}
	if param != "" {
		params := strings.Split(param, ",")
		if err := p.args.Unpack(params); err != nil {
			return err
		}
	}

	// 从proto文件的go_package选项提取go module
	if err := p.extractGoModuleFromProto(); err != nil {
		return err
	}

	return nil
}

// extractGoModuleFromProto 从proto文件的go_package选项提取go module
func (p *HZPlugin) extractGoModuleFromProto() error {
	if len(p.gen.Files) == 0 {
		return fmt.Errorf("no proto files to generate")
	}

	for _, file := range p.gen.Files {
		if file.Generate && file.Proto != nil {
			goPackage := file.Proto.GetOptions().GetGoPackage()
			if goPackage != "" {
				// go_package格式: "github.com/example/project/biz/model"
				// 需要提取Go模块根路径
				// 方法：去掉最后的package部分，通常是去掉 /biz/xxx 这样的路径

				if p.args.Gomod == "" {
					// 尝试智能提取模块根路径
					// 如果包含 /biz/, 则提取到 /biz 之前的部分
					moduleRoot := goPackage
					if idx := strings.Index(goPackage, "/biz/"); idx != -1 {
						moduleRoot = goPackage[:idx]
					} else if idx := strings.LastIndex(goPackage, "/"); idx != -1 {
						// 如果没有 /biz/，则去掉最后一个路径段
						moduleRoot = goPackage[:idx]
					}
					p.args.Gomod = moduleRoot
					p.logger.Debugf("Extracted go module root from proto: %s (from go_package: %s)", moduleRoot, goPackage)
				}
				return nil
			}
		}
	}

	return fmt.Errorf("no go_package option found in proto files")
}

// autoDetectCommand 自动检测命令类型
func (p *HZPlugin) autoDetectCommand() string {
	// 检查项目是否已存在：检查go.mod或关键目录是否存在
	// 如果handler或router目录存在，认为项目已存在
	outDir := p.args.OutDir
	if outDir == "" || outDir == "." {
		outDir = "."
	}

	handlerPath := filepath.Join(outDir, p.args.HandlerDir)
	routerPath := filepath.Join(outDir, p.args.RouterDir)

	p.logger.Debugf("Auto-detect: checking handlerPath=%s, routerPath=%s", handlerPath, routerPath)

	_, handlerExists := os.Stat(handlerPath)
	_, routerExists := os.Stat(routerPath)

	p.logger.Debugf("Auto-detect: handlerExists=%v, routerExists=%v", handlerExists, routerExists)

	if handlerExists == nil || routerExists == nil {
		// 至少一个目录存在，认为是update
		p.logger.Debug("Auto-detect: directory exists, returning update")
		return meta.CmdUpdate
	}

	// 否则是new
	p.logger.Debug("Auto-detect: directory not exists, returning new")
	return meta.CmdNew
}

// handleNewCommand 处理new命令，生成项目布局和代码
func (p *HZPlugin) handleNewCommand() error {
	p.logger.Info("Handling new command")

	// 1. 生成项目布局
	if err := p.generateLayout(); err != nil {
		return fmt.Errorf("generate layout failed: %w", err)
	}

	// 2. 生成HTTP代码 (不生成模型代码，由protoc-gen-go负责)
	if err := p.generateHTTPCode(); err != nil {
		return fmt.Errorf("generate http code failed: %w", err)
	}

	return nil
}

// handleUpdateCommand 处理update命令，更新现有项目
func (p *HZPlugin) handleUpdateCommand() error {
	p.logger.Info("Handling update command")

	// 生成HTTP代码 (不生成模型代码，由protoc-gen-go负责)
	if err := p.generateHTTPCode(); err != nil {
		return fmt.Errorf("generate http code failed: %w", err)
	}

	return nil
}

// handleModelCommand 处理model命令，只生成模型代码
// 注意: 实际的protobuf模型代码应该由protoc-gen-go生成
// 这个命令保留是为了兼容性，但建议直接使用protoc-gen-go
func (p *HZPlugin) handleModelCommand() error {
	p.logger.Warn("Model generation should be handled by protoc-gen-go plugin")
	p.logger.Info("Please use: protoc --go_out=. --go_opt=paths=source_relative your.proto")
	return nil
}

// handleClientCommand 处理client命令，生成客户端代码
func (p *HZPlugin) handleClientCommand() error {
	p.logger.Info("Handling client command")

	// 生成客户端代码 (不生成模型代码，由protoc-gen-go负责)
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
// 已废弃: protobuf模型代码应该由protoc-gen-go生成
// 保留此函数仅为向后兼容
func (p *HZPlugin) generateModels() error {
	p.logger.Warn("generateModels is deprecated, use protoc-gen-go instead")
	return nil
}

// generateFile 生成单个protobuf文件的模型代码
// 已废弃: protobuf模型代码应该由protoc-gen-go生成
func (p *HZPlugin) generateFile(file *protogen.File) error {
	p.logger.Warn("generateFile is deprecated, use protoc-gen-go instead")
	return nil
}

// generateEnum 生成枚举
// 已废弃: protobuf模型代码应该由protoc-gen-go生成
func (p *HZPlugin) generateEnum(g *protogen.GeneratedFile, enum *protogen.Enum) {
	// 不再生成枚举代码
}

// generateMessage 生成消息
// 已废弃: protobuf模型代码应该由protoc-gen-go生成
func (p *HZPlugin) generateMessage(g *protogen.GeneratedFile, message *protogen.Message) {
	// 不再生成消息代码
}

// generateHTTPCode 生成HTTP相关代码
func (p *HZPlugin) generateHTTPCode() error {
	p.logger.Debugf("Generating HTTP code with args: %+v", p.args)

	// 使用已经确定的命令类型（在Run()中已经检测过）
	cmdType := p.args.CmdType
	if cmdType == "" {
		cmdType = p.autoDetectCommand()
	}

	// 创建HTTP包生成器
	pkgGen := &generator.HTTPPackageGenerator{
		CmdType:          cmdType,
		ProjPackage:      p.args.Gomod,
		HandlerDir:       p.args.HandlerDir,
		RouterDir:        p.args.RouterDir,
		ModelDir:         p.args.ModelDir,
		ClientDir:        p.args.ClientDir,
		BaseDomain:       p.args.BaseDomain,
		HandlerByMethod:  p.args.HandlerByMethod,
		SortRouter:       p.args.SortRouter,
		CustomizePackage: p.args.CustomizePackage,
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

		// 根据文件路径确定Go import路径
		// 例如: biz/handler/SayHello.go -> github.com/example/project/biz/handler
		goImportPath := p.buildGoImportPath(file.Path)

		g := p.gen.NewGeneratedFile(file.Path, protogen.GoImportPath(goImportPath))
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
	// 获取 model 包路径（从 proto 的 go_package 中获取）
	modelPkg := ""
	for _, file := range p.gen.Files {
		if file.Generate && file.Proto != nil {
			goPackage := file.Proto.GetOptions().GetGoPackage()
			if goPackage != "" {
				modelPkg = goPackage
				break
			}
		}
	}

	httpPkg := &generator.HTTPPackage{
		IdlName:    p.getMainIDLName(),
		Package:    p.args.Gomod,
		ModelPkg:   modelPkg, // 添加 model 包路径
		Services:   []*generator.Service{},
		Models:     []*model.Model{},
		RouterInfo: &generator.Router{},
	}

	// 解析protobuf文件，提取服务信息
	for _, file := range p.gen.Files {
		if file.Generate {
			for _, service := range file.Services {
				svc := &generator.Service{
					Name:          string(service.GoName),
					Methods:       []*generator.HTTPMethod{},
					ClientMethods: []*generator.ClientMethod{},
					Models:        []*model.Model{},
					BaseDomain:    p.args.BaseDomain,
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

// buildGoImportPath 根据文件路径构建Go import路径
func (p *HZPlugin) buildGoImportPath(filePath string) string {
	// 从文件路径提取包路径
	// 例如: biz/handler/SayHello.go -> biz/handler
	dir := filepath.Dir(filePath)

	// 拼接模块根路径
	// 例如: github.com/example/project + biz/handler -> github.com/example/project/biz/handler
	return p.args.Gomod + "/" + dir
}

// getMainIDLName 获取主IDL文件名
func (p *HZPlugin) getMainIDLName() string {
	if len(p.gen.Request.FileToGenerate) > 0 {
		return filepath.Base(p.gen.Request.FileToGenerate[0])
	}
	return "unknown.proto"
}
