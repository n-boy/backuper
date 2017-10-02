package testutils

import (
	"github.com/n-boy/backuper/base"

	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const defaultFileSize = 10

type TestFileSystem struct {
	basePath    string
	dataPath    string
	restorePath string
	storagePath string
	appPath     string
	fileSize    int
}

type CmdsToApply map[string][]string

func CreateTestFileSystem() TestFileSystem {
	fs := TestFileSystem{}
	fs.Init()
	return fs
}

func (fs *TestFileSystem) Init() {
	dirPath := filepath.Join(TmpDir(), RandString(20))
	err := os.Mkdir(dirPath, 0770)
	if err == nil {
		fs.basePath = dirPath
		fs.dataPath = filepath.Join(dirPath, "_data")
		fs.restorePath = filepath.Join(dirPath, "_restore")
		fs.storagePath = filepath.Join(dirPath, "_storage")
		fs.appPath = filepath.Join(dirPath, "_app")
		for _, subDir := range []string{fs.storagePath, fs.dataPath, fs.appPath, fs.restorePath} {
			err = os.Mkdir(subDir, 0770)
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		panic(err)
	}

	fs.fileSize = defaultFileSize
}

func (fs *TestFileSystem) Destroy() {
	if fs.basePath == "" {
		panic("TestFileSystem must be initialized before destroy")
	}
	if base.IsPathInBasePath(TmpDir(), fs.basePath) {
		os.RemoveAll(fs.basePath)
		fs.basePath = ""
		fs.dataPath = ""
		fs.storagePath = ""
	} else {
		panic("Could not remove files from non-tmp path")
	}
}

func (fs *TestFileSystem) ApplyCmds(cmds CmdsToApply) error {
	for cmd, pathes := range cmds {
		for _, relPath := range pathes {
			var err error
			switch cmd {
			case "create":
				err = fs.Create(relPath)
			case "modify":
				err = fs.Modify(relPath)
			case "modify_time":
				err = fs.ModifyTime(relPath)
			case "modify_size":
				err = fs.ModifySize(relPath)
			case "delete":
				err = fs.Delete(relPath)
			default:
				err = fmt.Errorf("Command is not supported: %v", cmd)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (fs *TestFileSystem) Create(relPath string) error {
	isDir, absNodePath, absParentPath, err := fs.SplitPath(relPath)
	if err != nil {
		return err
	}

	os.MkdirAll(absParentPath, 0770)
	if isDir {
		err := os.Mkdir(absNodePath, 0770)
		if err != nil {
			return err
		}
	} else {
		fw, err := os.Create(absNodePath)
		if err != nil {
			return err
		}
		defer fw.Close()

		_, err = fw.WriteString(RandString(fs.fileSize))
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *TestFileSystem) Modify(relPath string) error {
	isDir, _, _, err := fs.SplitPath(relPath)
	if err != nil {
		return err
	}

	if isDir {
		return fs.ModifyTime(relPath)
	} else {
		if err := fs.ModifySize(relPath); err != nil {
			return err
		}
		return fs.ModifyTime(relPath)
	}
}

func (fs *TestFileSystem) ModifyTime(relPath string) error {
	_, absNodePath, _, err := fs.SplitPath(relPath)
	if err != nil {
		return err
	}
	if err = fs.CheckExists(absNodePath); err != nil {
		return err
	}

	fi, _ := os.Stat(absNodePath)
	newModTime := fi.ModTime().Add(time.Second)

	return os.Chtimes(absNodePath, newModTime, newModTime)
}

func (fs *TestFileSystem) ModifySize(relPath string) error {
	isDir, absNodePath, _, err := fs.SplitPath(relPath)
	if err != nil {
		return err
	}
	if err = fs.CheckExists(absNodePath); err != nil {
		return err
	}

	if isDir {
		return fmt.Errorf("Can not change size for directory, relpath: %v", relPath)
	} else {
		fi, _ := os.Stat(absNodePath)
		oldModTime := fi.ModTime()

		fw, err := os.OpenFile(absNodePath, os.O_APPEND|os.O_WRONLY, 0600)
		if err == nil {
			_, err = fw.WriteString(RandString(10))
			fw.Close()
			if err == nil {
				err = os.Chtimes(absNodePath, oldModTime, oldModTime)
			}
		}
		return err
	}
}

func (fs *TestFileSystem) Delete(relPath string) error {
	_, absNodePath, _, err := fs.SplitPath(relPath)
	if err != nil {
		return err
	}
	if err = fs.CheckExists(absNodePath); err != nil {
		return err
	}
	if base.IsPathInBasePath(TmpDir(), fs.basePath) && base.IsPathInBasePath(fs.basePath, absNodePath) {
		os.RemoveAll(absNodePath)
	} else {
		return fmt.Errorf("Could not remove files from non-tmp path")
	}

	return nil
}

func (fs *TestFileSystem) CheckExists(absNodePath string) error {
	_, err := os.Stat(absNodePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("Path for modify not exists, abspath: %v", absNodePath)
	}
	return nil
}

func (fs *TestFileSystem) SplitPath(relPath string) (isDir bool, absNodePath string, absParentPath string, err error) {
	relPath = filepath.FromSlash(relPath)
	nodeName := filepath.Base(relPath)
	if nodeName == "" {
		err = fmt.Errorf("Relative path could not be empty: %v", relPath)
		return
	}

	if !regexp.MustCompile(`.+\..+`).MatchString(nodeName) {
		isDir = true
	}

	absParentPath = filepath.Join(fs.dataPath, filepath.Dir(relPath))
	absNodePath = filepath.Join(absParentPath, nodeName)

	// fmt.Printf("%v, %v, %v\n", relPath, absParentPath)
	return
}

func (fs *TestFileSystem) BasePath() string {
	return fs.basePath
}

func (fs *TestFileSystem) DataPath() string {
	return fs.dataPath
}

func (fs *TestFileSystem) RestorePath() string {
	return fs.restorePath
}

func (fs *TestFileSystem) StoragePath() string {
	return fs.storagePath
}

func (fs *TestFileSystem) AppPath() string {
	return fs.appPath
}

func (fs *TestFileSystem) SetFileSize(fileSize int) {
	if fileSize <= 0 {
		panic("File size must be positive integer")
	}
	fs.fileSize = fileSize
}
