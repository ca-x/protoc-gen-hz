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

package generator

import (
	"fmt"
	"strings"

	"github.com/ca-x/protoc-gen-go-hz/pkg/version"
	"github.com/cloudwego/hertz/cmd/hz/generator/model"
)

// HTTPPackageGenerator HTTP包生成器，封装hz的HttpPackageGenerator
type HTTPPackageGenerator struct {
	ConfigPath       string
	CmdType          string
	ProjPackage      string
	HandlerDir       string
	RouterDir        string
	ModelDir         string
	UseDir           string
	ClientDir        string
	IdlClientDir     string
	ForceClientDir   string
	BaseDomain       string
	QueryEnumAsInt   bool
	ServiceGenDir    string
	CustomizePackage string // 自定义包模板路径

	NeedModel            bool
	HandlerByMethod      bool
	SnakeStyleMiddleware bool
	SortRouter           bool
	ForceUpdateClient    bool

	TemplateGenerator
	customTemplates *CustomTemplateConfig // 自定义模板配置
}

// HTTPPackage HTTP包数据结构
type HTTPPackage struct {
	IdlName    string
	Package    string
	ModelPkg   string // Model 包的完整路径（从 proto go_package 获取）
	Services   []*Service
	Models     []*model.Model
	RouterInfo *Router
}

// Service 服务结构
type Service struct {
	Name          string
	Methods       []*HTTPMethod
	ClientMethods []*ClientMethod
	Models        []*model.Model
	BaseDomain    string
	ServiceGroup  string
	ServiceGenDir string
}

// HTTPMethod HTTP方法结构
type HTTPMethod struct {
	Name         string
	HTTPMethod   string
	Path         string
	RequestType  string
	ResponseType string
}

// ClientMethod 客户端方法结构
type ClientMethod struct {
	Name         string
	HTTPMethod   string
	Path         string
	RequestType  string
	ResponseType string
}

// Router 路由信息
type Router struct {
	Registers []string
}

// Init 初始化生成器
func (pkgGen *HTTPPackageGenerator) Init() error {
	// 加载自定义模板配置（如果指定）
	if pkgGen.CustomizePackage != "" {
		config, err := LoadCustomTemplate(pkgGen.CustomizePackage)
		if err != nil {
			return fmt.Errorf("load custom template failed: %v", err)
		}
		pkgGen.customTemplates = config
	}
	return nil
}

// Generate 生成HTTP代码
func (pkgGen *HTTPPackageGenerator) Generate(httpPkg *HTTPPackage) ([]*GeneratedFile, error) {
	var files []*GeneratedFile

	// 生成handler代码
	handlerFiles, err := pkgGen.generateHandlers(httpPkg)
	if err != nil {
		return nil, err
	}
	files = append(files, handlerFiles...)

	// 生成router代码
	routerFiles, err := pkgGen.generateRouters(httpPkg)
	if err != nil {
		return nil, err
	}
	files = append(files, routerFiles...)

	// 生成client代码
	clientFiles, err := pkgGen.generateClients(httpPkg)
	if err != nil {
		return nil, err
	}
	files = append(files, clientFiles...)

	return files, nil
}

// generateHandlers 生成handler代码
func (pkgGen *HTTPPackageGenerator) generateHandlers(httpPkg *HTTPPackage) ([]*GeneratedFile, error) {
	var files []*GeneratedFile

	for _, service := range httpPkg.Services {
		for _, method := range service.Methods {
			// 使用相对路径，符合protoc插件标准
			path := pkgGen.HandlerDir + "/" + method.Name + ".go"
			file := &GeneratedFile{
				Path:    path,
				Content: pkgGen.generateHandlerCode(httpPkg, service, method),
			}
			files = append(files, file)
		}
	}

	return files, nil
}

// generateRouters 生成router代码
func (pkgGen *HTTPPackageGenerator) generateRouters(httpPkg *HTTPPackage) ([]*GeneratedFile, error) {
	var files []*GeneratedFile

	// 检查是否有自定义模板覆盖 router.go
	if pkgGen.customTemplates != nil {
		for _, tpl := range pkgGen.customTemplates.Layouts {
			if tpl.Path == "router.go" && !tpl.Disable {
				// 使用自定义模板
				return pkgGen.generateRoutersWithCustomTemplate(httpPkg, &tpl)
			}
		}
	}

	// 使用默认模板
	path := pkgGen.RouterDir + "/router.go"
	file := &GeneratedFile{
		Path:    path,
		Content: pkgGen.generateRouterCode(httpPkg),
	}
	files = append(files, file)

	return files, nil
}

// generateRoutersWithCustomTemplate 使用自定义模板生成router代码
func (pkgGen *HTTPPackageGenerator) generateRoutersWithCustomTemplate(httpPkg *HTTPPackage, tpl *CustomTemplate) ([]*GeneratedFile, error) {
	var files []*GeneratedFile

	// 准备模板数据
	data := map[string]interface{}{
		"PackageName": "router",
		"HandlerPackages": map[string]string{
			"handler": pkgGen.ProjPackage + "/" + pkgGen.HandlerDir,
		},
		"Router": httpPkg.RouterInfo,
	}

	// 渲染模板
	content, err := RenderCustomTemplate(tpl, data)
	if err != nil {
		return nil, fmt.Errorf("render custom router template failed: %v", err)
	}

	path := pkgGen.RouterDir + "/router.go"
	file := &GeneratedFile{
		Path:    path,
		Content: content,
	}
	files = append(files, file)

	return files, nil
}

// generateClients 生成client代码
func (pkgGen *HTTPPackageGenerator) generateClients(httpPkg *HTTPPackage) ([]*GeneratedFile, error) {
	var files []*GeneratedFile

	// 如果没有指定client目录，则不生成客户端代码
	if pkgGen.ClientDir == "" {
		return files, nil
	}

	for _, service := range httpPkg.Services {
		// 使用相对路径，符合protoc插件标准
		path := pkgGen.ClientDir + "/" + service.Name + "_client.go"
		file := &GeneratedFile{
			Path:    path,
			Content: pkgGen.generateClientCode(service),
		}
		files = append(files, file)
	}

	return files, nil
}

// generateHandlerCode 生成单个handler的代码
func (pkgGen *HTTPPackageGenerator) generateHandlerCode(httpPkg *HTTPPackage, service *Service, method *HTTPMethod) string {
	// 使用 ModelPkg 如果有，否则回退到默认路径
	modelImport := httpPkg.ModelPkg
	if modelImport == "" {
		modelImport = pkgGen.ProjPackage + "/biz/model"
	}

	// 从导入路径中提取包名（最后一个路径段）
	modelPkgName := "model"
	if idx := strings.LastIndex(modelImport, "/"); idx != -1 {
		modelPkgName = modelImport[idx+1:]
	}

	return `// Code generated by protoc-gen-go-hz ` + version.Version + `. DO NOT EDIT.

package handler

import (
    "context"
    "github.com/cloudwego/hertz/pkg/app"
    "` + modelImport + `"
)

// ` + method.Name + ` .
func ` + method.Name + `(ctx context.Context, c *app.RequestContext) {
    var err error
    var req ` + modelPkgName + `.` + method.RequestType + `
    err = c.BindAndValidate(&req)
    if err != nil {
        c.JSON(400, map[string]interface{}{
            "error": err.Error(),
        })
        return
    }

    // TODO: implement your business logic here
    resp := &` + modelPkgName + `.` + method.ResponseType + `{}
    c.JSON(200, resp)
}
`
}

// generateRouterCode 生成router代码
func (pkgGen *HTTPPackageGenerator) generateRouterCode(httpPkg *HTTPPackage) string {
	code := `// Code generated by protoc-gen-go-hz ` + version.Version + `. DO NOT EDIT.

package router

import (
    "github.com/cloudwego/hertz/pkg/app/server"
    "` + pkgGen.ProjPackage + `/biz/handler"
)

// Register registers HTTP handlers.
func Register(r *server.Hertz) {
`

	for _, service := range httpPkg.Services {
		for _, method := range service.Methods {
			code += `    r.` + method.HTTPMethod + `("` + method.Path + `", handler.` + method.Name + `)
`
		}
	}

	code += `}
`
	return code
}

// generateClientCode 生成client代码
func (pkgGen *HTTPPackageGenerator) generateClientCode(service *Service) string {
	code := `// Code generated by protoc-gen-go-hz ` + version.Version + `. DO NOT EDIT.

package client

import (
    "context"
    "github.com/cloudwego/hertz/pkg/app/client"
    "github.com/cloudwego/hertz/pkg/protocol"
    "` + pkgGen.ProjPackage + `/biz/model"
)

// ` + service.Name + `Client .
type ` + service.Name + `Client struct {
    client *client.Client
}

// New` + service.Name + `Client creates a new ` + service.Name + `Client.
func New` + service.Name + `Client(client *client.Client) *` + service.Name + `Client {
    return &` + service.Name + `Client{
        client: client,
    }
}

`

	for _, method := range service.Methods {
		code += `// ` + method.Name + ` calls ` + method.Name + ` endpoint.
func (c *` + service.Name + `Client) ` + method.Name + `(ctx context.Context, req *model.` + method.RequestType + `) (*model.` + method.ResponseType + `, error) {
    var resp model.` + method.ResponseType + `
    err := c.client.` + method.HTTPMethod + `(ctx, nil, "` + method.Path + `", req, &resp)
    return &resp, err
}

`
	}

	return code
}

// GeneratedFile 生成的文件
type GeneratedFile struct {
	Path    string
	Content string
}
