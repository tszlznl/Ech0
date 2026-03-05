package zip

import (
	fileUtil "github.com/lin-snow/ech0/internal/util/file"
	"github.com/spf13/afero"
)

type Port interface {
	ZipDirectory(sourceDir, zipPath string, options fileUtil.ZipOptions) error
	UnzipFile(src, dest string) error
	CopyDirectory(src, dest string) error
}

type Module struct {
	fs afero.Fs
}

func NewModule(fs afero.Fs) *Module {
	return &Module{fs: fs}
}

func (m *Module) ZipDirectory(sourceDir, zipPath string, options fileUtil.ZipOptions) error {
	return fileUtil.ZipDirectoryWithOptions(m.fs, sourceDir, zipPath, options)
}

func (m *Module) UnzipFile(src, dest string) error {
	return fileUtil.UnzipFile(m.fs, src, dest)
}

func (m *Module) CopyDirectory(src, dest string) error {
	return fileUtil.CopyDirectory(m.fs, src, dest)
}

