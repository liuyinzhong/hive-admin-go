package services

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

const (
	maxUploadFileSize = 200 * 1024 * 1024
	uploadPathValue   = "/uploads/"
	thumbPathValue    = "/uploads/thumb"
)

type FileService struct{}

func NewFileService() *FileService {
	return &FileService{}
}

func (s *FileService) UploadFile(fileHeader *multipart.FileHeader, creatorID string) (*models.FileResponse, error) {
	if fileHeader == nil {
		return nil, errors.New("请选择上传文件")
	}
	if fileHeader.Size <= 0 {
		return nil, errors.New("上传文件不能为空")
	}
	if fileHeader.Size >= maxUploadFileSize {
		return nil, errors.New("文件大小必须小于200M")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	contentType, err := detectContentType(src, fileHeader)
	if err != nil {
		return nil, err
	}

	fileID := utils.GenerateUUID()
	fileExt := resolveFileExt(fileHeader.Filename, contentType)
	storedName := strings.ReplaceAll(fileID, "-", "") + fileExt

	uploadDir := filepath.Join(".", "static", "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, err
	}

	dstPath := filepath.Join(uploadDir, storedName)
	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		_ = os.Remove(dstPath)
		return nil, err
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(dstPath)
		return nil, err
	}

	fullPathValue := uploadPathValue + storedName
	urlValue := fullPathValue

	var creatorIDPtr *string
	if creatorID != "" {
		creatorIDPtr = &creatorID
	}

	var thumbnailPath *string
	var thumbnailURL *string
	if strings.HasPrefix(strings.ToLower(contentType), "image/") {
		thumbDir := filepath.Join(uploadDir, "thumb")
		if err := os.MkdirAll(thumbDir, 0755); err != nil {
			_ = os.Remove(dstPath)
			return nil, err
		}

		thumbFilePath := filepath.Join(thumbDir, storedName)
		if err := copyFile(dstPath, thumbFilePath); err != nil {
			_ = os.Remove(dstPath)
			return nil, err
		}

		thumbURLValue := thumbPathValue + "/" + storedName
		thumbnailPath = stringToPtr(thumbPathValue)
		thumbnailURL = &thumbURLValue
	}

	now := time.Now()
	file := models.SysFile{
		FileID:        fileID,
		URL:           &urlValue,
		Name:          &storedName,
		Type:          &contentType,
		Size:          fileHeader.Size,
		FileExt:       stringToPtr(fileExt),
		OriginalName:  &fileHeader.Filename,
		Path:          stringToPtr(uploadPathValue),
		FullPath:      &fullPathValue,
		ThumbnailPath: thumbnailPath,
		ThumbnailURL:  thumbnailURL,
		CreatorID:     creatorIDPtr,
		CreateDate:    &now,
	}

	if err := database.DB.Create(&file).Error; err != nil {
		_ = os.Remove(dstPath)
		if thumbnailURL != nil {
			_ = os.Remove(filepath.Join(uploadDir, "thumb", storedName))
		}
		return nil, err
	}

	var creatorName string
	if creatorIDPtr != nil {
		var user models.SysUser
		if err := database.DB.Select("user_id", "real_name").Where("user_id = ? AND del_flag = 0", *creatorIDPtr).First(&user).Error; err == nil && user.RealName != nil {
			creatorName = *user.RealName
		}
	}

	return buildFileResponse(file, creatorName), nil
}

func (s *FileService) GetFileList(req models.FileListRequest) (*utils.PageResult, error) {
	db := database.DB.Model(&models.SysFile{})

	if req.OriginalName != "" {
		db = db.Where("original_name LIKE ?", "%"+req.OriginalName+"%")
	}
	if req.Type != "" {
		db = db.Where("LOWER(type) = ?", strings.ToLower(req.Type))
	}
	if req.FileExt != "" {
		db = db.Where("LOWER(file_ext) = ?", strings.ToLower(req.FileExt))
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"originalName": "original_name",
		"size":         "size",
		"createDate":   "create_date",
	})
	if order == "" {
		order = "create_date DESC"
	} else if !strings.Contains(strings.ToLower(order), "create_date") {
		order += ", create_date DESC"
	}
	db = db.Order(order)

	var files []models.SysFile
	pageResult, err := utils.Paginate(db, req.Page, req.PageSize, &files)
	if err != nil {
		return nil, err
	}

	creatorIDs := make([]string, 0)
	for _, file := range files {
		if file.CreatorID != nil && *file.CreatorID != "" {
			creatorIDs = append(creatorIDs, *file.CreatorID)
		}
	}

	creatorNames := make(map[string]string)
	if len(creatorIDs) > 0 {
		var users []models.SysUser
		database.DB.Select("user_id", "real_name").Where("user_id IN ? AND del_flag = 0", creatorIDs).Find(&users)
		for _, user := range users {
			if user.RealName != nil {
				creatorNames[user.UserID] = *user.RealName
			}
		}
	}

	items := make([]models.FileResponse, 0, len(files))
	for _, file := range files {
		items = append(items, *buildFileResponse(file, creatorNames[utils.StringValue(file.CreatorID)]))
	}

	pageResult.Items = items
	return pageResult, nil
}

func buildFileResponse(file models.SysFile, creatorName string) *models.FileResponse {
	return &models.FileResponse{
		FileID:        &file.FileID,
		URL:           file.URL,
		Name:          file.Name,
		Type:          file.Type,
		Size:          file.Size,
		FileExt:       file.FileExt,
		OriginalName:  file.OriginalName,
		Path:          file.Path,
		FullPath:      file.FullPath,
		ThumbnailPath: file.ThumbnailPath,
		ThumbnailURL:  file.ThumbnailURL,
		CreatorID:     file.CreatorID,
		CreatorName:   &creatorName,
		CreateDate:    models.TimeToStringPtr(file.CreateDate),
	}
}

func detectContentType(src multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buffer[:n])
	if contentType == "application/octet-stream" {
		headerContentType := fileHeader.Header.Get("Content-Type")
		if headerContentType != "" {
			contentType = headerContentType
		}
	}

	if contentType == "application/octet-stream" {
		fileExt := filepath.Ext(fileHeader.Filename)
		if fileExt != "" {
			if mimeType := mime.TypeByExtension(strings.ToLower(fileExt)); mimeType != "" {
				contentType = mimeType
			}
		}
	}

	return contentType, nil
}

func resolveFileExt(fileName string, contentType string) string {
	fileExt := strings.ToLower(filepath.Ext(fileName))
	if fileExt != "" {
		return fileExt
	}

	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil || len(extensions) == 0 {
		return ""
	}

	return strings.ToLower(extensions[0])
}

func copyFile(srcPath string, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		_ = os.Remove(dstPath)
		return err
	}

	return nil
}

func stringToPtr(s string) *string {
	return &s
}
