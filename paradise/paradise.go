package paradise

type FileSytem interface {
	GetFiles() []string
}

type FileManager struct {
	FileSystem
}

type DefaultFileSystem struct {
}

func (dfs *FileSystem) GetFiles() []string {
	files := make([]string, 5)

	return files
}

func NewDefaultFileSystem() *FileManager {
	fm := FileManager{}
	fm.FileSystem = DefaultFileSystem{}
	return &fm
}
