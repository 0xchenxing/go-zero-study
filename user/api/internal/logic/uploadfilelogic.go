// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"user/api/internal/svc"
	"user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
)

type UploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadFileLogic {
	return &UploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadFileLogic) UploadFile(file multipart.File, header *multipart.FileHeader) (*types.UploadResp, error) {
	// 校验大小
	const maxSize = 10 << 20 // 10 MB
	if header.Size > maxSize {
		return nil, errors.New(422, "文件超过 10 MB 限制")
	}

	// 安全处理文件名，防路径穿越
	safeFilename := fmt.Sprintf("%d_%s", time.Now().UnixNano(),
		filepath.Base(filepath.Clean(header.Filename)))

	if _, err := os.Stat(l.svcCtx.Config.UploadDir); os.IsNotExist(err) {
		err = os.MkdirAll(l.svcCtx.Config.UploadDir, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("无法创建上传目录: %w", err)
		}
	}

	dst, err := os.Create(filepath.Join(l.svcCtx.Config.UploadDir, safeFilename))
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	size, err := io.Copy(dst, file)
	if err != nil {
		return nil, err
	}

	return &types.UploadResp{
		Filename: header.Filename,
		Size:     size,
		URL:      "/files/" + safeFilename,
	}, nil
}

//func (l *UploadFileLogic) uploadToS3(file multipart.File, key string) (string, error) {
//	_, err := l.svcCtx.S3.PutObject(l.ctx, &s3.PutObjectInput{
//		Bucket: aws.String(l.svcCtx.Config.S3Bucket),
//		Key:    aws.String(key),
//		Body:   file,
//	})
//	if err != nil {
//		return "", err
//	}
//	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s",
//		l.svcCtx.Config.S3Bucket, key), nil
//}
