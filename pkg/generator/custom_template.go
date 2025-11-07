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
	"bytes"
	"fmt"
	"os"
	"text/template"

	"gopkg.in/yaml.v2"
)

// CustomTemplateConfig 自定义模板配置
type CustomTemplateConfig struct {
	Layouts []CustomTemplate `yaml:"layouts"`
}

// CustomTemplate 自定义模板定义
type CustomTemplate struct {
	Path           string         `yaml:"path"`            // 文件路径
	Delims         [2]string      `yaml:"delims"`          // 模板分隔符，默认 {{}}
	Body           string         `yaml:"body"`            // 模板内容
	Disable        bool           `yaml:"disable"`         // 禁用生成
	LoopMethod     bool           `yaml:"loop_method"`     // 按方法循环生成
	LoopService    bool           `yaml:"loop_service"`    // 按服务循环生成
	UpdateBehavior UpdateBehavior `yaml:"update_behavior"` // 更新行为
}

// UpdateBehavior 更新行为配置
type UpdateBehavior struct {
	Type           string   `yaml:"type"`            // skip/cover/append
	AppendKey      string   `yaml:"append_key"`      // 追加的键：method/service
	InsertKey      string   `yaml:"insert_key"`      // 插入的键模板
	AppendTpl      string   `yaml:"append_tpl"`      // 追加内容模板
	ImportTpl      []string `yaml:"import_tpl"`      // 导入模板
	AppendLocation string   `yaml:"append_location"` // 追加位置
}

// LoadCustomTemplate 加载自定义模板配置
func LoadCustomTemplate(path string) (*CustomTemplateConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template config file %s failed: %v", path, err)
	}

	var config CustomTemplateConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal template config failed: %v", err)
	}

	// 设置默认分隔符
	for i := range config.Layouts {
		if config.Layouts[i].Delims[0] == "" {
			config.Layouts[i].Delims = [2]string{"{{", "}}"}
		}
	}

	return &config, nil
}

// RenderCustomTemplate 渲染自定义模板
func RenderCustomTemplate(tpl *CustomTemplate, data interface{}) (string, error) {
	// 创建模板并设置分隔符
	t := template.New(tpl.Path)
	t.Delims(tpl.Delims[0], tpl.Delims[1])

	// 解析模板
	t, err := t.Parse(tpl.Body)
	if err != nil {
		return "", fmt.Errorf("parse template %s failed: %v", tpl.Path, err)
	}

	// 渲染模板
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %s failed: %v", tpl.Path, err)
	}

	return buf.String(), nil
}
