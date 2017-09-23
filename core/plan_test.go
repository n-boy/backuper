package core_test

import (
	"github.com/n-boy/backuper/base"
	"github.com/n-boy/backuper/core"
	"github.com/n-boy/backuper/storage"

	"github.com/n-boy/backuper/ut/testutils"

	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"testing"
	"time"
)

type FilesysTestCase struct {
	name              string
	skip_on_platforms []string
	cmds_to_apply     map[string][]string
	result            []string
}

var TestCasesGetProcessNodes []FilesysTestCase = []FilesysTestCase{
	{
		name: "init filesystem",
		cmds_to_apply: map[string][]string{
			"create": {
				"dir1/file1.txt",
				"dir1/file2.txt",
				"dir1/dir2/file3.txt",
				"dir3",
			},
		},
		result: []string{
			".",
			"dir1",
			"dir1/file1.txt",
			"dir1/file2.txt",
			"dir1/dir2",
			"dir1/dir2/file3.txt",
			"dir3",
		},
	},

	{
		name: "create file",
		cmds_to_apply: map[string][]string{
			"create": {
				"dir3/file4.txt",
			},
		},
		result: []string{
			"dir3/file4.txt",
		},
	},

	{
		name: "create dir",
		cmds_to_apply: map[string][]string{
			"create": {
				"dir3/dir4",
			},
		},
		result: []string{
			"dir3/dir4",
		},
	},

	{
		name: "modify files",
		cmds_to_apply: map[string][]string{
			"modify": {
				"dir1/file1.txt",
				"dir1/file2.txt",
			},
		},
		result: []string{
			"dir1/file1.txt",
			"dir1/file2.txt",
		},
	},

	{
		name: "modify_time for file & dir",
		cmds_to_apply: map[string][]string{
			"modify_time": {
				"dir1/file1.txt",
				"dir1",
			},
		},
		result: []string{
			"dir1/file1.txt",
		},
	},

	{
		name: "modify_size for file",
		cmds_to_apply: map[string][]string{
			"modify_size": {
				"dir1/file2.txt",
			},
		},
		result: []string{
			"dir1/file2.txt",
		},
	},

	{
		name:   "modify nothing",
		result: []string{},
	},
}

func TestGetProcessNodes(t *testing.T) {
	testTmpDir := createTestTmpDir()
	base.InitApp(base.AppConfig{
		AppDir:         filepath.Join(testTmpDir, "_app"),
		LogToStdout:    false,
		LogErrToStderr: false,
	})
	plan := createTestPlan(testTmpDir)
	testDataTmpDir := filepath.Join(testTmpDir, "_data")

	platform := runtime.GOOS

	for step, tc := range TestCasesGetProcessNodes {
		if err := fsApplyTestCase(testDataTmpDir, tc); err != nil {
			t.Fatalf("Test died. Step: %v, Name: %v, error: %v\n", step, tc.name, err)
		}

		guardNodes := plan.GetGuardedNodes()
		archNodesMap := plan.GetArchivedNodesMap()

		procNodes := plan.GetProcessNodes(guardNodes, archNodesMap)
		var procNodesPathes sort.StringSlice
		for _, node := range procNodes {
			relPath, _ := filepath.Rel(testDataTmpDir, node.GetNodePath())
			procNodesPathes = append(procNodesPathes, relPath)
		}

		procNodesPathes.Sort()

		correctResult := make(sort.StringSlice, 0)
		for _, item := range tc.result {
			correctResult = append(correctResult, filepath.FromSlash(item))
		}
		correctResult.Sort()

		skip := false
		for _, p := range tc.skip_on_platforms {
			if p == platform {
				skip = true
				continue
			}
		}
		if skip {
			t.Logf("Test skipped. Step: %v, Name: %v\n", step, tc.name)
		} else {
			if fmt.Sprintf("%v", procNodesPathes) == fmt.Sprintf("%v", correctResult) {
				t.Logf("Test passed. Step: %v, Name: %v\n", step, tc.name)
			} else {
				t.Errorf("Test failed. Step: %v, Name: %v, expected: %v, got: %v\n", step, tc.name, correctResult, procNodesPathes)
			}
		}

		if err := plan.DoBackup(); err != nil {
			t.Fatalf("Test died. Step: %v, Name: %v, error: %v\n", step, tc.name, err)
		}
	}
	removeTestTmpDir(testTmpDir)
}

func fsApplyTestCase(basePath string, tc FilesysTestCase) error {
	for cmd, pathes := range tc.cmds_to_apply {
		for _, relPath := range pathes {
			var err error
			switch cmd {
			case "create":
				err = fsCreate(relPath, basePath)
			case "modify":
				err = fsModify(relPath, basePath)
			case "modify_time":
				err = fsModifyTime(relPath, basePath)
			case "modify_size":
				err = fsModifySize(relPath, basePath)
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

func fsCreate(relPath, basePath string) error {
	isDir, absNodePath, absParentPath, err := fsSplitPath(relPath, basePath)
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

		_, err = fw.WriteString(testutils.RandString(10))
		if err != nil {
			return err
		}
	}
	return nil
}

func fsModify(relPath, basePath string) error {
	isDir, _, _, err := fsSplitPath(relPath, basePath)
	if err != nil {
		return err
	}

	if isDir {
		return fsModifyTime(relPath, basePath)
	} else {
		if err := fsModifySize(relPath, basePath); err != nil {
			return err
		}
		return fsModifyTime(relPath, basePath)
	}
}

func fsModifyTime(relPath, basePath string) error {
	_, absNodePath, _, err := fsSplitPath(relPath, basePath)
	if err != nil {
		return err
	}
	if err = fsCheckExists(absNodePath); err != nil {
		return err
	}

	fi, _ := os.Stat(absNodePath)
	newModTime := fi.ModTime().Add(time.Second)

	return os.Chtimes(absNodePath, newModTime, newModTime)
}

func fsModifySize(relPath, basePath string) error {
	isDir, absNodePath, _, err := fsSplitPath(relPath, basePath)
	if err != nil {
		return err
	}
	if err = fsCheckExists(absNodePath); err != nil {
		return err
	}

	if isDir {
		return fmt.Errorf("Can not change size for directory, relpath: %v", relPath)
	} else {
		fi, _ := os.Stat(absNodePath)
		oldModTime := fi.ModTime()

		fw, err := os.OpenFile(absNodePath, os.O_APPEND|os.O_WRONLY, 0600)
		if err == nil {
			_, err = fw.WriteString(testutils.RandString(10))
			fw.Close()
			if err == nil {
				err = os.Chtimes(absNodePath, oldModTime, oldModTime)
			}
		}
		return err
	}
}

func fsCheckExists(absNodePath string) error {
	_, err := os.Stat(absNodePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("Path for modify not exists, abspath: %v", absNodePath)
	}
	return nil
}

func fsSplitPath(relPath, basePath string) (isDir bool, absNodePath string, absParentPath string, err error) {
	relPath = filepath.FromSlash(relPath)
	nodeName := filepath.Base(relPath)
	if nodeName == "" {
		err = fmt.Errorf("Relative path could not be empty: %v", relPath)
		return
	}

	if !regexp.MustCompile(`.+\..+`).MatchString(nodeName) {
		isDir = true
	}

	absParentPath = filepath.Join(basePath, filepath.Dir(relPath))
	absNodePath = filepath.Join(absParentPath, nodeName)

	// fmt.Printf("%v, %v, %v\n", relPath, absParentPath)
	return
}

func createTestPlan(testTmpDir string) core.BackupPlan {
	var err error
	if _, err = os.Stat(testTmpDir); os.IsNotExist(err) {
		panic("testTmpDir is not exists: " + testTmpDir)
	}

	var plan core.BackupPlan
	plan.Name = "test_plan_" + testutils.RandString(20)
	plan.ChunkSize = core.DefaultChunkSizeMB * 1024 * 1024
	plan.NodesToArchive = append(plan.NodesToArchive, filepath.Join(testTmpDir, "_data"))

	testStorageConfig := map[string]string{
		"path": filepath.Join(testTmpDir, "_storage"),
		"type": "localfs"}
	if plan.Storage, err = storage.NewStorage(testStorageConfig); err != nil {
		panic(err)
	}

	if err = plan.SavePlan(false); err != nil {
		panic(err)
	}

	plan, err = core.GetBackupPlan(plan.Name)
	if err != nil {
		panic(err)
	}

	return plan
}

func createTestTmpDir() string {
	dirPath := filepath.Join(testutils.TmpDir(), testutils.RandString(20))
	err := os.Mkdir(dirPath, 0770)
	if err == nil {
		for _, subDir := range []string{"_storage", "_data", "_app"} {
			err = os.Mkdir(filepath.Join(dirPath, subDir), 0770)
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		panic(err)
	}
	return dirPath
}

func removeTestTmpDir(dirPath string) {
	if base.IsPathInBasePath(testutils.TmpDir(), dirPath) {
		os.RemoveAll(dirPath)
	} else {
		panic("Could not remove files from non-tmp path")
	}
}
