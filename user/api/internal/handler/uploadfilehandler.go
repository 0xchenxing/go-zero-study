// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"fmt"
	"io"
	"net/http"

	"user/api/internal/logic"
	"user/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UploadFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 解析 Multipart 表单，限制内存缓存为 32MB，超出部分自动落盘
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			httpx.Error(w, err)
			return
		}

		// 2. 获取文件
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.Error(w, err)
			return
		}
		defer file.Close()

		// 3. 通过读取前 512 字节检测 MIME 类型 (不能仅依赖后缀名)
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		mimeType := http.DetectContentType(buf[:n])

		if !isAllowedType(mimeType) {
			httpx.Error(w, fmt.Errorf("不支持的文件类型: %s", mimeType))
			// 注意：图片中使用了 http.StatusUnsupportedMediaType (415)
			// 这里使用 httpx.Error 配合自定义错误处理，或者直接返回 415
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		// 4. 重要：检测完 MIME 类型后，必须将文件指针重置回开头，否则后续读取为空
		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		// 5. 调用 Logic 层处理
		l := logic.NewUploadFileLogic(r.Context(), svcCtx)
		resp, err := l.UploadFile(file, header)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, resp)
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
