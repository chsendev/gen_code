package gencode

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCode(t *testing.T) {
	// 准备测试数据 - 模拟用户表
	userTable := Table{
		TableName:    "user",
		TableComment: "用户表",
		Fields: []Field{
			{
				ColumnName:    "id",
				ColumnType:    "bigint",
				ColumnComment: "主键ID",
				IsNullable:    false,
				IsPrimaryKey:  true,
				JavaType:      "Long",
				FieldName:     "id",
			},
			{
				ColumnName:    "username",
				ColumnType:    "varchar",
				ColumnComment: "用户名",
				IsNullable:    false,
				IsPrimaryKey:  false,
				JavaType:      "String",
				FieldName:     "username",
			},
			{
				ColumnName:    "email",
				ColumnType:    "varchar",
				ColumnComment: "邮箱",
				IsNullable:    true,
				IsPrimaryKey:  false,
				JavaType:      "String",
				FieldName:     "email",
			},
			{
				ColumnName:    "age",
				ColumnType:    "int",
				ColumnComment: "年龄",
				IsNullable:    true,
				IsPrimaryKey:  false,
				JavaType:      "Integer",
				FieldName:     "age",
			},
			{
				ColumnName:    "created_time",
				ColumnType:    "datetime",
				ColumnComment: "创建时间",
				IsNullable:    false,
				IsPrimaryKey:  false,
				JavaType:      "Date",
				FieldName:     "createdTime",
			},
			{
				ColumnName:    "updated_time",
				ColumnType:    "datetime",
				ColumnComment: "更新时间",
				IsNullable:    true,
				IsPrimaryKey:  false,
				JavaType:      "Date",
				FieldName:     "updatedTime",
			},
		},
		PrimaryKey: Field{
			ColumnName:    "id",
			ColumnType:    "bigint",
			ColumnComment: "主键ID",
			IsNullable:    false,
			IsPrimaryKey:  true,
			JavaType:      "Long",
			FieldName:     "id",
		},
	}

	// 准备测试数据 - 模拟产品表
	productTable := Table{
		TableName:    "product",
		TableComment: "产品表",
		Fields: []Field{
			{
				ColumnName:    "id",
				ColumnType:    "bigint",
				ColumnComment: "主键ID",
				IsNullable:    false,
				IsPrimaryKey:  true,
				JavaType:      "Long",
				FieldName:     "id",
			},
			{
				ColumnName:    "product_name",
				ColumnType:    "varchar",
				ColumnComment: "产品名称",
				IsNullable:    false,
				IsPrimaryKey:  false,
				JavaType:      "String",
				FieldName:     "productName",
			},
			{
				ColumnName:    "price",
				ColumnType:    "decimal",
				ColumnComment: "价格",
				IsNullable:    false,
				IsPrimaryKey:  false,
				JavaType:      "BigDecimal",
				FieldName:     "price",
			},
			{
				ColumnName:    "description",
				ColumnType:    "text",
				ColumnComment: "产品描述",
				IsNullable:    true,
				IsPrimaryKey:  false,
				JavaType:      "String",
				FieldName:     "description",
			},
			{
				ColumnName:    "status",
				ColumnType:    "tinyint",
				ColumnComment: "状态(0:下架 1:上架)",
				IsNullable:    false,
				IsPrimaryKey:  false,
				JavaType:      "Integer",
				FieldName:     "status",
			},
		},
		PrimaryKey: Field{
			ColumnName:    "id",
			ColumnType:    "bigint",
			ColumnComment: "主键ID",
			IsNullable:    false,
			IsPrimaryKey:  true,
			JavaType:      "Long",
			FieldName:     "id",
		},
	}

	// 准备表列表
	tables := []Table{userTable, productTable}

	// 创建配置
	config := Config{
		ProjectName: "gentest",
		GenConfig: GenConfig{
			OutputPath:    "tmp/maven_project",
			EnableLombok:  true,
			EnableSwagger: true,
			Author:        "CodeGenerator",
		},
		PackageConfig: PackageConfig{
			BasePackage:       "com.example",
			EntityPackage:     "com.example.entity",
			MapperPackage:     "com.example.mapper",
			ServicePackage:    "com.example.service",
			ControllerPackage: "com.example.controller",
		},
	}

	// 创建生成器
	generator := NewGenerator(config, tables)

	// 调试：打印模板路径
	t.Logf("模板路径: %s", generator.TemplatePath)
	t.Logf("实体模板完整路径: %s", filepath.Join(generator.TemplatePath))

	// 初始化生成器
	err := generator.Init()
	if err != nil {
		t.Fatalf("初始化生成器失败: %v", err)
	}

	// 生成代码
	err = generator.GenerateCode()
	if err != nil {
		t.Fatalf("生成代码失败: %v", err)
	}

	t.Logf("代码生成成功！")

	// 验证生成的文件是否存在
	outputPath := "tmp/maven_project"

	// 检查目录结构
	expectedDirs := []string{
		"src/main/java/com/example/entity",
		"src/main/java/com/example/mapper",
		"src/main/java/com/example/service",
		"src/main/java/com/example/service/impl",
		"src/main/java/com/example/controller",
		"src/main/resources",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(outputPath, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("目录不存在: %s", dirPath)
		}
	}

	// 检查生成的文件
	expectedFiles := []string{
		// Maven pom.xml
		"pom.xml",

		// SpringBoot配置文件
		"src/main/java/com/example/Application.java",
		"src/main/resources/application.yml",

		// User 相关文件
		"src/main/java/com/example/entity/User.java",
		"src/main/java/com/example/mapper/UserMapper.java",
		"src/main/java/com/example/service/IUserService.java",
		"src/main/java/com/example/service/impl/UserServiceImpl.java",
		"src/main/java/com/example/controller/UserController.java",
		"src/main/java/com/example/mapper/UserMapper.xml",

		// Product 相关文件
		"src/main/java/com/example/entity/Product.java",
		"src/main/java/com/example/mapper/ProductMapper.java",
		"src/main/java/com/example/service/IProductService.java",
		"src/main/java/com/example/service/impl/ProductServiceImpl.java",
		"src/main/java/com/example/controller/ProductController.java",
		"src/main/java/com/example/mapper/ProductMapper.xml",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(outputPath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("文件不存在: %s", filePath)
		} else {
			t.Logf("文件生成成功: %s", filePath)
		}
	}

	// 关闭生成器
	err = generator.Close()
	if err != nil {
		t.Errorf("关闭生成器失败: %v", err)
	}

	t.Logf("代码生成测试完成，生成的文件位于: %s", outputPath)
}

func TestToCamelCase(t *testing.T) {
	generator := &Generator{}

	testCases := []struct {
		input    string
		expected string
	}{
		{"user_name", "userName"},
		{"created_time", "createdTime"},
		{"id", "id"},
		{"product_category_id", "productCategoryId"},
		{"", ""},
	}

	for _, tc := range testCases {
		result := generator.toCamelCase(tc.input)
		if result != tc.expected {
			t.Errorf("toCamelCase(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestToPascalCase(t *testing.T) {
	generator := &Generator{}

	testCases := []struct {
		input    string
		expected string
	}{
		{"user", "User"},
		{"user_info", "UserInfo"},
		{"product_category", "ProductCategory"},
		{"", ""},
	}

	for _, tc := range testCases {
		result := generator.toPascalCase(tc.input)
		if result != tc.expected {
			t.Errorf("toPascalCase(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

// 清理测试生成的文件
func TestCleanup(t *testing.T) {
	outputPath := "tmp/maven_project"
	err := os.RemoveAll(outputPath)
	if err != nil {
		t.Logf("清理测试文件失败: %v", err)
	} else {
		t.Logf("测试文件清理完成: %s", outputPath)
	}
}
