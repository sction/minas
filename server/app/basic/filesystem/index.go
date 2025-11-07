package filesystem

import (
	"os"
	"path/filepath"
	"server/core/app/response"
	"server/utils/logger"
	"strings"

	"github.com/gin-gonic/gin"
)

// List 列出指定路径下的文件和目录
func List(ctx *gin.Context) {
	fstype := ctx.Query("fstype")
	suffix := ctx.Query("suffix")

	path := ctx.Query("path")
	if path == "" {
		path = "/"
	}
	suffixs := []string{}
	if suffix != "" {
		suffixs = strings.Split(suffix, ",")
	}
	logger.LOG.Debugf("List files in path: %s, fstype: %s, suffix: %v", path, fstype, suffixs)
	files := listFiles(path, fstype, suffixs)
	response.Data(ctx, "", files)
}

// FileItem 文件项
type FileItem struct {
	Name      string `json:"name"`             // 文件名
	IsDir     bool   `json:"isDir"`            // 是否为目录
	IsLink    bool   `json:"isLink"`           // 是否为符号链接
	LinkTo    string `json:"linkTo,omitempty"` // 如果是符号链接，指向的目标路径
	Size      int64  `json:"size,omitempty"`
	Path      string `json:"path"`                // 文件的完整路径
	CreatedAt string `json:"createdAt,omitempty"` // 创建时间
}

// ListFiles 列出指定路径下的文件和目录
func listFiles(path string, fstype string, suffixs []string) []FileItem {
	var files []FileItem

	// 检查路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return files
	}

	// 读取目录内容
	entries, err := os.ReadDir(path)
	if err != nil {
		return files
	}

	// 处理每个条目
	for _, entry := range entries {
		itemPath := filepath.Join(path, entry.Name())

		// 使用 os.Stat 而不是 entry.Info() 来正确处理符号链接
		// os.Stat 会跟随符号链接获取目标文件的信息
		info, err := os.Stat(itemPath)
		if err != nil {
			// 如果符号链接指向不存在的文件，跳过该条目
			continue
		}
		if (fstype == "dir" && !info.IsDir()) || (fstype == "file" && info.IsDir()) {
			continue // 根据文件类型过滤
		}

		if !info.IsDir() && len(suffixs) > 0 {
			// 检查文件后缀是否匹配
			matched := false
			for _, suffix := range suffixs {
				if !strings.HasPrefix(suffix, ".") {
					suffix = "." + suffix
				}
				if filepath.Ext(entry.Name()) == suffix {
					matched = true
					break
				}
			}
			if !matched {
				continue // 如果不匹配后缀，跳过该条目
			}
		}

		item := FileItem{
			Name:      entry.Name(),
			IsDir:     info.IsDir(), // 使用真实文件信息判断是否为目录
			Path:      itemPath,
			CreatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
		}
		if entry.Type()&os.ModeSymlink != 0 {
			item.IsLink = true
			item.LinkTo, _ = os.Readlink(itemPath) // 获取符号链接指向的目标路径
		}
		if !info.IsDir() {
			item.Size = info.Size()
		}

		files = append(files, item)
	}

	return files
}
