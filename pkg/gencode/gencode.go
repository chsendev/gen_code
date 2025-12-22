package gencode

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// Config 代码生成器配置
type Config struct {
	DBConfig      DBConfig      `json:"db_config"`
	GenConfig     GenConfig     `json:"gen_config"`
	PackageConfig PackageConfig `json:"package_config"`
}

// DBConfig 数据库配置
type DBConfig struct {
	DriverName    string   `json:"driver_name"`
	Host          string   `json:"host"`
	Port          int      `json:"port"`
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	DatabaseName  string   `json:"database_name"`
	TablePrefix   string   `json:"table_prefix"`
	IncludeTables []string `json:"include_tables"`
	ExcludeTables []string `json:"exclude_tables"`
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
	DB           *sql.DB
	TemplatePath string
}

// 模板文件路径常量
const (
	EntityTemplateFile      = "template/java/src/main/java/entity.java.tpl"
	MapperTemplateFile      = "template/java/src/main/java/mapper.java.tpl"
	MapperXmlTemplateFile   = "template/java/src/main/resources/mapper/mapper.xml.tpl"
	ServiceTemplateFile     = "template/java/src/main/java/service.java.tpl"
	ServiceImplTemplateFile = "template/java/src/main/java/serviceImpl.java.tpl"
	ControllerTemplateFile  = "template/java/src/main/java/controller.java.tpl"
)

// NewGenerator 创建代码生成器实例
func NewGenerator(config Config) *Generator {
	if config.GenConfig.Date == "" {
		config.GenConfig.Date = time.Now().Format("2006-01-02")
	}
	
	// 设置模板路径为当前包下的template目录
	// 获取当前文件所在目录
	currentDir, _ := os.Getwd()
	return &Generator{
		Config:       config,
		TemplatePath: currentDir,
	}
}

// Init 初始化代码生成器
func (g *Generator) Init() error {
	// 连接数据库
	dsn := g.buildDSN()
	var err error
	g.DB, err = sql.Open(g.Config.DBConfig.DriverName, dsn)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	// 测试连接
	err = g.DB.Ping()
	if err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	// 确保输出目录存在
	err = g.EnsureOutputDirs()
	if err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	return nil
}

// buildDSN 构建数据库连接字符串
func (g *Generator) buildDSN() string {
	dbConfig := g.Config.DBConfig
	switch dbConfig.DriverName {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DatabaseName)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Password, dbConfig.DatabaseName)
	default:
		return ""
	}
}

// EnsureOutputDirs 确保输出目录存在
func (g *Generator) EnsureOutputDirs() error {
	outputDir := g.Config.GenConfig.OutputPath
	if outputDir == "" {
		outputDir = "./output"
	}

	// 创建包目录结构
	pkgConfig := g.Config.PackageConfig
	dirs := []string{
		filepath.Join(outputDir, strings.ReplaceAll(pkgConfig.EntityPackage, ".", "/")),
		filepath.Join(outputDir, strings.ReplaceAll(pkgConfig.MapperPackage, ".", "/")),
		filepath.Join(outputDir, strings.ReplaceAll(pkgConfig.ServicePackage, ".", "/")),
		filepath.Join(outputDir, strings.ReplaceAll(pkgConfig.ControllerPackage, ".", "/")),
		filepath.Join(outputDir, "src/main/resources/mapper"),
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTables 获取所有表信息
func (g *Generator) GetTables() ([]string, error) {
	dbConfig := g.Config.DBConfig
	var tables []string

	var query string
	if dbConfig.DriverName == "mysql" {
		query = `SELECT table_name FROM information_schema.tables WHERE table_schema = ?`
	} else if dbConfig.DriverName == "postgres" {
		query = `SELECT table_name FROM information_schema.tables WHERE table_catalog = ?`
	} else {
		return nil, fmt.Errorf("不支持的数据库驱动: %s", dbConfig.DriverName)
	}

	rows, err := g.DB.Query(query, dbConfig.DatabaseName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}

		// 过滤表
		if !g.shouldIncludeTable(tableName) {
			continue
		}

		tables = append(tables, tableName)
	}

	return tables, nil
}

// shouldIncludeTable 判断是否应该包含该表
func (g *Generator) shouldIncludeTable(tableName string) bool {
	dbConfig := g.Config.DBConfig

	// 检查排除列表
	for _, excludeTable := range dbConfig.ExcludeTables {
		if tableName == excludeTable {
			return false
		}
	}

	// 如果包含列表为空，则包含所有表
	if len(dbConfig.IncludeTables) == 0 {
		return true
	}

	// 检查包含列表
	for _, includeTable := range dbConfig.IncludeTables {
		if tableName == includeTable {
			return true
		}
	}

	return false
}

// GetTableMetaData 获取表元数据
func (g *Generator) GetTableMetaData(tableName string) (*Table, error) {
	table, err := g.getTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	fields, err := g.getTableFields(tableName)
	if err != nil {
		return nil, err
	}

	table.Fields = fields

	// 设置主键
	for _, field := range fields {
		if field.IsPrimaryKey {
			table.PrimaryKey = field
			break
		}
	}

	return table, nil
}

// getTableInfo 获取表基本信息
func (g *Generator) getTableInfo(tableName string) (*Table, error) {
	dbConfig := g.Config.DBConfig
	var query string
	var tableComment string

	if dbConfig.DriverName == "mysql" {
		query = `SELECT table_comment FROM information_schema.tables WHERE table_schema = ? AND table_name = ?`
	} else if dbConfig.DriverName == "postgres" {
		query = `SELECT obj_description(oid) FROM pg_class WHERE relname = ? AND relkind = 'r'`
	} else {
		return nil, fmt.Errorf("不支持的数据库驱动: %s", dbConfig.DriverName)
	}

	if dbConfig.DriverName == "mysql" {
		err := g.DB.QueryRow(query, dbConfig.DatabaseName, tableName).Scan(&tableComment)
		if err != nil {
			return nil, err
		}
	} else {
		err := g.DB.QueryRow(query, tableName).Scan(&tableComment)
		if err != nil {
			// PostgreSQL如果没有注释会返回空
			tableComment = ""
		}
	}

	return &Table{
		TableName:    tableName,
		TableComment: tableComment,
	}, nil
}

// getTableFields 获取表字段信息
func (g *Generator) getTableFields(tableName string) ([]Field, error) {
	dbConfig := g.Config.DBConfig
	var query string
	var fields []Field

	if dbConfig.DriverName == "mysql" {
		query = `
			SELECT 
				column_name, 
				column_type, 
				column_comment, 
				is_nullable, 
				column_key 
			FROM information_schema.columns 
			WHERE table_schema = ? AND table_name = ? 
			ORDER BY ordinal_position
		`
	} else if dbConfig.DriverName == "postgres" {
		query = `
			SELECT 
				column_name, 
				data_type, 
				column_description(c.oid, a.attnum), 
				a.attnotnull, 
				CASE WHEN i.indisprimary THEN 'PRI' ELSE '' END 
			FROM pg_class c 
			JOIN pg_attribute a ON a.attrelid = c.oid 
			LEFT JOIN pg_index i ON i.indrelid = c.oid AND a.attnum = ANY(i.indkey) 
			WHERE c.relname = ? AND a.attnum > 0 
			ORDER BY a.attnum
		`
	} else {
		return nil, fmt.Errorf("不支持的数据库驱动: %s", dbConfig.DriverName)
	}

	var rows *sql.Rows
	var err error

	if dbConfig.DriverName == "mysql" {
		rows, err = g.DB.Query(query, dbConfig.DatabaseName, tableName)
	} else {
		rows, err = g.DB.Query(query, tableName)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName, columnType, columnComment, isNullable, columnKey string

		if dbConfig.DriverName == "mysql" {
			err := rows.Scan(&columnName, &columnType, &columnComment, &isNullable, &columnKey)
			if err != nil {
				return nil, err
			}
		} else {
			var notNull bool
			err := rows.Scan(&columnName, &columnType, &columnComment, &notNull, &columnKey)
			if err != nil {
				return nil, err
			}
			isNullable = "YES"
			if notNull {
				isNullable = "NO"
			}
		}

		field := Field{
			ColumnName:    columnName,
			ColumnType:    columnType,
			ColumnComment: columnComment,
			IsNullable:    isNullable == "YES",
			IsPrimaryKey:  columnKey == "PRI",
		}

		// 设置Java类型
		field.JavaType = g.getJavaType(columnType, dbConfig.DriverName)

		// 设置字段名（驼峰命名）
		field.FieldName = g.toCamelCase(columnName)

		fields = append(fields, field)
	}

	return fields, nil
}

// getJavaType 获取Java类型
func (g *Generator) getJavaType(columnType string, driverName string) string {
	// 简化的类型映射，实际使用中可以根据需要扩展
	typeMap := map[string]string{
		"int":       "Integer",
		"integer":   "Integer",
		"bigint":    "Long",
		"varchar":   "String",
		"char":      "String",
		"text":      "String",
		"date":      "Date",
		"datetime":  "Date",
		"timestamp": "Date",
		"time":      "Date",
		"decimal":   "BigDecimal",
		"double":    "Double",
		"float":     "Float",
		"boolean":   "Boolean",
		"tinyint":   "Integer",
	}

	// 处理PostgreSQL特定类型
	if driverName == "postgres" {
		if strings.HasPrefix(columnType, "character varying") {
			columnType = "varchar"
		} else if columnType == "serial" || columnType == "bigserial" {
			columnType = "bigint"
		} else if columnType == "smallint" {
			columnType = "int"
		} else if columnType == "real" {
			columnType = "float"
		} else if columnType == "boolean" {
			columnType = "boolean"
		}
	}

	// 提取类型名（去除括号和长度）
	typeName := strings.Split(columnType, "(")[0]
	typeName = strings.ToLower(typeName)

	if javaType, ok := typeMap[typeName]; ok {
		return javaType
	}

	// 默认返回String
	return "String"
}

// toCamelCase 下划线转驼峰命名
func (g *Generator) toCamelCase(str string) string {
	parts := strings.Split(str, "_")
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

// toPascalCase 下划线转大驼峰命名
func (g *Generator) toPascalCase(str string) string {
	parts := strings.Split(str, "_")
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

// GenerateCode 生成代码
func (g *Generator) GenerateCode() error {
	// 获取所有表
	tables, err := g.GetTables()
	if err != nil {
		return err
	}

	for _, tableName := range tables {
		// 获取表元数据
		table, err := g.GetTableMetaData(tableName)
		if err != nil {
			return err
		}

		// 准备模板数据
		templateData := g.prepareTemplateData(table)

		// 生成实体类
		err = g.GenerateEntity(templateData)
		if err != nil {
			return err
		}

		// 生成Mapper接口
		err = g.GenerateMapper(templateData)
		if err != nil {
			return err
		}

		// 生成Mapper XML
		err = g.GenerateMapperXML(templateData)
		if err != nil {
			return err
		}

		// 生成Service接口
		err = g.GenerateService(templateData)
		if err != nil {
			return err
		}

		// 生成Service实现类
		err = g.GenerateServiceImpl(templateData)
		if err != nil {
			return err
		}

		// 生成Controller
		err = g.GenerateController(templateData)
		if err != nil {
			return err
		}
	}

	return nil
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

	// 生成类名（去掉前缀）
	className := table.TableName
	if g.Config.DBConfig.TablePrefix != "" {
		className = strings.TrimPrefix(className, g.Config.DBConfig.TablePrefix)
	}
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

// GenerateEntity 生成实体类
func (g *Generator) GenerateEntity(data TemplateData) error {
	// 加载外部模板文件
	tmplPath := filepath.Join(g.TemplatePath, EntityTemplateFile)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("加载实体类模板失败: %v", err)
	}
	return g.generateFile(g.getEntityOutputPath(data), tmpl, data)
}

// GenerateMapper 生成Mapper接口
func (g *Generator) GenerateMapper(data TemplateData) error {
	// 加载外部模板文件
	tmplPath := filepath.Join(g.TemplatePath, MapperTemplateFile)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("加载Mapper接口模板失败: %v", err)
	}
	return g.generateFile(g.getMapperOutputPath(data), tmpl, data)
}

// GenerateMapperXML 生成Mapper XML
func (g *Generator) GenerateMapperXML(data TemplateData) error {
	// 加载外部模板文件
	tmplPath := filepath.Join(g.TemplatePath, MapperXmlTemplateFile)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("加载Mapper XML模板失败: %v", err)
	}
	return g.generateFile(g.getMapperXmlOutputPath(data), tmpl, data)
}

// GenerateService 生成Service接口
func (g *Generator) GenerateService(data TemplateData) error {
	// 加载外部模板文件
	tmplPath := filepath.Join(g.TemplatePath, ServiceTemplateFile)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("加载Service接口模板失败: %v", err)
	}
	return g.generateFile(g.getServiceOutputPath(data), tmpl, data)
}

// GenerateServiceImpl 生成Service实现类
func (g *Generator) GenerateServiceImpl(data TemplateData) error {
	// 加载外部模板文件
	tmplPath := filepath.Join(g.TemplatePath, ServiceImplTemplateFile)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("加载Service实现类模板失败: %v", err)
	}
	return g.generateFile(g.getServiceImplOutputPath(data), tmpl, data)
}

// GenerateController 生成Controller
func (g *Generator) GenerateController(data TemplateData) error {
	// 加载外部模板文件
	tmplPath := filepath.Join(g.TemplatePath, ControllerTemplateFile)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("加载Controller模板失败: %v", err)
	}
	return g.generateFile(g.getControllerOutputPath(data), tmpl, data)
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

// getEntityOutputPath 获取实体类输出路径
func (g *Generator) getEntityOutputPath(data TemplateData) string {
	pkgConfig := g.Config.PackageConfig
	outputPath := g.Config.GenConfig.OutputPath
	if outputPath == "" {
		outputPath = "./output"
	}

	return filepath.Join(outputPath, strings.ReplaceAll(pkgConfig.EntityPackage, ".", "/"), data.ClassName+".java")
}

// getMapperOutputPath 获取Mapper接口输出路径
func (g *Generator) getMapperOutputPath(data TemplateData) string {
	pkgConfig := g.Config.PackageConfig
	outputPath := g.Config.GenConfig.OutputPath
	if outputPath == "" {
		outputPath = "./output"
	}

	return filepath.Join(outputPath, strings.ReplaceAll(pkgConfig.MapperPackage, ".", "/"), data.ClassName+"Mapper.java")
}

// getMapperXmlOutputPath 获取Mapper XML输出路径
func (g *Generator) getMapperXmlOutputPath(data TemplateData) string {
	outputPath := g.Config.GenConfig.OutputPath
	if outputPath == "" {
		outputPath = "./output"
	}

	return filepath.Join(outputPath, "src/main/resources/mapper", data.ClassName+"Mapper.xml")
}

// getServiceOutputPath 获取Service接口输出路径
func (g *Generator) getServiceOutputPath(data TemplateData) string {
	pkgConfig := g.Config.PackageConfig
	outputPath := g.Config.GenConfig.OutputPath
	if outputPath == "" {
		outputPath = "./output"
	}

	return filepath.Join(outputPath, strings.ReplaceAll(pkgConfig.ServicePackage, ".", "/"), "I"+data.ClassName+"Service.java")
}

// getServiceImplOutputPath 获取Service实现类输出路径
func (g *Generator) getServiceImplOutputPath(data TemplateData) string {
	pkgConfig := g.Config.PackageConfig
	outputPath := g.Config.GenConfig.OutputPath
	if outputPath == "" {
		outputPath = "./output"
	}

	return filepath.Join(outputPath, strings.ReplaceAll(pkgConfig.ServicePackage, ".", "/"), "impl", data.ClassName+"ServiceImpl.java")
}

// getControllerOutputPath 获取Controller输出路径
func (g *Generator) getControllerOutputPath(data TemplateData) string {
	pkgConfig := g.Config.PackageConfig
	outputPath := g.Config.GenConfig.OutputPath
	if outputPath == "" {
		outputPath = "./output"
	}

	return filepath.Join(outputPath, strings.ReplaceAll(pkgConfig.ControllerPackage, ".", "/"), data.ClassName+"Controller.java")
}

`

// Close 关闭数据库连接
func (g *Generator) Close() error {
	if g.DB != nil {
		return g.DB.Close()
	}
	return nil
}

// toCamelCase 下划线转驼峰命名
func (g *Generator) toCamelCase(str string) string {
	parts := strings.Split(str, "_")
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += strings.Title(parts[i])
	}
	return result
}
