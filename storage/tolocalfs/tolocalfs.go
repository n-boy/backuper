package tolocalfs

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/n-boy/backuper/base"
	storageutils "github.com/n-boy/backuper/storage/utils"
)

type LocalFSStorage struct {
	path string `name:"path" title:"Storage path"`
}

type LocalFSFileInfo struct {
	filename string
}

func NewStorage(config map[string]string) (LocalFSStorage, error) {
	var ls LocalFSStorage

	ls.path = config["path"]

	return ls, nil
}

func GetEmptyStorage() LocalFSStorage {
	return LocalFSStorage{}
}

func (ls LocalFSStorage) GetType() string {
	return "localfs"
}

func (ls LocalFSStorage) GetStorageConfig() map[string]string {
	config := make(map[string]string)

	config["path"] = ls.path

	return config
}

func (ls LocalFSStorage) UploadFile(filePath string, remoteFileName string) (map[string]string, error) {
	result := make(map[string]string)

	fileReader, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer fileReader.Close()

	filename := filepath.Base(filePath)
	if remoteFileName != "" {
		filename = filepath.Base(remoteFileName)
	}
	remoteFilepath := filepath.Join(ls.path, filename)
	remoteFilepathShadow := remoteFilepath + "~"
	fileWriter, err := os.OpenFile(remoteFilepathShadow, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0660)
	if err != nil {
		return result, err
	}
	defer fileWriter.Close()

	_, err = io.Copy(fileWriter, fileReader)
	if err != nil {
		return result, err
	}
	if fileWriter.Close(); err != nil {
		return result, err
	}

	if err = os.Rename(remoteFilepathShadow, remoteFilepath); err != nil {
		return result, err
	}
	result["filename"] = filename

	return result, nil
}

func (ls LocalFSStorage) DownloadFile(fileStorageId map[string]string, localFilePath string) error {
	downloadAction := func(pipe io.Writer) error {
		return ls.DownloadFileToPipe(fileStorageId, pipe)
	}
	return storageutils.DownloadFile(downloadAction, localFilePath)
}

func (ls LocalFSStorage) DownloadFileToPipe(fileStorageId map[string]string, pipe io.Writer) error {
	fileReader, err := os.Open(filepath.Join(ls.path, fileStorageId["filename"]))
	if err != nil {
		return err
	}
	defer fileReader.Close()

	_, err = io.Copy(pipe, fileReader)

	return err
}

func (ls LocalFSStorage) DeleteFile(fileStorageInfo map[string]string) error {
	return os.Remove(filepath.Join(ls.path, fileStorageInfo["filename"]))
}

func (ls LocalFSStorage) GetFilesList() ([]base.GenericStorageFileInfo, error) {
	var filesList []base.GenericStorageFileInfo

	dirNodes, err := ioutil.ReadDir(ls.path)
	if err != nil {
		return filesList, err
	}

	for _, fi := range dirNodes {
		if !fi.IsDir() {
			filesList = append(filesList, base.GenericStorageFileInfo(LocalFSFileInfo{filename: fi.Name()}))
		}
	}

	return filesList, nil
}

func (lfi LocalFSFileInfo) GetFilename() string {
	return lfi.filename
}

func (lfi LocalFSFileInfo) GetFileStorageId() map[string]string {
	return map[string]string{"filename": lfi.filename}
}
