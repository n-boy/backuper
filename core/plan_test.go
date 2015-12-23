package core_test

import (
	"github.com/n-boy/backuper/core"
	"github.com/n-boy/backuper/storage"

	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type FilesysTestCase struct {
	create      []string
	modify      []string
	modify_size []string
	modify_time []string
	result      []string
}

func TestGetProcessNodes(t *testing.T) {
	testTmpDir := createTestTmpDir()
	plan := createTestPlan(testTmpDir)

	testCases := make([]FilesysTestCase)

	//init filesystem
	append(testCases, FilesysTestCase{
		create: []string{
			"dir1/file1.txt",
			"dir1/file2.txt",
			"dir1/dir2/file3.txt",
			"dir3",
		},
		result: []string{
			"dir1",
			"dir1/file1.txt",
			"dir1/file2.txt",
			"dir1/dir2",
			"dir1/dir2/file3.txt",
			"dir3",
		},
	})

	//create file
	append(testCases, FilesysTestCase{
		create: []string{
			"dir3/file4.txt",
		},
		result: []string{
			"dir3/file4.txt",
		},
	})

	//create dir
	append(testCases, FilesysTestCase{
		create: []string{
			"dir3/dir4",
		},
		result: []string{
			"dir3/dir4",
		},
	})

	//modify files
	append(testCases, FilesysTestCase{
		modify: []string{
			"dir1/file1.txt",
			"dir1/file2.txt",
		},
		result: []string{
			"dir1/file1.txt",
			"dir1/file2.txt",
		},
	})

	//modify_time for file & dir
	append(testCases, FilesysTestCase{
		modify_time: []string{
			"dir1/file1.txt",
			"dir1",
		},
		result: []string{
			"dir1/file1.txt",
		},
	})

	//modify_size for file
	append(testCases, FilesysTestCase{
		modify_size: []string{
			"dir1/file2.txt",
		},
		result: []string{
			"dir1/file2.txt",
		},
	})

	//modify nothing
	append(testCases, FilesysTestCase{
		result: []string{},
	})

}

func createTestPlan(testTmpDir string) core.BackupPlan {
	var err error
	if _, err = os.Stat(testTmpDir); os.IsNotExist(err) {
		panic("testTmpDir is not exists: " + testTmpDir)
	}

	var plan core.BackupPlan
	plan.Name = "test_plan_" + randString(20)
	plan.ChunkSize = core.DefaultChunkSizeMB * 1024 * 1024
	append(plan.NodesToArchive, filepath.Join(testTmpDir, "_data"))

	testStorageConfig := map[string]string{
		"path": filepath.Join(testTmpDir, "_storage"),
		"type": "localfs"}
	if plan.Storage, err = storage.NewStorage(testStorageConfig); err != nil {
		panic(err)
	}

	if err = plan.SavePlan(false); err != nil {
		panic(err)
	}

	return plan
}

func createTestTmpDir() string {
	dirPath = filepath.Join(tmpDir(), randString(20))
	os.Mkdir(dirPath, 0770)
	for i, subDir := range []string{"_storage", "_data"} {
		os.Mkdir(filepath.Join(dirPath, subDir), 0770)
	}
	return dirPath
}

func removeTestTmpDir(dirPath string) {
	if isPathInBasePath(tmpDir(), dirPath) {
		os.RemoveAll(dirPath)
	} else {
		panic("Could not remove files from non-tmp path")
	}
}

func tmpDir() string {
	varNames := []string{"TMPDIR", "TMP", "TEMP"}
	val := ""
	for _, v := range varNames {
		val = os.GetEnv(v)
		if val != "" {
			break
		}
	}
	if val == "" {
		panic("No tmp path determined")
	}
	return val
}

func isPathInBasePath(basePath, path string) bool {
	r, err := filepath.Rel(basePath, path)
	if err == nil && strings.Split(r, string(filepath.Separator))[0] != ".." {
		return true
	}
	return false
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
