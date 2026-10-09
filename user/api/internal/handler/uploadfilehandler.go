// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"io"
	"net/http"

	"user/api/common/errcode"
	"user/api/internal/logic"
	"user/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UploadFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 解析 Multipart 表单，限制内存缓存为 32MB，超出部分自动落盘
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, errcode.ErrBadRequest.WithDetail(err.Error()))
			return
		}

		// 2. 获取文件
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, errcode.ErrBadRequest.WithDetail(err.Error()))
			return
		}
		defer file.Close()

		// 3. 通过读取前 512 字节检测 MIME 类型 (不能仅依赖后缀名)
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		mimeType := http.DetectContentType(buf[:n])

		if !isAllowedType(mimeType) {
			httpx.ErrorCtx(r.Context(), w, errcode.ErrUploadFileType.WithDetail(mimeType))
			return
		}

		// 4. 重要：检测完 MIME 类型后，必须将文件指针重置回开头，否则后续读取为空
		if _, err = file.Seek(0, io.SeekStart); err != nil {
			httpx.ErrorCtx(r.Context(), w, errcode.ErrInternalServer.WithDetail(err.Error()))
			return
		}

		// 5. 调用 Logic 层处理
		l := logic.NewUploadFileLogic(r.Context(), svcCtx)
		resp, err := l.UploadFile(file, header)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		okJson(r.Context(), w, resp)
	}
}

// 允许上传的 MIME 类型白名单
var allowedTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"application/pdf": true,
}

func isAllowedType(mimeType string) bool {
	return allowedTypes[mimeType]
}
