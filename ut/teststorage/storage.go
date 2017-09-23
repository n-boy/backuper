// real tests are in storage-type dependent packages
// here is generic methods that helps run tests on defined GenericStorage

package teststorage

import (
	"github.com/n-boy/backuper/base"
	"github.com/n-boy/backuper/storage"

	"github.com/n-boy/backuper/ut/testutils"

	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var timeoutError = errors.New("waitRequestInProgress timeout reached")

func CheckUploadDownloadFile(t *testing.T, s storage.GenericStorage, fileSize int) {
	testutils.InitAppForTests()
	testName := "UploadDownloadFile"

	var sourceFilePath string
	var err error
	sourceFilePath, err = createSourceFile(t, fileSize)
	if err != nil {
		t.Fatalf("Test died. Step: CreateSourceFile, Name: %v, error: %v\n", testName, err)
	}
	defer deleteTempFile(t, sourceFilePath)

	var fileStorageInfo map[string]string
	fileStorageInfo, err = s.UploadFile(sourceFilePath, nil, filepath.Base(sourceFilePath))
	if err != nil {
		t.Fatalf("Test died. Step: UploadFile, Name: %v, error: %v\n", testName, err)
	} else {
		t.Logf("Successful upload file to storage: %v", sourceFilePath)
	}
	defer s.DeleteFile(fileStorageInfo)

	var sourceMD5 string
	sourceMD5, err = testutils.CalcFileMD5(sourceFilePath)
	if err != nil {
		t.Fatalf("Test died. Step: CalcSourceFileMD5, Name: %v, error: %v\n", testName, err)
	} else {
		t.Logf("Source file md5 calculated: %v", sourceMD5)
	}

	deleteTempFile(t, sourceFilePath)

	var restoredFilePath = filepath.Join(testutils.TmpDir(), testutils.RandString(20)+".bin")
	err = downloadFile(t, s, fileStorageInfo, restoredFilePath)
	if err != nil {
		t.Fatalf("Test died. Step: DownloadFile, Name: %v, error: %v\n", testName, err)
	} else {
		t.Logf("Restored file downloaded: %v", restoredFilePath)
	}
	defer deleteTempFile(t, restoredFilePath)

	var restoredMD5 string
	restoredMD5, err = testutils.CalcFileMD5(restoredFilePath)
	if err != nil {
		t.Fatalf("Test died. Step: CalcRestoredFileMD5, Name: %v, error: %v\n", testName, err)
	} else {
		t.Logf("Restored file md5 calculated: %v", restoredMD5)
	}

	if sourceMD5 != restoredMD5 {
		t.Errorf("Test failed. Step: CompareMD5, Name: %v, expected: %v, got: %v\n", testName, sourceMD5, restoredMD5)
	} else {
		t.Logf("Test passed. Step: CompareMD5, Name: %v", testName)
	}
}

func CheckGetFilesList(t *testing.T, s storage.GenericStorage, fileSize int, waitForActualListSeconds int64) {
	testutils.InitAppForTests()
	testName := "GetFilesList"

	var sourceFilePath string
	var err error
	sourceFilePath, err = createSourceFile(t, fileSize)
	if err != nil {
		t.Fatalf("Test died. Step: CreateSourceFile, Name: %v, error: %v\n", testName, err)
	}
	defer deleteTempFile(t, sourceFilePath)

	var fileStorageInfo map[string]string
	fileStorageInfo, err = s.UploadFile(sourceFilePath, nil, filepath.Base(sourceFilePath))
	if err != nil {
		t.Fatalf("Test died. Step: UploadFile, Name: %v, error: %v\n", testName, err)
	} else {
		t.Logf("Successful upload file to storage: %v", sourceFilePath)
	}
	defer s.DeleteFile(fileStorageInfo)

	// in some storages like aws glacier, files list does not refreshed immediately
	fmt.Printf("Waiting for files list actuality, seconds: %v, now: %v\n", waitForActualListSeconds, time.Now())
	time.Sleep(time.Duration(waitForActualListSeconds) * time.Second)

	var filesList []base.GenericStorageFileInfo
	filesList, err = getFilesList(t, s)
	if err != nil {
		t.Fatalf("Test died. Step: GetFilesList, Name: %v, error: %v\n", testName, err)
	} else {
		t.Logf("Files list successfully downloaded")
	}

	sourceFileName := filepath.Base(sourceFilePath)
	var founded bool
	for _, f := range filesList {
		if f.GetFilename() == sourceFileName {
			founded = true
			break
		}
	}
	if founded {
		t.Logf("Test passed. Step: FindSourceFileInList, Name: %v", testName)
	} else {
		t.Errorf("Test failed. Step: FindSourceFileInList, Name: %v, expected: source file founded in list, got: not founded\n", testName)
	}
}

func CheckDeleteFile(t *testing.T, s storage.GenericStorage, fileSize int, useFilesList bool) {
	testutils.InitAppForTests()
	testName := "DeleteFile"

	var sourceFilePath string
	var err error
	sourceFilePath, err = createSourceFile(t, fileSize)
	if err != nil {
		t.Fatalf("Test died. Step: CreateSourceFile, Name: %v, error: %v\n", testName, err)
	}
	defer deleteTempFile(t, sourceFilePath)

	var fileStorageInfo map[string]string
	fileStorageInfo, err = s.UploadFile(sourceFilePath, nil, filepath.Base(sourceFilePath))
	if err != nil {
		t.Fatalf("Test died. Step: UploadFile, Name: %v, error: %v\n", testName, err)
	} else {
		t.Logf("Successful upload file to storage: %v", sourceFilePath)
	}

	sourceFileName := filepath.Base(sourceFilePath)
	err = s.DeleteFile(fileStorageInfo)
	if err != nil {
		t.Fatalf("Test died. Step: DeleteFile, Name: %v, error: %v\n", testName, err)
	} else {
		t.Logf("File successfully deleted from storage: %v", sourceFileName)
	}

	if useFilesList {
		var filesList []base.GenericStorageFileInfo
		filesList, err = getFilesList(t, s)
		if err != nil {
			t.Fatalf("Test failed. Step: GetFilesList, Name: %v, error: %v\n", testName, err)
		} else {
			t.Logf("Files list successfully downloaded")
		}

		var founded bool
		for _, f := range filesList {
			if f.GetFilename() == sourceFileName {
				founded = true
				break
			}
		}
		if founded {
			t.Errorf("Test failed. Step: FindDeletedFileInList, Name: %v, expected: source file NOT founded in list, got: founded\n", testName)
		} else {
			t.Logf("Test passed. Step: FindDeletedFileInList, Name: %v", testName)
		}
	} else {
		var restoredFilePath = filepath.Join(testutils.TmpDir(), testutils.RandString(20)+".bin")
		err = downloadFile(t, s, fileStorageInfo, restoredFilePath)
		if err == nil {
			t.Errorf("Test failed. Step: DownloadFile, Name: %v, expected: source file NOT downloaded, got: downloaded\n", testName)
		} else {
			t.Logf("TestPassed. Step: DownloadFile, Name: %v, error downloading deleted file: %v ", restoredFilePath, err)
		}
		defer deleteTempFile(t, restoredFilePath)
	}
}

func createSourceFile(t *testing.T, fileSize int) (string, error) {
	filePath := filepath.Join(testutils.TmpDir(), testutils.RandString(20)+".bin")
	t.Logf("Creating temporary file %s of size %d bytes", filePath, fileSize)
	fw, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return filePath, err
	}
	defer fw.Close()

	bytesWritten := 0
	for bytesWritten < fileSize {
		chunkSize := 1024
		if fileSize-bytesWritten < chunkSize {
			chunkSize = fileSize - bytesWritten
		}

		var n int
		n, err = fw.WriteString(testutils.RandString(chunkSize))
		if err != nil {
			return filePath, err
		}
		bytesWritten += n
	}

	return filePath, nil
}

func deleteTempFile(t *testing.T, filePath string) error {
	if base.IsPathInBasePath(testutils.TmpDir(), filePath) {
		t.Logf("Deleting temporary file %s", filePath)
		return os.Remove(filePath)
	} else {
		return fmt.Errorf("Could not remove files from non-tmp path: %s", filePath)
	}
}

func downloadFile(t *testing.T, s storage.GenericStorage, fileStorageInfo map[string]string, restoredFilePath string) error {
	action := func() error {
		return s.DownloadFile(fileStorageInfo, restoredFilePath, nil)
	}
	return waitRequestInProgress(t, action, base.StorageRequestInProgressRetrySeconds, 0)
}

func getFilesList(t *testing.T, s storage.GenericStorage) ([]base.GenericStorageFileInfo, error) {
	filesList := []base.GenericStorageFileInfo{}
	action := func() error {
		var err error
		filesList, err = s.GetFilesList()
		return err
	}
	err := waitRequestInProgress(t, action, base.StorageRequestInProgressRetrySeconds, 0)
	return filesList, err
}

func waitRequestInProgress(t *testing.T, action func() error, iterationPauseSeconds int64, timeout int64) error {
	timeStart := time.Now()
	for {
		err := action()
		if err != nil {
			if err == base.ErrStorageRequestInProgress {
				fmt.Printf("Request to storage is in progress. Waiting for %v seconds...\n", iterationPauseSeconds)
				elapsedSeconds := time.Since(timeStart).Seconds()
				if timeout > 0 && elapsedSeconds+float64(iterationPauseSeconds) > float64(timeout) {
					return timeoutError
				}
				time.Sleep(time.Duration(iterationPauseSeconds) * time.Second)
			} else {
				return err
			}
		} else {
			return nil
		}
	}
}
