// functional tests on backup+restore process generally

package core_test

import (
	"github.com/n-boy/backuper/base"
	"github.com/n-boy/backuper/core"

	"github.com/n-boy/backuper/ut/testutils"

	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

var fileSize = 400 * 1024
var chunkSize int64 = 1 * 1024 * 1024

// one iteration, all files feat to one archive
func TestBROnePoint(t *testing.T) {
	tfs, plan := InitTfsAndPlan(t)
	defer tfs.Destroy()

	err := tfs.ApplyCmds(testutils.CmdsToApply{
		"create": {
			"dir1/file1.txt",
			"dir1/file2.txt",
			"dir3",
		},
	})
	if err != nil {
		t.Fatalf("Test died. Error while initializing filesystem: %v\n", err)
	}

	err = plan.DoBackup()
	if err != nil {
		t.Fatalf("Test died. Error while backuping files: %v\n", err)
	}

	points := plan.GetRestorePoints([]string{tfs.DataPath()})
	if len(points) == 0 {
		t.Fatalf("Test died. No restore points founded\n")
	}
	if len(points) != 1 {
		t.Errorf("Test failed. Qty of archives (restore points) in storage not as expected: got %v, expected %v\n",
			len(points), 1)
	}

	err = plan.InitRestore([]string{tfs.DataPath()}, &points[0], tfs.RestorePath())
	if err != nil {
		t.Fatalf("Test died. Error while initializing restore: %v\n", err)
	}

	err = plan.DoRestore()
	if err != nil {
		t.Fatalf("Test died. Error while restoring files: %v\n", err)
	}

	dataPathAfterRestore := filepath.Join(tfs.RestorePath(), core.GetPathInArchive(tfs.DataPath()))
	cmpRes, err1 := testutils.CompareDirs(tfs.DataPath(), dataPathAfterRestore)
	if err1 != nil {
		t.Fatalf("Test died. Error while comparing directories: %v\n", err1)
	}

	if !cmpRes.Equals {
		t.Errorf("Test failed. Restored dir content differs from source dir content:\n%s", cmpRes.String())
	}

}

// two iterations, all files feat to one archive, restore folder from the first restore point
func TestBRFewPointsRestoreFirst(t *testing.T) {
	tfs, plan := InitTfsAndPlan(t)
	defer tfs.Destroy()

	err := tfs.ApplyCmds(testutils.CmdsToApply{
		"create": {
			"dir1/file1.txt",
			"dir1/file2.txt",
			"dir3",
		},
	})
	if err != nil {
		t.Fatalf("Test died. Error while initializing filesystem: %v\n", err)
	}

	err = plan.DoBackup()
	if err != nil {
		t.Fatalf("Test died. Error while backuping files: %v\n", err)
	}

	dataPathSnapshot, err1 := testutils.GetDirNodes(tfs.DataPath())
	if err1 != nil {
		t.Fatalf("Test died. Error while taking data path snapshot: %v\n", err1)
	}

	err = tfs.ApplyCmds(testutils.CmdsToApply{
		"modify": {
			"dir1/file1.txt",
		},
	})
	if err != nil {
		t.Fatalf("Test died. Error while modifying filesystem: %v\n", err)
	}

	err = plan.DoBackup()
	if err != nil {
		t.Fatalf("Test died. Error while backuping files: %v\n", err)
	}

	points := plan.GetRestorePoints([]string{tfs.DataPath()})
	if len(points) == 0 {
		t.Fatalf("Test died. No restore points founded\n")
	}
	if len(points) != 2 {
		t.Errorf("Test failed. Qty of archives (restore points) in storage not as expected: got %v, expected %v\n",
			len(points), 2)
	}
	err = plan.InitRestore([]string{tfs.DataPath()}, &points[0], tfs.RestorePath())
	if err != nil {
		t.Fatalf("Test died. Error while initializing restore: %v\n", err)
	}

	err = plan.DoRestore()
	if err != nil {
		t.Fatalf("Test died. Error while restoring files: %v\n", err)
	}

	dataPathAfterRestore := filepath.Join(tfs.RestorePath(), core.GetPathInArchive(tfs.DataPath()))
	restoredPathSnapshot, err2 := testutils.GetDirNodes(dataPathAfterRestore)
	if err2 != nil {
		t.Fatalf("Test died. Error while taking restore path snapshot: %v\n", err2)
	}

	cmpRes, err3 := testutils.CompareDirNodes(dataPathSnapshot, restoredPathSnapshot)
	if err3 != nil {
		t.Fatalf("Test died. Error while comparing directories: %v\n", err3)
	}

	if !cmpRes.Equals {
		t.Errorf("Test failed. Restored dir content differs from source dir (before first backup iteration) content:\n%s", cmpRes.String())
	}
}

// two iterations, all files NOT feat to one archive, restore folder from the last restore point
func TestBRFewPointsRestoreLast(t *testing.T) {
	tfs, plan := InitTfsAndPlan(t)
	defer tfs.Destroy()

	err := tfs.ApplyCmds(testutils.CmdsToApply{
		"create": {
			"dir1/file1.txt",
			"dir1/file2.txt",
			"dir1/file3.txt",
			"dir3",
		},
	})
	if err != nil {
		t.Fatalf("Test died. Error while initializing filesystem: %v\n", err)
	}

	err = plan.DoBackup()
	if err != nil {
		t.Fatalf("Test died. Error while backuping files: %v\n", err)
	}

	err = tfs.ApplyCmds(testutils.CmdsToApply{
		"modify": {
			"dir1/file1.txt",
		},
	})
	if err != nil {
		t.Fatalf("Test died. Error while modifying filesystem: %v\n", err)
	}

	err = plan.DoBackup()
	if err != nil {
		t.Fatalf("Test died. Error while backuping files: %v\n", err)
	}

	points := plan.GetRestorePoints([]string{tfs.DataPath()})
	if len(points) == 0 {
		t.Fatalf("Test died. No restore points founded\n")
	}
	if len(points) != 3 {
		t.Errorf("Test failed. Qty of archives (restore points) in storage not as expected: got %v, expected %v\n",
			len(points), 3)
	}
	err = plan.InitRestore([]string{tfs.DataPath()}, &points[len(points)-1], tfs.RestorePath())
	if err != nil {
		t.Fatalf("Test died. Error while initializing restore: %v\n", err)
	}

	err = plan.DoRestore()
	if err != nil {
		t.Fatalf("Test died. Error while restoring files: %v\n", err)
	}

	dataPathAfterRestore := filepath.Join(tfs.RestorePath(), core.GetPathInArchive(tfs.DataPath()))

	cmpRes, err3 := testutils.CompareDirs(tfs.DataPath(), dataPathAfterRestore)
	if err3 != nil {
		t.Fatalf("Test died. Error while comparing directories: %v\n", err3)
	}

	if !cmpRes.Equals {
		t.Errorf("Test failed. Restored dir content differs from source dir (after second backup iteration) content:\n%s", cmpRes.String())
	}
}

// one iteration, all files feat to one archive, restore folder from the last restore point, encrypted
func TestBRFewPointsRestoreLastEncrypted(t *testing.T) {
	tfs, plan := InitTfsAndPlan(t)
	defer tfs.Destroy()

	err := tfs.ApplyCmds(testutils.CmdsToApply{
		"create": {
			"dir1/file1.txt",
			"dir1/file2.txt",
			"dir3",
		},
	})
	if err != nil {
		t.Fatalf("Test died. Error while initializing filesystem: %v\n", err)
	}

	plan.Encrypt = true
	plan.Encrypt_passphrase = "encryptpassphrasefortest1"

	err = plan.DoBackup()
	if err != nil {
		t.Fatalf("Test died. Error while backuping files: %v\n", err)
	}

	// check is storage files are seem to be encrypted
	encryptedStoragePathSnapshot, err4 := testutils.GetDirNodes(tfs.StoragePath())
	if err4 != nil {
		t.Fatalf("Test died. Error while taking snapshot of storage path: %v\n", err4)
	}

	var encMetaFileName, encArchiveFileName string

	for p, _ := range encryptedStoragePathSnapshot {
		if core.GetArchiveFileNameRE().MatchString(p) {
			encArchiveFileName = p
		}
		if core.GetMetaFileNameRE().MatchString(p) {
			encMetaFileName = p
		}
	}

	_, err = core.ParseMetaFile(filepath.Join(tfs.StoragePath(), encMetaFileName))
	if err == nil {
		t.Errorf("Test failed. Encrypted metafile parsed without decryption\n")
	}

	_, err = zip.OpenReader(filepath.Join(tfs.StoragePath(), encArchiveFileName))
	if err == nil {
		t.Errorf("Test failed. Encrypted archive opened by unzipper without decryption\n")
	}
	// ---

	err = tfs.ApplyCmds(testutils.CmdsToApply{
		"modify": {
			"dir1/file1.txt",
		},
	})
	if err != nil {
		t.Fatalf("Test died. Error while modifying filesystem: %v\n", err)
	}

	err = plan.DoBackup()
	if err != nil {
		t.Fatalf("Test died. Error while backuping files: %v\n", err)
	}

	points := plan.GetRestorePoints([]string{tfs.DataPath()})
	if len(points) == 0 {
		t.Fatalf("Test died. No restore points founded\n")
	}
	if len(points) != 2 {
		t.Errorf("Test failed. Qty of archives (restore points) in storage not as expected: got %v, expected %v\n",
			len(points), 2)
	}
	err = plan.InitRestore([]string{tfs.DataPath()}, &points[len(points)-1], tfs.RestorePath())
	if err != nil {
		t.Fatalf("Test died. Error while initializing restore: %v\n", err)
	}

	err = plan.DoRestore()
	if err != nil {
		t.Fatalf("Test died. Error while restoring files: %v\n", err)
	}

	dataPathAfterRestore := filepath.Join(tfs.RestorePath(), core.GetPathInArchive(tfs.DataPath()))

	cmpRes, err3 := testutils.CompareDirs(tfs.DataPath(), dataPathAfterRestore)
	if err3 != nil {
		t.Fatalf("Test died. Error while comparing directories: %v\n", err3)
	}

	if !cmpRes.Equals {
		t.Errorf("Test failed. Restored dir content differs from source dir (after second backup iteration) content:\n%s", cmpRes.String())
	}
}

// one iteration, all files feat to one archive, two different backup dirs
func TestBROnePointFewBackupPathes(t *testing.T) {
	tfs, plan := InitTfsAndPlan(t)
	defer tfs.Destroy()

	err := tfs.ApplyCmds(testutils.CmdsToApply{
		"create": {
			"dir1/file1.txt",
			"dir3/file3.txt",
			"dir3/dir2",
		},
	})
	if err != nil {
		t.Fatalf("Test died. Error while initializing filesystem: %v\n", err)
	}

	plan.NodesToArchive = []string{filepath.Join(tfs.DataPath(), "dir1"), filepath.Join(tfs.DataPath(), "dir3")}
	if err = plan.SavePlan(true); err != nil {
		t.Fatalf("Test died. Error while saving plan: %v", err)
	}

	err = plan.DoBackup()
	if err != nil {
		t.Fatalf("Test died. Error while backuping files: %v\n", err)
	}

	points := plan.GetRestorePoints([]string{tfs.DataPath()})
	if len(points) == 0 {
		t.Fatalf("Test died. No restore points founded\n")
	}
	if len(points) != 1 {
		t.Errorf("Test failed. Qty of archives (restore points) in storage not as expected: got %v, expected %v\n",
			len(points), 1)
	}

	err = plan.InitRestore([]string{filepath.Join(tfs.DataPath(), "dir1"), filepath.Join(tfs.DataPath(), "dir3")},
		&points[0], tfs.RestorePath())
	if err != nil {
		t.Fatalf("Test died. Error while initializing restore: %v\n", err)
	}

	err = plan.DoRestore()
	if err != nil {
		t.Fatalf("Test died. Error while restoring files: %v\n", err)
	}

	dataPathAfterRestore := filepath.Join(tfs.RestorePath(), core.GetPathInArchive(tfs.DataPath()))
	cmpRes, err1 := testutils.CompareDirs(tfs.DataPath(), dataPathAfterRestore)
	if err1 != nil {
		t.Fatalf("Test died. Error while comparing directories: %v\n", err1)
	}

	if !cmpRes.Equals {
		t.Errorf("Test failed. Restored dir content differs from source dir content:\n%s", cmpRes.String())
	}

}

// sync meta files
func TestSyncMeta(t *testing.T) {
	checkSyncMeta(t, false)
}

// sync meta files, encrypted
func TestSyncMetaEncrypted(t *testing.T) {
	checkSyncMeta(t, true)
}

// sync meta files, plain/encrypted
func checkSyncMeta(t *testing.T, encrypted bool) {
	tfs, plan := InitTfsAndPlan(t)
	defer tfs.Destroy()

	if encrypted {
		plan.Encrypt = true
		plan.Encrypt_passphrase = "encryptpassphrasefortest1"
	}

	err := tfs.ApplyCmds(testutils.CmdsToApply{
		"create": {
			"dir1/file1.txt",
			"dir1/file2.txt",
			"dir3",
		},
	})
	if err != nil {
		t.Fatalf("Test died. Error while initializing filesystem: %v\n", err)
	}

	err = plan.DoBackup()
	if err != nil {
		t.Fatalf("Test died. Error while backuping files: %v\n", err)
	}

	err = tfs.ApplyCmds(testutils.CmdsToApply{
		"modify": {
			"dir1/file1.txt",
		},
	})
	if err != nil {
		t.Fatalf("Test died. Error while modifying filesystem: %v\n", err)
	}

	err = plan.DoBackup()
	if err != nil {
		t.Fatalf("Test died. Error while backuping files: %v\n", err)
	}

	planPathSnapshot, err1 := testutils.GetDirNodes(core.GetPlanDir(plan.Name))
	if err1 != nil {
		t.Fatalf("Test died. Error while taking plan path snapshot: %v\n", err1)
	}
	leaveOnlyArchiveMetaFiles(&planPathSnapshot)

	pointsQty := 2
	if len(planPathSnapshot) != pointsQty {
		t.Errorf("Test failed. Qty of archives (restore points) not as expected: got %v, expected %v\n",
			len(planPathSnapshot), pointsQty)
	}

	for _, node := range planPathSnapshot {
		err = os.Remove(node.GetNodePath())
		if err != nil {
			t.Fatalf("Test died. Error while deleting local meta files: %v\n", err)
		}
	}

	err = plan.SyncMeta(false)
	if err != nil {
		t.Fatalf("Test died. Error while synchronizing metafiles from remote to local storage: %v\n", err)
	}

	planPathSnapshot2, err2 := testutils.GetDirNodes(core.GetPlanDir(plan.Name))
	if err2 != nil {
		t.Fatalf("Test died. Error while taking plan path snapshot: %v\n", err2)
	}
	leaveOnlyArchiveMetaFiles(&planPathSnapshot2)

	if len(planPathSnapshot2) != pointsQty {
		t.Errorf("Test failed. Qty of archives (restore points) after sync not as expected: got %v, expected %v\n",
			len(planPathSnapshot2), pointsQty)
	}

	cmpRes, err3 := testutils.CompareDirNodes(planPathSnapshot, planPathSnapshot2)
	if err3 != nil {
		t.Fatalf("Test died. Error while comparing directories: %v\n", err3)
	}

	if !cmpRes.Equals {
		t.Errorf("Test failed. Restored metafiles content differs from source metafiles content:\n%s", cmpRes.String())
	}

	metaArchiveFileName := ""
	for p, _ := range planPathSnapshot2 {
		metaArchiveFileName = p
		break
	}
	_, err = core.ParseMetaFile(filepath.Join(core.GetPlanDir(plan.Name), metaArchiveFileName))
	if err != nil {
		t.Errorf("Test failed. Restored metafile can not be parsed\n")
	}
}

func InitTfsAndPlan(t *testing.T) (tfs testutils.TestFileSystem, plan core.BackupPlan) {
	tfs = testutils.CreateTestFileSystem()
	t.Logf("Test file system created with base path: %v\n", tfs.BasePath())

	base.InitApp(base.AppConfig{
		AppDir:         tfs.AppPath(),
		LogToStdout:    false,
		LogErrToStderr: false,
	})
	plan = testutils.CreateTestPlan(tfs, chunkSize)
	tfs.SetFileSize(fileSize)

	return
}

func leaveOnlyArchiveMetaFiles(nodesMap *map[string]core.NodeMetaInfo) {
	for p, _ := range *nodesMap {
		if !core.GetMetaFileNameRE().MatchString(p) {
			delete(*nodesMap, p)
		}
	}
}
