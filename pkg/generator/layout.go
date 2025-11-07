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
	"path/filepath"
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
	files     []GeneratedFile
}

// TemplateData 模板渲染数据
type TemplateData struct {
	GoModule   string
	HandlerDir string
	RouterDir  string
	ModelDir   string
}

// GenerateByService 根据服务信息生成布局
func (lg *LayoutGenerator) GenerateByService(layout Layout) error {
	if lg.dirs == nil {
		lg.dirs = make(map[string]bool)
	}
	if lg.files == nil {
		lg.files = []GeneratedFile{}
	}

	// 准备模板数据
	data := TemplateData{
		GoModule:   layout.GoModule,
		HandlerDir: layout.HandlerDir,
		RouterDir:  layout.RouterDir,
		ModelDir:   layout.ModelDir,
	}

	// 生成所有布局文件
	for _, tpl := range DefaultLayoutTemplates {
		// 渲染文件路径（如果包含模板变量）
		filePath := tpl.Path
		if tpl.NeedRender {
			renderedPath, err := renderTemplate(tpl.Path, data)
			if err != nil {
				return fmt.Errorf("render path template %s failed: %v", tpl.Path, err)
			}
			filePath = renderedPath
		}

		// 如果是 go.mod 文件且不需要生成，则跳过
		if filepath.Base(filePath) == "go.mod" && !layout.NeedGoMod {
			continue
		}

		// 完整路径
		fullPath := filepath.Join(layout.OutDir, filePath)

		// 确保目录存在
		dir := filepath.Dir(fullPath)
		if !lg.dirs[dir] {
			lg.dirs[dir] = true
		}

		// 渲染文件内容（如果需要）
		content := tpl.Body
		if tpl.NeedRender {
			renderedContent, err := renderTemplate(tpl.Body, data)
			if err != nil {
				return fmt.Errorf("render content template %s failed: %v", tpl.Path, err)
			}
			content = renderedContent
		}

		// 添加到文件列表
		lg.files = append(lg.files, GeneratedFile{
			Path:    filePath,
			Content: content,
		})
	}

	return nil
}

// Persist 持久化生成的文件
func (lg *LayoutGenerator) Persist() error {
	// 创建所有需要的目录
	for dir := range lg.dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s failed: %v", dir, err)
		}
	}

	// 写入所有文件
	for _, file := range lg.files {
		fullPath := filepath.Join(lg.OutputDir, file.Path)

		// 确保父目录存在
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s failed: %v", dir, err)
		}

		// 写入文件
		if err := os.WriteFile(fullPath, []byte(file.Content), 0o644); err != nil {
			return fmt.Errorf("write file %s failed: %v", fullPath, err)
		}
	}

	return nil
}

// renderTemplate 渲染模板字符串
func renderTemplate(tplStr string, data interface{}) (string, error) {
	tpl, err := template.New("").Parse(tplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
