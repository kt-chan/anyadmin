package service

import (
	"os"
	"path/filepath"
	"time"

	"github.com/anyzearch/admin/core/internal/global"
)

func StartImportTask(taskID uint) {
	var task global.ImportTask
	if err := global.DB.First(&task, taskID).Error; err != nil {
		return
	}

	// 1. 扫描目录获取文件总数
	var files []string
	filepath.Walk(task.SourcePath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	total := len(files)
	if total == 0 {
		global.DB.Model(&task).Updates(map[string]interface{}{
			"Status": "Completed",
			"Message": "目录为空，无文件可处理",
		})
		return
	}

	global.DB.Model(&task).Updates(map[string]interface{}{
		"TotalFiles": total,
		"Status":     "Running",
	})

	// 2. 模拟文件处理过程
	for i, _ := range files {
		// 检查任务是否被手动暂停/停止（此处可增加状态检查逻辑）
		
		time.Sleep(200 * time.Millisecond) // 模拟处理耗时
		
		processed := i + 1
		progress := (processed * 100) / total
		
		global.DB.Model(&task).Updates(map[string]interface{}{
			"Processed": processed,
			"Progress":  progress,
		})
	}

	global.DB.Model(&task).Update("Status", "Completed")
}
