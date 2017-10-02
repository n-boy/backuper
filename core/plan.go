package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/n-boy/backuper/base"
	"github.com/n-boy/backuper/crypter"
	"github.com/n-boy/backuper/storage"
)

type BackupPlan struct {
	Name               string
	BaseDir            string
	TmpDir             string
	ChunkSize          int64
	Encrypt            bool
	Encrypt_passphrase string
	NodesToArchive     []string
	Storage            storage.GenericStorage

	cacheMetaFiles map[string]ArchiveMetafile
}

type yamlBackupPlanStruct struct {
	FilesList         []string `yaml:"files_list"`
	Storage           map[string]string
	ChunkSizeMB       int64  `yaml:"chunk_size_mb"`
	Encrypt           bool   `yaml:"encrypt"`
	EncryptPassphrase string `yaml:"encrypt_passphrase"`
}

var planFilename string = "plan.yaml"

// суммарный размер пачки файлов пакуемых в отдельный архив, МБ
var DefaultChunkSizeMB int64 = 1024

// допустипое превышение размера пачки файлов для архива, %
var chunkSizeExcessPct int64 = 10

func GetBackupPlan(planName string) (BackupPlan, error) {
	var plan BackupPlan
	if planName == "" {
		return plan, fmt.Errorf("Plan name can not be empty")
	}

	planDir := GetPlanDir(planName)
	var (
		err         error
		yamlContent []byte
	)
	yamlContent, err = ioutil.ReadFile(filepath.Join(planDir, planFilename))
	if err != nil {
		base.LogErr.Fatalln(err)
	}

	yamlBP := yamlBackupPlanStruct{}
	err = yaml.Unmarshal(yamlContent, &yamlBP)
	if err != nil {
		base.LogErr.Fatalln(err)
	}

	plan.NodesToArchive = yamlBP.FilesList
	plan.Storage, err = storage.NewStorage(yamlBP.Storage)
	if err != nil {
		base.LogErr.Fatalln(err)
	}
	if yamlBP.ChunkSizeMB == 0 {
		plan.ChunkSize = DefaultChunkSizeMB * 1024 * 1024
	} else {
		plan.ChunkSize = yamlBP.ChunkSizeMB * 1024 * 1024
	}
	plan.Encrypt = yamlBP.Encrypt
	plan.Encrypt_passphrase = yamlBP.EncryptPassphrase

	plan.Name = planName
	plan.BaseDir = planDir
	plan.TmpDir = filepath.Join(plan.BaseDir, "tmp")

	return plan, nil
}

func GetPlanDir(planName string) string {
	return filepath.Join(base.GetAppDir(), "plans", planName)
}

func IsBackupPlanExists(planName string) bool {
	_, err := os.Stat(filepath.Join(GetPlanDir(planName), planFilename))
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func (plan *BackupPlan) CheckTmpDir() error {
	if plan.TmpDir == "" {
		return fmt.Errorf("Temporary dir is not defined")
	}
	_, err := os.Stat(plan.TmpDir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(plan.TmpDir, 0700)
	}

	return err
}

func (plan *BackupPlan) SavePlan(overwrite bool) error {
	if plan.Name == "" {
		return fmt.Errorf("Plan name can not be empty")
	} else if !overwrite {
		if IsBackupPlanExists(plan.Name) {
			return fmt.Errorf("Plan with specified name already exists")
		}

	}

	yamlBP := yamlBackupPlanStruct{
		FilesList:         plan.NodesToArchive,
		ChunkSizeMB:       plan.ChunkSize / 1024 / 1024,
		Encrypt:           plan.Encrypt,
		EncryptPassphrase: plan.Encrypt_passphrase,
		Storage:           plan.Storage.GetStorageConfig(),
	}
	yamlBP.Storage["type"] = plan.Storage.GetType()

	planDir := GetPlanDir(plan.Name)
	if err := os.MkdirAll(planDir, 0700); err != nil {
		return err
	}

	yamlData, err := yaml.Marshal(&yamlBP)
	if err != nil {
		return err
	}

	planFilePath := filepath.Join(planDir, planFilename)
	planFilePathTmp := planFilePath + "~"

	if !overwrite {
		_, err := os.Stat(planFilePath)
		if err == nil {
			return fmt.Errorf("Plan is already exists")
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	err = ioutil.WriteFile(planFilePathTmp, yamlData, 0600)
	if err == nil {
		err = os.Rename(planFilePathTmp, planFilePath)
	}

	return err
}

func (plan BackupPlan) GetGuardedNodes() []NodeMetaInfo {
	var nodes NodeList
	for _, path := range plan.NodesToArchive {
		err := filepath.Walk(path, nodes.AddNodeToList)
		if err != nil {
			base.LogErr.Fatalln(err)
		}
	}
	return nodes.GetList()
}

func (plan BackupPlan) GetArchivedNodesMap() map[string]NodeMetaInfo {
	nodesMap := make(map[string]NodeMetaInfo)

	metafiles := plan.GetMetaFiles()
	for _, filename := range metafiles {
		archMeta := plan.GetMetaFile(filename)
		for _, node := range archMeta.GetNodes() {
			nodesMap[node.path] = node
		}
	}
	return nodesMap
}

func (plan BackupPlan) GetArchivedNodesAllRevMap() map[string][]NodeMetaInfo {
	nodesMap := make(map[string][]NodeMetaInfo)

	metafiles := plan.GetMetaFiles()
	for _, filename := range metafiles {
		archMeta := plan.GetMetaFile(filename)
		for _, node := range archMeta.GetNodes() {
			if _, exists := nodesMap[node.path]; !exists {
				nodesMap[node.path] = make([]NodeMetaInfo, 0)
			}
			nodesMap[node.path] = append(nodesMap[node.path], node)
		}
	}
	return nodesMap
}

func (plan BackupPlan) GetMetaFiles() MetafileList {
	var metafiles []string
	var err error
	err = os.Chdir(plan.BaseDir)
	if err != nil {
		base.LogErr.Fatalln(err)
	}

	metafiles, err = filepath.Glob(GetMetaFileGlobMask())
	if err != nil {
		base.LogErr.Fatalln(err)
	}

	var metafilesClean MetafileList
	for _, mf := range metafiles {
		if GetMetaFileNameRE().MatchString(mf) {
			metafilesClean = append(metafilesClean, mf)
		}
	}
	sort.Sort(metafilesClean)
	return metafilesClean
}

func (plan BackupPlan) GetMetaFile(filename string) ArchiveMetafile {
	if plan.cacheMetaFiles == nil {
		plan.cacheMetaFiles = make(map[string]ArchiveMetafile)
	}
	_, ok := plan.cacheMetaFiles[filename]
	if !ok {
		plan.cacheMetaFiles[filename] = GetMetaFile(filepath.Join(plan.BaseDir, filename))
	}
	return plan.cacheMetaFiles[filename]
}

func (plan BackupPlan) GetRemoteMetaFiles() ([]base.GenericStorageFileInfo, error) {
	metaFiles := []base.GenericStorageFileInfo{}

	remoteFiles, err := plan.Storage.GetFilesList()
	if err != nil {
		return metaFiles, err
	}

	for _, rf := range remoteFiles {
		if GetMetaFileNameRE().MatchString(rf.GetFilename()) {
			metaFiles = append(metaFiles, rf)
		}
	}
	return metaFiles, nil
}

func (plan BackupPlan) GetNextArchiveName() string {
	var err error
	metafiles := plan.GetMetaFiles()

	lastInd := 0
	if len(metafiles) > 0 {
		lastName := metafiles[len(metafiles)-1]
		lastInd, err = strconv.Atoi(strings.Split(lastName, "_")[1])
		if err != nil {
			base.LogErr.Fatalln(err)
		}
	}

	return fmt.Sprint("archive_", lastInd+1, "_", time.Now().Format("20060102150405"))
}

func (plan BackupPlan) GetProcessNodes(guardNodes []NodeMetaInfo, archNodesMap map[string]NodeMetaInfo) []NodeMetaInfo {
	procNodes := make([]NodeMetaInfo, 0)
	for _, node := range guardNodes {
		anode, anode_exists := archNodesMap[node.path]
		if !anode_exists ||
			(!node.is_dir &&
				(anode.size != node.size ||
					!anode.modtime.Truncate(time.Second).Equal(node.modtime.Truncate(time.Second)))) {
			procNodes = append(procNodes, node)
		}
	}
	return procNodes
}

func (plan BackupPlan) GetNodeChunks(nodes []NodeMetaInfo) [][]NodeMetaInfo {
	var chunks [][]NodeMetaInfo
	var chunkSize int64 = 0
	var indStart int = 0

	for ind, node := range nodes {
		if ind == len(nodes)-1 ||
			chunkSize+node.size >= plan.ChunkSize ||
			((chunkSize > 0 || node.size > 0) && chunkSize+nodes[ind+1].size > plan.ChunkSize*(100+chunkSizeExcessPct)/100) {

			chunks = append(chunks, nodes[indStart:ind+1])
			indStart = ind + 1
			chunkSize = 0

		} else {
			chunkSize += node.size
		}
	}
	return chunks
}

func (plan BackupPlan) DoBackup() error {
	base.Log.Printf("Start doing backup for plan: %v\n", plan.Name)

	if err := plan.CheckOpLockAllowed("backup"); err != nil {
		return err
	}

	if err := plan.CheckTmpDir(); err != nil {
		return err
	}

	// доливаем недокачанный архив
	if err := os.Chdir(plan.TmpDir); err != nil {
		return err
	}
	metafiles, err := filepath.Glob(GetMetaFileGlobMask())
	if err != nil {
		return err
	}
	for _, mf := range metafiles {
		archName := GetArchName(mf)
		_, err := os.Stat(fmt.Sprint(archName, ".zip"))
		if err != nil {
			base.LogErr.Println(err)
			base.Log.Printf("Remove metafile %v from tmp dir\n", mf)
			if err := os.Remove(filepath.Join(plan.TmpDir, mf)); err != nil {
				base.LogErr.Println(err)
			}
		} else {
			plan.uploadArchiveToStorage(archName)
		}
	}

	// получаем список файлов под наблюдением
	guardNodes := plan.GetGuardedNodes()

	// строим список архивированных файлов
	archNodesMap := plan.GetArchivedNodesMap()

	//  вычисляем список файлов к архивации
	procNodes := plan.GetProcessNodes(guardNodes, archNodesMap)

	// обрабатываем файлы по частям
	for _, chunk := range plan.GetNodeChunks(procNodes) {
		archName := plan.GetNextArchiveName()
		archFilepath := filepath.Join(plan.TmpDir, fmt.Sprint(archName, ".zip"))
		var encrypter *crypter.Encrypter
		if plan.Encrypt {
			encrypter = crypter.GetEncrypter(plan.Encrypt_passphrase)
		}
		doneNodes := ArchiveNodes(chunk, archFilepath, encrypter)
		base.Log.Printf("Archive %v created", archName)
		archMetaFilepath := filepath.Join(plan.TmpDir, GetMetaFileName(archName))
		archMeta := NewMetaFile(doneNodes, plan.Encrypt)
		err := archMeta.SaveMetaFile(archMetaFilepath)
		if err != nil {
			os.Remove(archFilepath)
			os.Remove(archMetaFilepath)
			base.LogErr.Fatalln(err)
		}
		base.Log.Printf("Metafile for archive %v created", archName)

		// заливаем архив в хранилище
		plan.uploadArchiveToStorage(archName)
	}

	base.Log.Printf("Finish doing backup for plan: %v", plan.Name)

	return nil
}

func (plan BackupPlan) uploadArchiveToStorage(archName string) {
	// заливаем архив в хранилище
	archMetaFilepath := filepath.Join(plan.TmpDir, GetMetaFileName(archName))
	archMeta := GetMetaFile(archMetaFilepath)

	archFilepath := filepath.Join(plan.TmpDir, fmt.Sprint(archName, ".zip"))
	archiveStorageInfo, err := plan.Storage.UploadFile(archFilepath, "")
	if err != nil {
		base.LogErr.Fatalln(err)
	}
	base.Log.Printf("Archive %v uploaded to storage", archName)
	archMeta.SetStorageInfo(archiveStorageInfo)
	err = archMeta.SaveMetaFile(archMetaFilepath)
	if err != nil {
		os.Remove(archMetaFilepath)
		os.Remove(archFilepath)
		plan.Storage.DeleteFile(archiveStorageInfo)
		base.LogErr.Fatalln(err)
	}

	// заливаем метафайл в хранилище
	// если в плане включено шифрование - шифруем и метафайл
	metaFilePathToUpload := archMetaFilepath
	var encArchMetaFilepath string
	if plan.Encrypt {
		encArchMetaFilepath = filepath.Join(filepath.Dir(archMetaFilepath), GetMetaFileNameEnc(archName))
		err = crypter.EncryptFile(plan.Encrypt_passphrase, archMetaFilepath, encArchMetaFilepath)
		if err != nil {
			base.LogErr.Fatalf("Error while encrypting metafile: %v\n", err)
		}
		metaFilePathToUpload = encArchMetaFilepath
	}

	_, err = plan.Storage.UploadFile(metaFilePathToUpload, "")
	if err != nil {
		plan.Storage.DeleteFile(archiveStorageInfo)
		base.LogErr.Fatalf("Error while uploading metafile to storage: %v\n", err)
	}
	base.Log.Printf("Metafile for archive %v uploaded to storage", archName)

	err = os.Remove(archFilepath)
	if err != nil {
		base.LogErr.Println(err)
	}
	if plan.Encrypt {
		err = os.Remove(encArchMetaFilepath)
		if err != nil {
			base.LogErr.Println(err)
		}
	}
	err = os.Rename(archMetaFilepath, filepath.Join(plan.BaseDir, GetMetaFileName(archName)))
	if err != nil {
		base.LogErr.Println(err)
	}
	base.Log.Printf("Metafile for archive %v moved to the base directory", archName)
}

func (plan BackupPlan) SyncMeta(cleanLocalMeta bool) error {
	base.Log.Printf("Trying to sync metafiles from storage for plan: %v\n", plan.Name)
	syncLocked := plan.CheckOpLocked("sync")

	if !syncLocked && len(plan.GetMetaFiles()) != 0 {
		if cleanLocalMeta {
			err := plan.CleanLocalMeta()
			if err != nil {
				return err
			}
		} else {
			return base.ErrLocalMetaExists
		}
	}

	if err := plan.CreateOpLock("sync"); err != nil {
		return err
	}

	base.Log.Printf("Start doing sync metafiles from storage for plan: %v\n", plan.Name)

	remoteMetaFiles, err := plan.GetRemoteMetaFiles()
	if err != nil {
		return err
	}
	localMetaFilesMap := make(map[string]bool)
	for _, lmf := range plan.GetMetaFiles() {
		localMetaFilesMap[lmf] = true
	}
	var procMetaFiles []base.GenericStorageFileInfo
	for _, rmf := range remoteMetaFiles {
		if cf, _ := CleanMetaFileNameEnc(rmf.GetFilename()); !localMetaFilesMap[cf] {
			procMetaFiles = append(procMetaFiles, rmf)
		}
	}

	errInProgress := false
	for _, pmf := range procMetaFiles {
		base.Log.Printf("Start downloading metafile %v\n", pmf.GetFilename())
		cf, encrypted := CleanMetaFileNameEnc(pmf.GetFilename())

		downloadedFilePath := filepath.Join(plan.BaseDir, cf)
		err := plan.DownloadAndDecryptFile(pmf.GetFileStorageId(), downloadedFilePath, encrypted)
		if err != nil {
			if err == base.ErrStorageRequestInProgress {
				base.Log.Println(err)
				errInProgress = true
			} else {
				return err
			}
		} else {
			if encrypted {
				base.Log.Printf("Finish downloading metafile %v and decrypting to %v\n",
					pmf.GetFilename(), cf)
			} else {
				base.Log.Printf("Finish downloading metafile %v\n", pmf.GetFilename())
			}
		}
	}

	if !errInProgress {
		err := plan.RemoveOpLock("sync")
		if err == nil {
			base.Log.Printf("Finish doing sync metafiles from storage for plan: %v\n", plan.Name)
		}
		return err
	} else {
		base.Log.Printf("Start doing sync metafiles from storage for plan: %v\n", plan.Name)
		return base.ErrStorageRequestInProgress
	}
}

func (plan BackupPlan) CleanLocalMeta() error {
	err := os.Chdir(plan.BaseDir)
	if err != nil {
		return err
	}

	for _, filename := range plan.GetMetaFiles() {
		err := os.Remove(filename)
		if err != nil {
			return err
		}
	}
	return nil
}
