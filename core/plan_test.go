package core_test

import (
	"github.com/n-boy/backuper/base"
	"github.com/n-boy/backuper/core"

	"github.com/n-boy/backuper/ut/testutils"

	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

type FilesysTestCase struct {
	name              string
	skip_on_platforms []string
	cmds_to_apply     testutils.CmdsToApply
	result            []string
}

var TestCasesGetProcessNodes []FilesysTestCase = []FilesysTestCase{
	{
		name: "init filesystem",
		cmds_to_apply: testutils.CmdsToApply{
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
		cmds_to_apply: testutils.CmdsToApply{
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
		cmds_to_apply: testutils.CmdsToApply{
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
		cmds_to_apply: testutils.CmdsToApply{
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
		cmds_to_apply: testutils.CmdsToApply{
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
		cmds_to_apply: testutils.CmdsToApply{
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
	tfs := testutils.CreateTestFileSystem()
	defer tfs.Destroy()

	base.InitApp(base.AppConfig{
		AppDir:         tfs.AppPath(),
		LogToStdout:    false,
		LogErrToStderr: false,
	})
	plan := testutils.CreateTestPlan(tfs, core.DefaultChunkSizeMB*1024*1024)

	platform := runtime.GOOS

	for step, tc := range TestCasesGetProcessNodes {
		if err := tfs.ApplyCmds(tc.cmds_to_apply); err != nil {
			t.Fatalf("Test died. Step: %v, Name: %v, error: %v\n", step, tc.name, err)
		}

		guardNodes := plan.GetGuardedNodes()
		archNodesMap := plan.GetArchivedNodesMap()

		procNodes := plan.GetProcessNodes(guardNodes, archNodesMap)
		var procNodesPathes sort.StringSlice
		for _, node := range procNodes {
			relPath, _ := filepath.Rel(tfs.DataPath(), node.GetNodePath())
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
}
