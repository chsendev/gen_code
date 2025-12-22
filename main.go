package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"gen_code/pkg/gencode"
)

func main() {
	// 命令行参数
	configPath := flag.String("config", "config.json", "配置文件路径")
	flag.Parse()

	// 读取配置文件
	config, err := readConfig(*configPath)
	if err != nil {
		fmt.Printf("读取配置文件失败: %v\n", err)
		os.Exit(1)
	}

	// 创建代码生成器
	generator := gencode.NewGenerator(config)

	// 初始化
	if err := generator.Init(); err != nil {
		fmt.Printf("初始化代码生成器失败: %v\n", err)
		os.Exit(1)
	}
	defer generator.Close()

	// 生成代码
	if err := generator.GenerateCode(); err != nil {
		fmt.Printf("生成代码失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("代码生成成功！")
}

// readConfig 读取配置文件
func readConfig(path string) (gencode.Config, error) {
	var config gencode.Config

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 如果配置文件不存在，生成默认配置
		config = generateDefaultConfig()
		// 保存默认配置到文件
		if err := saveConfig(config, path); err != nil {
			return config, fmt.Errorf("保存默认配置失败: %v", err)
		}
		fmt.Printf("已生成默认配置文件: %s\n", path)
		return config, nil
	}

	// 读取配置文件
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析配置
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return config, nil
}

// generateDefaultConfig 生成默认配置
func generateDefaultConfig() gencode.Config {
	return gencode.Config{
		DBConfig: gencode.DBConfig{
			DriverName:    "mysql",
			Host:          "localhost",
			Port:          3306,
			Username:      "root",
			Password:      "password",
			DatabaseName:  "test_db",
			TablePrefix:   "t_",
			IncludeTables: []string{},
			ExcludeTables: []string{},
		},
		GenConfig: gencode.GenConfig{
			OutputPath:    "./output",
			EnableLombok:  true,
			EnableSwagger: true,
			Author:        "admin",
			Date:          "",
		},
		PackageConfig: gencode.PackageConfig{
			BasePackage:       "com.example",
			EntityPackage:     "com.example.entity",
			MapperPackage:     "com.example.mapper",
			ServicePackage:    "com.example.service",
			ControllerPackage: "com.example.controller",
		},
	}
}

// saveConfig 保存配置到文件
func saveConfig(config gencode.Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}
