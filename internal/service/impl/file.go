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

	"github.com/mycodeLife01/qa/internal/model"
	"github.com/mycodeLife01/qa/internal/pkg/api"
	"github.com/mycodeLife01/qa/internal/service"
	"github.com/tencentyun/cos-go-sdk-v5"
	"gorm.io/gorm"
)

type fileService struct {
	DB        *gorm.DB
	CosClient *cos.Client
}

func NewFileService(db *gorm.DB, client *cos.Client) service.FileService {
	return &fileService{DB: db, CosClient: client}
}

func (fs *fileService) Upload(fileHeader *multipart.FileHeader) (string, error) {
	// 1. 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		fmt.Printf("failed to open file: %v\n", err)
	}
	defer file.Close()

	// 2. 查询是否已存储该文件
	hashCalculator := sha256.New()
	_, copyErr := io.Copy(hashCalculator, file)
	if copyErr != nil {
		fmt.Printf("failed to copy file: %v\n", copyErr)
		return "", copyErr
	}
	hashString := hex.EncodeToString(hashCalculator.Sum(nil))

	var existingFile model.File
	queryErr := fs.DB.Where("content_hash = ?", hashString).First(&existingFile).Error

	if queryErr == nil {
		// 文件已存在，直接返回
		return existingFile.ContentHash, nil
	} else if errors.Is(queryErr, gorm.ErrRecordNotFound) {
		// 文件不存在，继续执行上传操作
		_, seekErr := file.Seek(0, io.SeekStart)
		if seekErr != nil {
			fmt.Printf("failed to seek file: %v\n", seekErr)
			return "", seekErr
		}
		ext := filepath.Ext(fileHeader.Filename)
		if ext == "" || (ext != ".pdf" && ext != ".docx" && ext != ".doc" && ext != ".txt") {
			return "", api.ErrUploadFileExtInvalid
		}
		fullname := fmt.Sprintf("uploads/%s%s", hashString, ext)
		_, putErr := fs.CosClient.Object.Put(context.Background(), fullname, file, nil)
		if putErr != nil {
			fmt.Printf("failed to put file: %v\n", putErr)
			return "", putErr
		}
		// 保存文件记录到数据库
		uploadFile := model.File{
			ContentHash: hashString,
			ObjectKey:   fullname,
			BucketName:  "my-qa-go-1313494932",
			FileType:    ext[1:],
		}
		createErr := fs.DB.Create(&uploadFile).Error
		if createErr != nil {
			fmt.Printf("failed to create file record: %v\n", createErr)
			return "", createErr
		}
		return hashString, nil

	} else {
		fmt.Printf("failed to query file: %v\n", queryErr)
		return "", queryErr
	}
}
