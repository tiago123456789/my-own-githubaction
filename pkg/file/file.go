package file

import (
	"os"

	"go.uber.org/zap"
)

type IFile interface {
	WriteFile(fileName string, data string) error
}

type File struct {
	logger *zap.Logger
}

func New(
	logger *zap.Logger,
) *File {
	return &File{
		logger: logger,
	}
}

func (f *File) WriteFile(fileName string, data string) error {
	file, err := os.Create(fileName)
	defer file.Close()

	if err != nil {
		return err
	}

	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}
