package paradise

type FileSystem interface {
	GetFiles() []string
}

type FileManager struct {
	FileSystem
}

type DefaultFileSystem struct {
}

func (dfs DefaultFileSystem) GetFiles() []string {
	files := make([]string, 5)

	return files
}

func NewDefaultFileSystem() *FileManager {
	fm := FileManager{}
	fm.FileSystem = DefaultFileSystem{}
	return &fm
}
