package script

import (
	"ai-ops/internal/model"
	"ai-ops/internal/repository"
	"log"
)

// InitBuiltinScripts 初始化内置脚本
func InitBuiltinScripts(repo repository.ScriptRepository) error {
	scripts := GetBuiltinScripts()

	for _, script := range scripts {
		// 检查是否已存在
		existing, _ := repo.GetByName(script.Name)
		if existing != nil {
			continue // 已存在则跳过
		}

		if err := repo.Create(script); err != nil {
			log.Printf("创建内置脚本 %s 失败: %v", script.Name, err)
			continue
		}
		log.Printf("已创建内置脚本: %s", script.Name)
	}

	return nil
}

// GetBuiltinScriptNames 获取内置脚本名称列表
func GetBuiltinScriptNames() []string {
	scripts := GetBuiltinScripts()
	names := make([]string, len(scripts))
	for i, s := range scripts {
		names[i] = s.Name
	}
	return names
}

// IsBuiltinScript 检查是否为内置脚本
func IsBuiltinScript(name string) bool {
	for _, n := range GetBuiltinScriptNames() {
		if n == name {
			return true
		}
	}
	return false
}

// ResetBuiltinScript 重置内置脚本到默认状态
func ResetBuiltinScript(repo repository.ScriptRepository, name string) error {
	for _, script := range GetBuiltinScripts() {
		if script.Name == name {
			existing, err := repo.GetByName(name)
			if err != nil {
				return repo.Create(script)
			}
			script.ID = existing.ID
			return repo.Update(script)
		}
	}
	return nil
}

// ResetAllBuiltinScripts 重置所有内置脚本
func ResetAllBuiltinScripts(repo repository.ScriptRepository) error {
	for _, script := range GetBuiltinScripts() {
		existing, _ := repo.GetByName(script.Name)
		if existing != nil {
			script.ID = existing.ID
			if err := repo.Update(script); err != nil {
				log.Printf("重置内置脚本 %s 失败: %v", script.Name, err)
			}
		} else {
			if err := repo.Create(script); err != nil {
				log.Printf("创建内置脚本 %s 失败: %v", script.Name, err)
			}
		}
	}
	return nil
}

// GetBuiltinScript 获取指定的内置脚本
func GetBuiltinScript(name string) *model.Script {
	for _, script := range GetBuiltinScripts() {
		if script.Name == name {
			return script
		}
	}
	return nil
}
