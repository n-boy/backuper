package core

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/n-boy/backuper/base"
	"github.com/n-boy/backuper/crypter"
)

const (
	OriginTargetPath string = "ORIGIN"
	restoreOp        string = "restore"
)

type RestorePlan struct {
	TargetPath         string
	ArchNodesToRestore map[string][]NodeMetaInfo
}

type yamlRestorePlan struct {
	TargetPath      string              `yaml:"target_path"`
	NodesFormatCSV  string              `yaml:"files_format"`
	ArchiveNodesCSV map[string][]string `yaml:"archive_files"`
}

func (plan BackupPlan) CheckRestoreAllowed() error {
	return plan.CheckOpLockAllowed("restore")
}

func (plan BackupPlan) GetRestorePoints(pathList []string) []ArchiveMetafile {
	metaFiles := make([]ArchiveMetafile, 0)
	for _, filename := range plan.GetMetaFiles() {
		mf := plan.GetMetaFile(filename)
		for _, node := range mf.GetNodes() {
			nodeInPath := false
			for _, path := range pathList {
				nodeInPath = node.isNodeInPath(path)
				if nodeInPath {
					break
				}
			}
			if nodeInPath {
				metaFiles = append(metaFiles, mf)
				break
			}
		}
	}
	return metaFiles
}

func (plan BackupPlan) InitRestore(pathList []string, restorePoint *ArchiveMetafile, targetPath string) error {
	base.Log.Printf("Start initialize data restore for plan: %v\n", plan.Name)
	if targetPath != OriginTargetPath {
		tps, err := os.Stat(targetPath)
		if err != nil || !tps.IsDir() || !filepath.IsAbs(targetPath) {
			return fmt.Errorf("Target path should be absolute path to existing directory")
		}
	}

	if restorePoint != nil && restorePoint.GetMetaFileId() == 0 {
		return fmt.Errorf("Restore point is not defined")
	}

	// pathes founded in archives
	pathFounded := make(map[string]bool)
	nodesRegistered := make(map[string]bool)
	archNodesToRestore := make(map[string][]NodeMetaInfo)

	// clear path list from subpathes of each other
	for i := range pathList {
		pathList[i] = filepath.Clean(pathList[i])
	}
	pathListUniq := []string{}
	for i, pathToCheck := range pathList {
		isSubPath := false
		for j, path := range pathList {
			if i == j {
				continue
			}
			if base.IsPathInBasePath(path, pathToCheck) {
				isSubPath = true
				break
			}
		}
		if !isSubPath {
			pathListUniq = append(pathListUniq, pathToCheck)
		}
	}

	metaFiles := plan.GetMetaFiles()
	restorePointOk := false
	for i := len(metaFiles) - 1; i >= 0; i-- {
		mf := plan.GetMetaFile(metaFiles[i])
		if !restorePointOk && restorePoint != nil && mf.GetMetaFileId() != restorePoint.GetMetaFileId() {
			continue
		}
		restorePointOk = true

		nodesToRestore := make([]NodeMetaInfo, 0)
		for _, node := range mf.GetNodes() {
			if nodesRegistered[node.GetNodePath()] {
				continue
			}
			for _, path := range pathListUniq {
				if node.isNodeInPath(path) {
					nodesRegistered[node.GetNodePath()] = true
					nodesToRestore = append(nodesToRestore, node)
					pathFounded[path] = true
					break
				}
			}
		}

		if len(nodesToRestore) > 0 {
			archNodesToRestore[mf.GetMetaFileNameId()] = nodesToRestore
		}
	}

	// check for nothing to restore at all
	if len(archNodesToRestore) == 0 {
		return fmt.Errorf("There is nothing to restore by selected pathes")
	}
	// check for nothing to restore by path
	missedPathList := []string{}
	for _, path := range pathListUniq {
		if !pathFounded[path] {
			missedPathList = append(missedPathList, path)
		}
	}
	if len(missedPathList) > 0 {
		return fmt.Errorf("There is nothing by some of selected pathes: %v", strings.Join(missedPathList, ", "))
	}
	// check locks
	if plan.CheckOpLocked(restoreOp) {
		return fmt.Errorf("Restore job is already initialized")
	}
	if err := plan.CheckOpLockAllowed("restore"); err != nil {
		return err
	}

	err := plan.SaveRestorePlan(RestorePlan{
		TargetPath:         targetPath,
		ArchNodesToRestore: archNodesToRestore,
	})
	if err == nil {
		base.Log.Printf("Finish initialize data restore for plan: %v\n", plan.Name)
	}
	return err
}

func (plan BackupPlan) GetRestorePlan() (RestorePlan, error) {
	var rplan RestorePlan

	yamlContent, err := ioutil.ReadFile(plan.getRestorePlanFilePath())
	if err != nil {
		return rplan, err
	}

	yamlRPlan := yamlRestorePlan{}
	err = yaml.Unmarshal(yamlContent, &yamlRPlan)
	if err != nil {
		return rplan, err
	}

	rplan.TargetPath = yamlRPlan.TargetPath
	rplan.ArchNodesToRestore = make(map[string][]NodeMetaInfo)

	nodes_format := strings.Split(yamlRPlan.NodesFormatCSV, ",")
	for archNameId, nodesCSV := range yamlRPlan.ArchiveNodesCSV {
		nodes := make([]NodeMetaInfo, 0)
		for _, nodeCSV := range nodesCSV {
			node, err := GetNodeFromString(nodeCSV, nodes_format)
			if err != nil {
				return rplan, err
			}
			nodes = append(nodes, node)
		}
		rplan.ArchNodesToRestore[archNameId] = nodes
	}

	return rplan, nil
}

func (plan BackupPlan) SaveRestorePlan(rplan RestorePlan) error {
	yamlRPlan := yamlRestorePlan{
		TargetPath:      rplan.TargetPath,
		NodesFormatCSV:  strings.Join(GetNodeCurrentFormat(), ","),
		ArchiveNodesCSV: make(map[string][]string),
	}
	for archNameId, nodes := range rplan.ArchNodesToRestore {
		nodesCSV := make([]string, 0)
		for _, node := range nodes {
			nodesCSV = append(nodesCSV, node.ToString())
		}
		yamlRPlan.ArchiveNodesCSV[archNameId] = nodesCSV
	}

	yamlData, err := yaml.Marshal(&yamlRPlan)
	if err != nil {
		return err
	}

	err = os.Remove(plan.getRestorePlanDoneFilePath())
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	restorePlanFilePathTmp := fmt.Sprint(plan.getRestorePlanFilePath(), "~")
	err = ioutil.WriteFile(restorePlanFilePathTmp, yamlData, 0666)
	if err == nil {
		err = os.Rename(restorePlanFilePathTmp, plan.getRestorePlanFilePath())
	}

	return err
}

func (plan BackupPlan) IsExistsRestorePlan() bool {
	_, err := plan.GetRestorePlan()

	return err == nil
}

func (plan BackupPlan) getRestorePlanFilePath() string {
	return filepath.Join(plan.BaseDir, "restore_plan.yaml")
}

func (plan BackupPlan) getRestorePlanDoneFilePath() string {
	return filepath.Join(plan.BaseDir, "restore_done.log")
}

func (plan BackupPlan) getRestoredNodes() (map[string]bool, error) {
	restoredNodes := make(map[string]bool)

	data, err := ioutil.ReadFile(plan.getRestorePlanDoneFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return restoredNodes, nil
		} else {
			return restoredNodes, err
		}
	}

	for _, f := range strings.Split(string(data), "\r\n") {
		if f != "" {
			restoredNodes[f] = true
		}
	}
	return restoredNodes, nil
}

func (plan BackupPlan) DoRestore() error {
	base.Log.Printf("Trying to start restore for plan: %v\n", plan.Name)
	rplan, err := plan.GetRestorePlan()
	if err != nil {
		return err
	}

	if err := plan.CreateOpLock("restore"); err != nil {
		return err
	}

	base.Log.Printf("Start doing restore for plan: %v\n", plan.Name)
	restoredNodes, err := plan.getRestoredNodes()
	if err != nil {
		return err
	}

	errInProgress := false
	if err := plan.CheckTmpDir(); err != nil {
		return err
	}

ARCH_LOOP:
	for archNameId, nodes := range rplan.ArchNodesToRestore {
		nodesToRestore := make([]NodeMetaInfo, 0)
		for _, node := range nodes {
			if !restoredNodes[node.GetNodePath()] {
				nodesToRestore = append(nodesToRestore, node)
			}
		}
		if len(nodesToRestore) > 0 {
			// TODO если восстанавливаются только директории - не качать архив

			archName := "archive_" + archNameId
			mf := GetMetaFile(filepath.Join(plan.BaseDir, GetMetaFileName(archName)))
			archLocalFilePath := filepath.Join(plan.TmpDir, "restore_archive_"+archName+".zip")
			_, err := os.Stat(archLocalFilePath)
			if err != nil {
				if os.IsNotExist(err) {
					base.Log.Printf("Start downloading archive %v\n", archName+".zip")
					err = plan.DownloadAndDecryptFile(mf.GetStorageInfo(), archLocalFilePath, mf.encrypted)
					if err == nil {
						base.Log.Printf("Finish downloading archive %v\n", archName+".zip")
					}
				}
				if err != nil {
					if err == base.ErrStorageRequestInProgress {
						base.Log.Println(err)
						errInProgress = true
						continue ARCH_LOOP
					} else {
						return err
					}
				}
			}

			nodesUnarch, err := UnarchiveNodes(archLocalFilePath, nodesToRestore, rplan.TargetPath)
			if err != nil {
				return err
			}

			fh, err := os.OpenFile(plan.getRestorePlanDoneFilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				return err
			}
			defer fh.Close()
			for _, node := range nodesUnarch {
				if _, err := fh.WriteString(node.GetNodePath() + "\r\n"); err != nil {
					return err
				}
			}
			if err = fh.Close(); err != nil {
				return err
			}
			if err = os.Remove(archLocalFilePath); err != nil {
				base.LogErr.Println(err)
			}
		}
	}
	if errInProgress {
		return base.ErrStorageRequestInProgress
	}

	if err = os.Remove(plan.getRestorePlanDoneFilePath()); err != nil {
		base.LogErr.Println(err)
	} else if err = os.Remove(plan.getRestorePlanFilePath()); err != nil {
		base.LogErr.Println(err)
	}
	err = plan.RemoveOpLock("restore")

	if err == nil {
		base.Log.Printf("Finish doing restore for plan: %v\n", plan.Name)
	}
	return err
	// достаем список уже восстановленных файлов
	// для каждого архива из файла
	// 	- определяем файлы которые еще не восстановили
	// 	- скачиваем архив (если еще не скачали)
	// 	- восстанавливаем файлы
	// 		входные параметры: путь к архиву, список файлов
	// 		выходные: список восстановленных файлов, ошибка
	// 		если восстанавливаемый файл существует, и совпадает размер и дата - пропускаем (считаем восстановлненным)
	// 	- удаляем архив
}

func (plan BackupPlan) DownloadAndDecryptFile(fileStorageInfo map[string]string, localFilePath string, isEncrypted bool) error {
	localFilePathShadow := localFilePath + "~"
	fileWriter, err := os.OpenFile(localFilePathShadow, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0660)
	if err != nil {
		return err
	}
	defer fileWriter.Close()

	w := io.Writer(fileWriter)
	var decrypter *crypter.Decrypter
	if isEncrypted {
		decrypter = crypter.GetDecrypter(plan.Encrypt_passphrase)
		decrypter.InitWriter(fileWriter)
		w = io.Writer(decrypter)
	}

	err = plan.Storage.DownloadFileToPipe(fileStorageInfo, w)
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
