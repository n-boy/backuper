package storage

import (
	"github.com/n-boy/backuper/base"
	"github.com/n-boy/backuper/storage/toglacier"
	"github.com/n-boy/backuper/storage/tolocalfs"

	"fmt"
	"io"
	"reflect"
)

type GenericStorage interface {
	GetStorageConfig() map[string]string
	UploadFile(filePath string, remoteFileName string) (map[string]string, error)
	DownloadFile(fileStorageId map[string]string, localFilePath string) error
	DownloadFileToPipe(fileStorageId map[string]string, pipe io.Writer) error
	DeleteFile(fileStorageInfo map[string]string) error
	GetFilesList() ([]base.GenericStorageFileInfo, error)
	GetType() string
}

var storageTypes = []string{"glacier", "localfs"}

func GetStorageTypes() []string {
	return storageTypes
}

func GetStorageTypesMap() map[string]bool {
	typesMap := make(map[string]bool)
	for _, stype := range storageTypes {
		typesMap[stype] = true
	}
	return typesMap
}

func NewStorage(config map[string]string) (GenericStorage, error) {
	stype, stype_exists := config["type"]
	if !stype_exists {
		return nil, fmt.Errorf("Storage type is not defined")
	} else if !GetStorageTypesMap()[stype] {
		return nil, fmt.Errorf("Storage type is not supported")
	}

	switch stype {
	case "glacier":
		return toglacier.NewStorage(config)
	case "localfs":
		return tolocalfs.NewStorage(config)
	}
	return nil, fmt.Errorf("Storage type is not supported: %v", stype)
}

type ConfigField struct {
	Name  string
	Title string
}

func GetStorageConfigFields(stype string) (fields []ConfigField, err error) {
	var emptyStorage GenericStorage
	if emptyStorage, err = GetEmptyStorage(stype); err != nil {
		return
	}

	reflect_val := reflect.ValueOf(emptyStorage)
	// reflect_val := reflect.New(reflect.Value(emptyStorage)).Elem()
	for i := 0; i < reflect_val.NumField(); i++ {
		tag := reflect_val.Type().Field(i).Tag

		if tag.Get("name") != "" {
			cf := ConfigField{Name: tag.Get("name")}
			if tag.Get("title") != "" {
				cf.Title = tag.Get("title")
			} else {
				cf.Title = cf.Name
			}

			fields = append(fields, cf)
		}
	}

	return
}

func GetEmptyStorage(stype string) (es GenericStorage, err error) {
	switch stype {
	case "glacier":
		return toglacier.GetEmptyStorage(), nil
	case "localfs":
		return tolocalfs.GetEmptyStorage(), nil
	}

	return nil, fmt.Errorf("Storage type is not supported")
}
