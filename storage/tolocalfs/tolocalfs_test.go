// +build integration storage localfs

package tolocalfs_test

import (
	"github.com/n-boy/backuper/storage/tolocalfs"

	"github.com/n-boy/backuper/ut/teststorage"
	"github.com/n-boy/backuper/ut/testutils"

	"testing"
)

const normalFileSize = 3 * 1024 * 1024
const smallFileSize = 1 * 1024 * 1024

func TestUploadDownloadFile(t *testing.T) {
	s, err := getStorage()
	if err != nil {
		t.Fatalf(err.Error())
	}
	teststorage.CheckUploadDownloadFile(t, s, normalFileSize)
}

func TestGetFilesList(t *testing.T) {
	s, err := getStorage()
	if err != nil {
		t.Fatalf(err.Error())
	}
	teststorage.CheckGetFilesList(t, s, smallFileSize, 0)
}

func TestDeleteFile(t *testing.T) {
	s, err := getStorage()
	if err != nil {
		t.Fatalf(err.Error())
	}
	teststorage.CheckDeleteFile(t, s, smallFileSize, true)
}

func getStorage() (tolocalfs.LocalFSStorage, error) {
	testsConfig, err := testutils.GetTestsConfig()
	if err != nil {
		return tolocalfs.LocalFSStorage{}, err
	}

	return tolocalfs.NewStorage(testsConfig.StorageLocalFS)
}
