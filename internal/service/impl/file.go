package impl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/mycodeLife01/qa/config"
	"github.com/mycodeLife01/qa/internal/model"
	"github.com/mycodeLife01/qa/internal/pkg/api"
	"github.com/mycodeLife01/qa/internal/pkg/client"
	"github.com/mycodeLife01/qa/internal/service"
	"github.com/tencentyun/cos-go-sdk-v5"
	"gorm.io/gorm"
)

type fileService struct {
	DB         *gorm.DB
	CosClient  *cos.Client
	TaskClient *client.CeleryClient
}

func NewFileService(db *gorm.DB, client *cos.Client, taskClient *client.CeleryClient) service.FileService {
	return &fileService{DB: db, CosClient: client, TaskClient: taskClient}
}

func (fs *fileService) Upload(username string, fileHeader *multipart.FileHeader, force bool) (map[string]string, error) {
	// 1. 查询用户
	var user model.User
	if err := fs.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.ErrUserInvalid
		}
		return nil, err
	}

	// 2. 打开文件并计算哈希
	file, err := fileHeader.Open()
	if err != nil {
		fmt.Printf("failed to open file: %v\n", err)
		return nil, err
	}
	defer file.Close()

	hashCalculator := sha256.New()
	if _, err := io.Copy(hashCalculator, file); err != nil {
		fmt.Printf("failed to calculate hash: %v\n", err)
		return nil, err
	}
	hashString := hex.EncodeToString(hashCalculator.Sum(nil))

	// 3. 验证文件扩展名
	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" || (ext != ".pdf" && ext != ".docx" && ext != ".doc" && ext != ".txt") {
		return nil, api.ErrUploadFileExtInvalid
	}

	// 4. 使用事务处理文件上传和关联创建
	var resultFile model.File
	var needIndex bool = false

	err = fs.DB.Transaction(func(tx *gorm.DB) error {
		// 4.1 查找或创建 File 记录（包括已软删除的）
		var existingFile model.File
		queryErr := tx.Unscoped().Where("content_hash = ?", hashString).First(&existingFile).Error

		if queryErr == nil {
			// 文件记录已存在
			if existingFile.DeletedAt.Valid {
				// 恢复软删除的文件
				if err := tx.Unscoped().Model(&existingFile).Update("deleted_at", nil).Error; err != nil {
					return err
				}
			}
			resultFile = existingFile

			// 检查是否需要索引
			if force {
				// 强制重新索引，重置状态为 PENDING
				if err := tx.Model(&existingFile).Update("status", "PENDING").Error; err != nil {
					return err
				}
				needIndex = true
			} else if existingFile.Status != "SUCCESS" {
				// 如果之前失败或挂起，重新索引
				if existingFile.Status != "PENDING" {
					if err := tx.Model(&existingFile).Update("status", "PENDING").Error; err != nil {
						return err
					}
				}
				needIndex = true
			}
			// 如果 Status == SUCCESS 且 force == false，不需要索引
		} else if errors.Is(queryErr, gorm.ErrRecordNotFound) {
			// 文件不存在，上传到 COS 并创建记录
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				return err
			}

			fullname := fmt.Sprintf("%s/%s%s", config.C.COS.UploadFolder, hashString, ext)
			if _, err := fs.CosClient.Object.Put(context.Background(), fullname, file, nil); err != nil {
				fmt.Printf("failed to upload file to COS: %v\n", err)
				return err
			}

			newFile := model.File{
				ContentHash: hashString,
				ObjectKey:   fullname,
				BucketName:  "my-qa-go-1313494932",
				FileType:    ext[1:],
				Status:      "PENDING",
			}
			if err := tx.Create(&newFile).Error; err != nil {
				return err
			}
			resultFile = newFile
			needIndex = true
		} else {
			return queryErr
		}

		// 4.2 创建用户-文件关联（如果不存在）
		var existingUserFile model.UserFile
		userFileErr := tx.Where("user_id = ? AND file_id = ?", user.ID, resultFile.ID).First(&existingUserFile).Error

		if errors.Is(userFileErr, gorm.ErrRecordNotFound) {
			// 关联不存在，创建新关联
			newUserFile := model.UserFile{
				UserID:   user.ID,
				FileID:   resultFile.ID,
				FileName: fileHeader.Filename,
			}
			if err := tx.Create(&newUserFile).Error; err != nil {
				return err
			}
		} else if userFileErr != nil {
			return userFileErr
		}
		// 如果关联已存在，不做任何操作（幂等）

		return nil
	})

	if err != nil {
		fmt.Printf("failed to upload file: %v\n", err)
		return nil, err
	}

	resultMap := map[string]string{
		"content_hash": resultFile.ContentHash,
		"file_name":    fileHeader.Filename,
	}

	// 5. 触发索引任务
	if needIndex {
		taskID, err := fs.TaskClient.CreateIndexTask(resultFile.ContentHash, fs.getFileUrl(resultFile.ObjectKey), resultFile.FileType)
		if err != nil {
			fmt.Printf("failed to create index task: %v\n", err)
			// 即使创建任务失败，文件已上传成功，所以不返回错误，只是没有 task_id
		} else {
			resultMap["task_id"] = taskID
		}
	}

	return resultMap, nil
}

func (fs *fileService) UpdateStatus(contentHash string, status string) error {
	return fs.DB.Model(&model.File{}).Where("content_hash = ?", contentHash).Update("status", status).Error
}

func (fs *fileService) getFileUrl(objectKey string) string {
	return fs.CosClient.Object.GetObjectURL(objectKey).String()
}


func (fs *fileService) DeleteFileByHash(username string, hash string) (bool, error) {
	// 1. 查询用户
	var user model.User
	if err := fs.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, api.ErrUserInvalid
		}
		return false, err
	}

	// 2. 使用事务处理删除逻辑
	err := fs.DB.Transaction(func(tx *gorm.DB) error {
		// 2.1 查找文件
		var file model.File
		if err := tx.Where("content_hash = ?", hash).First(&file).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return api.ErrFileNotFound
			}
			return err
		}

		// 2.2 删除当前用户与该文件的关联
		result := tx.Where("user_id = ? AND file_id = ?", user.ID, file.ID).Delete(&model.UserFile{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			// 用户与该文件没有关联
			return api.ErrFileNotFound
		}

		// 2.3 检查是否还有其他用户关联该文件
		var count int64
		if err := tx.Model(&model.UserFile{}).Where("file_id = ?", file.ID).Count(&count).Error; err != nil {
			return err
		}

		// 2.4 如果没有任何用户关联，软删除文件
		if count == 0 {
			if err := tx.Delete(&file).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

func (fs *fileService) GetAllUserFiles(username string) ([]map[string]string, error) {
	// 1. 查询用户
	var user model.User
	if err := fs.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.ErrUserInvalid
		}
		return nil, err
	}

	// 2. 通过 JOIN 查询用户的所有文件
	type FileInfo struct {
		ContentHash string
		FileName    string
		Status      string
	}

	var files []FileInfo
	err := fs.DB.Table("user_files").
		Select("files.content_hash, user_files.file_name, files.status").
		Joins("JOIN files ON files.id = user_files.file_id").
		Where("user_files.user_id = ? AND user_files.deleted_at IS NULL AND files.deleted_at IS NULL", user.ID).
		Find(&files).Error

	if err != nil {
		return nil, err
	}

	// 3. 转换为返回格式
	fileInfos := make([]map[string]string, len(files))
	for i, f := range files {
		fileInfos[i] = map[string]string{
			"content_hash": f.ContentHash,
			"file_name":    f.FileName,
			"status":       f.Status,
		}
	}

	return fileInfos, nil
}

func (fs *fileService) GetFileByHash(hash string) (*model.File, error) {
	var file model.File
	if err := fs.DB.Where("content_hash = ?", hash).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}
