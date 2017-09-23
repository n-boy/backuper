// +build integration storage glacier

package toglacier_test

import (
	"github.com/n-boy/backuper/storage/toglacier"

	"github.com/n-boy/backuper/ut/teststorage"
	"github.com/n-boy/backuper/ut/testutils"

	"testing"
)

const normalFileSize = 15 * 1024 * 1024
const smallFileSize = 1 * 1024 * 1024
const waitForActualFilesListSeconds = 24 * 60 * 60

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
	teststorage.CheckGetFilesList(t, s, smallFileSize, waitForActualFilesListSeconds)
}

func TestDeleteFile(t *testing.T) {
	s, err := getStorage()
	if err != nil {
		t.Fatalf(err.Error())
	}
	teststorage.CheckDeleteFile(t, s, smallFileSize, false)
}

func getStorage() (toglacier.GlacierStorage, error) {
	testsConfig, err := testutils.GetTestsConfig()
	if err != nil {
		return toglacier.GlacierStorage{}, err
	}

	return toglacier.NewStorage(testsConfig.StorageGlacier)
}
