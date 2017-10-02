package utils

import (
	"io"
	"os"
)

func DownloadFile(downloadAction func(pipe io.Writer) error, localFilePath string) error {
	localFilePathShadow := localFilePath + "~"
	fileWriter, err := os.OpenFile(localFilePathShadow, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0660)
	if err != nil {
		return err
	}
	defer fileWriter.Close()

	err = downloadAction(io.Writer(fileWriter))
	if err != nil {
		fileWriter.Close()
		os.Remove(localFilePathShadow)
		return err
	}
	if fileWriter.Close(); err != nil {
		return err
	}

	return os.Rename(localFilePathShadow, localFilePath)
}
