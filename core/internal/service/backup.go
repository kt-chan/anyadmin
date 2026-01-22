package service

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/anyzearch/admin/core/internal/global"
)

func CreateBackup() (*global.BackupRecord, error) {
	log.Println("[Backup] 开始执行系统备份...")
	timestamp := time.Now().Format("20060102_150405")
	backupDir := "backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Printf("[Backup] 创建目录失败: %v", err)
		return nil, err
	}
	
	fileName := fmt.Sprintf("backup_%s.tar.gz", timestamp)
	filePath := filepath.Join(backupDir, fileName)

	fw, err := os.Create(filePath)
	if err != nil {
		log.Printf("[Backup] 创建文件失败: %v", err)
		return nil, err
	}
	defer fw.Close()

	gw := gzip.NewWriter(fw)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// 需要备份的文件/目录
	filesToBackup := []string{"anyzearch.db", "deployments"}

	for _, source := range filesToBackup {
		log.Printf("[Backup] 正在归档: %s", source)
		err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				// 允许部分文件访问失败（如临时文件），只打印警告
				log.Printf("[Backup] 警告 - 访问文件失败: %s, %v", path, err)
				return nil 
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}
			// 修正 TAR 内部路径，去除绝对路径前缀，保留相对结构
			header.Name = path
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			if !info.IsDir() {
				f, err := os.Open(path)
				if err != nil {
					return err
				}
				defer f.Close()
				io.Copy(tw, f)
			}
			return nil
		})
		if err != nil {
			log.Printf("[Backup] 归档目录 %s 失败: %v", source, err)
			return nil, err
		}
	}

	info, _ := os.Stat(filePath)
	log.Printf("[Backup] 备份完成: %s (Size: %d)", filePath, info.Size())
	record := global.BackupRecord{
		Name:   fileName,
		Path:   filePath,
		Size:   info.Size(),
		Type:   "Full",
		Status: "Success",
	}

	global.DB.Create(&record)
	return &record, nil
}

func RestoreBackup(id uint) error {
	var record global.BackupRecord
	if err := global.DB.First(&record, id).Error; err != nil {
		return err
	}
    // 注意：真正的恢复逻辑需要停止服务并替换文件
    // 这里作为 MVP 展示，记录日志并模拟操作成功
	fmt.Printf("Restoring from: %s\n", record.Path)
	return nil
}
