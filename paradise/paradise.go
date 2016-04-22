package paradise

type FileSystem interface {
	GetFiles() []map[string]string
}

type FileManager struct {
	FileSystem
}

type DefaultFileSystem struct {
}

func (dfs DefaultFileSystem) GetFiles() []map[string]string {
	files := make([]map[string]string, 0)

	file := make(map[string]string)
	file["size"] = "123"
	file["name"] = "hello.txt"

	files = append(files, file)

	return files
}

func NewDefaultFileSystem() *FileManager {
	fm := FileManager{}
	fm.FileSystem = DefaultFileSystem{}
	return &fm
}
