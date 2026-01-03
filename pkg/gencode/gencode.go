package gencode

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// Config 代码生成器配置
type Config struct {
	ProjectName   string        `json:"project_name"`
	GenConfig     GenConfig     `json:"gen_config"`
	PackageConfig PackageConfig `json:"package_config"`
}

// GenConfig 代码生成配置
type GenConfig struct {
	OutputPath    string `json:"output_path"`
	EnableLombok  bool   `json:"enable_lombok"`
	EnableSwagger bool   `json:"enable_swagger"`
	Author        string `json:"author"`
	Date          string `json:"date"`
}

// PackageConfig 包名配置
type PackageConfig struct {
	BasePackage       string `json:"base_package"`
	EntityPackage     string `json:"entity_package"`
	MapperPackage     string `json:"mapper_package"`
	ServicePackage    string `json:"service_package"`
	ControllerPackage string `json:"controller_package"`
}

// Table 表信息
type Table struct {
	TableName    string
	TableComment string
	Fields       []Field
	PrimaryKey   Field
}

// Field 字段信息
type Field struct {
	ColumnName    string
	ColumnType    string
	ColumnComment string
	IsNullable    bool
	IsPrimaryKey  bool
	GoType        string
	JavaType      string
	FieldName     string
}

// Generator 代码生成器
type Generator struct {
	Config       Config
	Tables       []Table
	TemplatePath string
}

// TemplateInfo 模板信息
type TemplateInfo struct {
	FilePath   string // 模板文件路径
	OutputPath string // 输出路径（从元数据解析或根据模板路径生成）
	IsPerTable bool   // 是否需要为每个表生成（包含表相关变量）
}

// NewGenerator 创建代码生成器实例
func NewGenerator(config Config, tables []Table) *Generator {
	if config.GenConfig.Date == "" {
		config.GenConfig.Date = time.Now().Format("2006-01-02")
	}

	// 设置模板路径
	currentDir, _ := os.Getwd()
	// 如果当前目录以pkg/gencode结尾，说明是在测试环境，需要回到项目根目录
	if strings.HasSuffix(currentDir, "pkg/gencode") {
		currentDir = filepath.Join(currentDir, "../..")
	}

	return &Generator{
		Config:       config,
		Tables:       tables,
		TemplatePath: currentDir,
	}
}

// Init 初始化代码生成器
func (g *Generator) Init() error {
	// 确保输出目录存在
	err := g.EnsureOutputDirs()
	if err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	return nil
}

// EnsureOutputDirs 确保输出目录存在
func (g *Generator) EnsureOutputDirs() error {
	outputDir := g.Config.GenConfig.OutputPath
	if outputDir == "" {
		outputDir = "./output"
	}

	// 创建基础输出目录
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}

	return nil
}

// GenerateCode 生成代码
func (g *Generator) GenerateCode() error {
	// 扫描所有模板文件
	templates, err := g.scanTemplates()
	if err != nil {
		return fmt.Errorf("扫描模板文件失败: %v", err)
	}

	// 生成代码
	for _, tmplInfo := range templates {
		if tmplInfo.IsPerTable {
			// 需要为每个表生成
			for _, table := range g.Tables {
				templateData := g.prepareTemplateData(&table)
				err := g.generateFromTemplate(tmplInfo, templateData)
				if err != nil {
					return fmt.Errorf("生成文件失败 [%s]: %v", tmplInfo.FilePath, err)
				}
			}
		} else {
			// 只生成一次（如pom.xml, Application.java等）
			templateData := TemplateData{
				Config: g.Config,
			}
			err := g.generateFromTemplate(tmplInfo, templateData)
			if err != nil {
				return fmt.Errorf("生成文件失败 [%s]: %v", tmplInfo.FilePath, err)
			}
		}
	}

	return nil
}

// scanTemplates 扫描所有模板文件
func (g *Generator) scanTemplates() ([]TemplateInfo, error) {
	var templates []TemplateInfo
	templateDir := filepath.Join(g.TemplatePath, "pkg/gencode/template")

	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".tpl") {
			tmplInfo, err := g.parseTemplateInfo(path)
			if err != nil {
				return fmt.Errorf("解析模板信息失败 [%s]: %v", path, err)
			}
			templates = append(templates, tmplInfo)
		}

		return nil
	})

	return templates, err
}

// parseTemplateInfo 解析模板信息
func (g *Generator) parseTemplateInfo(templatePath string) (TemplateInfo, error) {
	file, err := os.Open(templatePath)
	if err != nil {
		return TemplateInfo{}, err
	}
	defer file.Close()

	var outputPath string
	var isPerTable bool

	// 读取文件前几行查找元数据
	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() && lineCount < 10 { // 只检查前10行
		line := strings.TrimSpace(scanner.Text())

		// 解析 @@Meta.Output 元数据
		if strings.HasPrefix(line, "@@Meta.Output=") {
			outputPath = strings.Trim(strings.TrimPrefix(line, "@@Meta.Output="), "\"")
			break
		}
		lineCount++
	}

	// 如果没有找到元数据，根据模板文件路径生成输出路径
	if outputPath == "" {
		outputPath = g.generateOutputPathFromTemplate(templatePath)
	}

	// 检查是否包含表相关的模板变量，判断是否需要为每个表生成
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return TemplateInfo{}, err
	}

	contentStr := string(content)
	// 检查是否包含表相关变量
	tableVarPattern := regexp.MustCompile(`\{\{\.(?:Table|ClassName)\b`)
	isPerTable = tableVarPattern.MatchString(contentStr)

	return TemplateInfo{
		FilePath:   templatePath,
		OutputPath: outputPath,
		IsPerTable: isPerTable,
	}, nil
}

// generateOutputPathFromTemplate 根据模板文件路径生成输出路径
func (g *Generator) generateOutputPathFromTemplate(templatePath string) string {
	// 获取相对于template目录的路径
	templateDir := filepath.Join(g.TemplatePath, "pkg/gencode/template")
	relPath, err := filepath.Rel(templateDir, templatePath)
	if err != nil {
		return ""
	}

	// 移除.tpl后缀
	outputPath := strings.TrimSuffix(relPath, ".tpl")

	// 如果是java目录下的文件，需要添加前缀斜杠
	if strings.HasPrefix(outputPath, "java/") {
		outputPath = "/" + outputPath
	}

	return outputPath
}

// generateFromTemplate 根据模板信息生成文件
func (g *Generator) generateFromTemplate(tmplInfo TemplateInfo, data TemplateData) error {
	// 创建模板
	tmpl, err := g.createTemplateWithFuncs(tmplInfo.FilePath)
	if err != nil {
		return fmt.Errorf("加载模板失败: %v", err)
	}

	// 渲染输出路径
	outputPath, err := g.renderOutputPath(tmplInfo.OutputPath, data)
	if err != nil {
		return fmt.Errorf("渲染输出路径失败: %v", err)
	}

	// 生成完整的输出路径
	baseOutputPath := g.Config.GenConfig.OutputPath
	if baseOutputPath == "" {
		baseOutputPath = "./output"
	}

	fullOutputPath := filepath.Join(baseOutputPath, outputPath)

	// 确保输出目录存在
	outputDir := filepath.Dir(fullOutputPath)
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 生成文件
	return g.generateFile(fullOutputPath, tmpl, data)
}

// renderOutputPath 渲染输出路径模板
func (g *Generator) renderOutputPath(pathTemplate string, data TemplateData) (string, error) {
	tmpl, err := template.New("outputPath").Funcs(g.getTemplateFuncMap()).Parse(pathTemplate)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	err = tmpl.Execute(&result, data)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

// TemplateData 模板数据
type TemplateData struct {
	Config            Config
	Table             Table
	ClassName         string
	EntityPackage     string
	MapperPackage     string
	ServicePackage    string
	ControllerPackage string
	EnableLombok      bool
	EnableSwagger     bool
	Author            string
	Date              string
}

// prepareTemplateData 准备模板数据
func (g *Generator) prepareTemplateData(table *Table) TemplateData {
	pkgConfig := g.Config.PackageConfig
	genConfig := g.Config.GenConfig

	// 生成类名（去掉前缀，如果有的话）
	className := table.TableName
	className = g.toPascalCase(className)

	return TemplateData{
		Config:            g.Config,
		Table:             *table,
		ClassName:         className,
		EntityPackage:     pkgConfig.EntityPackage,
		MapperPackage:     pkgConfig.MapperPackage,
		ServicePackage:    pkgConfig.ServicePackage,
		ControllerPackage: pkgConfig.ControllerPackage,
		EnableLombok:      genConfig.EnableLombok,
		EnableSwagger:     genConfig.EnableSwagger,
		Author:            genConfig.Author,
		Date:              genConfig.Date,
	}
}

// getTemplateFuncMap 获取模板自定义函数映射
func (g *Generator) getTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
		"add": func(a, b int) int {
			return a + b
		},
		"replace": func(old, new, s string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
		},
	}
}

// createTemplateWithFuncs 创建带有自定义函数的模板
func (g *Generator) createTemplateWithFuncs(templatePath string) (*template.Template, error) {
	// 读取模板文件内容
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, err
	}

	// 清理元数据行
	cleanedContent := g.cleanMetadataFromTemplate(string(content))

	// 创建模板并添加自定义函数
	tmpl := template.New(filepath.Base(templatePath)).Funcs(g.getTemplateFuncMap())
	tmpl, err = tmpl.Parse(cleanedContent)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// cleanMetadataFromTemplate 从模板内容中清理元数据行
func (g *Generator) cleanMetadataFromTemplate(content string) string {
	lines := strings.Split(content, "\n")
	var cleanedLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		// 跳过元数据行
		if strings.HasPrefix(trimmedLine, "@@Meta.") {
			continue
		}
		cleanedLines = append(cleanedLines, line)
	}

	// 移除开头的空行
	for len(cleanedLines) > 0 && strings.TrimSpace(cleanedLines[0]) == "" {
		cleanedLines = cleanedLines[1:]
	}

	return strings.Join(cleanedLines, "\n")
}

// generateFile 生成文件
func (g *Generator) generateFile(filePath string, tmpl *template.Template, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("模板渲染失败: %v", err)
	}

	return nil
}

// Close 关闭资源（现在无需关闭数据库连接）
func (g *Generator) Close() error {
	return nil
}

// toCamelCase 下划线转驼峰命名
func (g *Generator) toCamelCase(str string) string {
	parts := strings.Split(str, "_")
	if len(parts) == 0 {
		return str
	}

	var result strings.Builder
	result.WriteString(parts[0])
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result.WriteString(strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:]))
		}
	}
	return result.String()
}

// toPascalCase 下划线转大驼峰命名
func (g *Generator) toPascalCase(str string) string {
	parts := strings.Split(str, "_")
	var result strings.Builder
	for i := 0; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result.WriteString(strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:]))
		}
	}
	return result.String()
}
