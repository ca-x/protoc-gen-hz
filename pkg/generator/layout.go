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
	"text/template"
)

// LayoutGenerator 布局生成器，封装hz的LayoutGenerator
type LayoutGenerator struct {
	TemplateGenerator
}

// Layout 布局配置
type Layout struct {
	OutDir          string
	GoModule        string
	ServiceName     string
	UseApacheThrift bool
	HasIdl          bool
	NeedGoMod       bool
	ModelDir        string
	HandlerDir      string
	RouterDir       string
}

// TemplateGenerator 模板生成器
type TemplateGenerator struct {
	OutputDir string
	Excludes  []string
	tpls      map[string]*template.Template
	dirs      map[string]bool
}

// GenerateByService 根据服务信息生成布局
func (lg *LayoutGenerator) GenerateByService(service Layout) error {
	// 这里应该调用hz的布局生成逻辑
	// 为了简化，我们暂时返回nil
	return nil
}

// Persist 持久化生成的文件
func (lg *LayoutGenerator) Persist() error {
	// 这里应该调用hz的持久化逻辑
	// 为了简化，我们暂时返回nil
	return nil
}