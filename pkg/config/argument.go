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

package config

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/cloudwego/hertz/cmd/hz/meta"
)

// Argument 是插件参数配置，基于hz的Argument但简化
type Argument struct {
    // 基本命令参数
    CmdType    string   // 命令类型: new, update, model, client
    Verbose    bool     // 详细输出
    OutDir     string   // 输出目录
    HandlerDir string   // handler目录
    ModelDir   string   // model目录
    RouterDir  string   // router目录
    ClientDir  string   // client目录
    BaseDomain string   // 请求域名

    // Go模块相关
    Gomod       string // Go模块名
    ServiceName string // 服务名
    Use         string // 使用第三方模型包
    NeedGoMod   bool   // 是否需要生成go.mod

    // IDL相关
    IdlType       string            // IDL类型，这里固定为proto
    OptPkgMap     map[string]string // 包映射
    TrimGoPackage string            // 修剪go_package前缀

    // 代码生成选项
    JSONEnumStr          bool     // JSON枚举使用字符串
    QueryEnumAsInt       bool     // 查询参数枚举使用整数
    UnsetOmitempty       bool     // 移除omitempty标签
    ProtobufCamelJSONTag bool     // protobuf JSON标签使用驼峰命名
    SnakeName            bool     // 标签使用蛇形命名
    RmTags               []string // 要移除的标签
    Excludes             []string // 排除的文件
    NoRecurse            bool     // 不递归生成
    HandlerByMethod      bool     // 按方法生成handler文件
    SortRouter           bool     // 排序路由代码
    ForceUpdateClient    bool     // 强制更新客户端代码

    // 自定义选项
    CustomizeLayout  string // 自定义布局模板路径
    CustomizePackage string // 自定义包模板路径
}

// Unpack 解析参数列表
func (arg *Argument) Unpack(params []string) error {
    // 初始化默认值
    if arg.OptPkgMap == nil {
        arg.OptPkgMap = make(map[string]string)
    }
    if arg.Excludes == nil {
        arg.Excludes = []string{}
    }
    if arg.RmTags == nil {
        arg.RmTags = []string{}
    }

    // 设置默认值
    arg.IdlType = meta.IdlProto
    if arg.OutDir == "" {
        arg.OutDir = "."
    }
    if arg.ModelDir == "" {
        arg.ModelDir = meta.ModelDir
    }
    if arg.HandlerDir == "" {
        arg.HandlerDir = meta.HandlerDir
    }
    if arg.RouterDir == "" {
        arg.RouterDir = meta.RouterDir
    }
    if arg.ServiceName == "" {
        arg.ServiceName = meta.DefaultServiceName
    }

    // 解析参数
    for _, param := range params {
        if err := arg.parseParam(param); err != nil {
            return err
        }
    }

    return nil
}

// parseParam 解析单个参数
func (arg *Argument) parseParam(param string) error {
    if param == "" {
        return nil
    }

    // 查找等号分割的键值对
    parts := strings.SplitN(param, "=", 2)
    if len(parts) != 2 {
        return fmt.Errorf("invalid parameter format: %s", param)
    }

    key := strings.TrimSpace(parts[0])
    value := strings.TrimSpace(parts[1])

    switch key {
    case "command", "cmd":
        arg.CmdType = value
    case "verbose":
        arg.Verbose = value == "true" || value == "1"
    case "out_dir":
        arg.OutDir = value
    case "handler_dir":
        arg.HandlerDir = value
    case "model_dir":
        arg.ModelDir = value
    case "router_dir":
        arg.RouterDir = value
    case "client_dir":
        arg.ClientDir = value
    case "base_domain":
        arg.BaseDomain = value
    case "module", "go_module":
        arg.Gomod = value
    case "service":
        arg.ServiceName = value
    case "use":
        arg.Use = value
    case "need_go_mod":
        arg.NeedGoMod = value == "true" || value == "1"
    case "model":
        // model=true表示生成模型代码
        if value == "false" || value == "0" {
            // 不生成模型代码的标志
        }
    case "json_enumstr":
        arg.JSONEnumStr = value == "true" || value == "1"
    case "query_enumint":
        arg.QueryEnumAsInt = value == "true" || value == "1"
    case "unset_omitempty":
        arg.UnsetOmitempty = value == "true" || value == "1"
    case "pb_camel_json_tag":
        arg.ProtobufCamelJSONTag = value == "true" || value == "1"
    case "snake_tag":
        arg.SnakeName = value == "true" || value == "1"
    case "no_recurse":
        arg.NoRecurse = value == "true" || value == "1"
    case "handler_by_method":
        arg.HandlerByMethod = value == "true" || value == "1"
    case "sort_router":
        arg.SortRouter = value == "true" || value == "1"
    case "force_client":
        arg.ForceUpdateClient = value == "true" || value == "1"
    case "exclude_file":
        arg.Excludes = append(arg.Excludes, strings.Split(value, ",")...)
    case "rm_tag":
        arg.RmTags = append(arg.RmTags, strings.Split(value, ",")...)
    case "customize_layout":
        arg.CustomizeLayout = value
    case "customize_package":
        arg.CustomizePackage = value
    case "trim_gopackage":
        arg.TrimGoPackage = value
    default:
        if strings.HasPrefix(key, "option_package:") {
            // 解析option_package参数
            pkgParam := strings.TrimPrefix(key, "option_package:")
            if err := arg.parseOptionPackage(pkgParam, value); err != nil {
                return err
            }
        } else {
            // 未知的参数，忽略或返回错误
            return fmt.Errorf("unknown parameter: %s", key)
        }
    }

    return nil
}

// parseOptionPackage 解析option_package参数
func (arg *Argument) parseOptionPackage(key, value string) error {
    // 格式: include_path=import_path
    if value == "" {
        return fmt.Errorf("option_package value cannot be empty")
    }
    arg.OptPkgMap[key] = value
    return nil
}

// GetGoPackage 获取Go包名
func (arg *Argument) GetGoPackage() (string, error) {
    if arg.Gomod == "" {
        return "", fmt.Errorf("go module is required")
    }
    return arg.Gomod, nil
}

// GetModelDir 获取模型目录
func (arg *Argument) GetModelDir() (string, error) {
    if arg.ModelDir == "" {
        return meta.ModelDir, nil
    }
    return arg.ModelDir, nil
}

// Validate 验证参数
func (arg *Argument) Validate() error {
    if arg.Gomod == "" && arg.CmdType == meta.CmdNew {
        return fmt.Errorf("go module is required for new command")
    }
    
    if arg.OutDir == "" {
        arg.OutDir = "."
    }

    // 验证目录路径
    if !filepath.IsAbs(arg.OutDir) {
        arg.OutDir = filepath.Join(".", arg.OutDir)
    }

    return nil
}