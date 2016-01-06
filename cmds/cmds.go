package cmds

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/n-boy/backuper/base"
	"github.com/n-boy/backuper/core"
	"github.com/n-boy/backuper/storage"
	"github.com/n-boy/backuper/webui"
)

func Create() {
	plan := core.BackupPlan{}
	createOrEdit(plan, true)
}

func Edit(plan core.BackupPlan) {
	createOrEdit(plan, false)
}

func createOrEdit(plan core.BackupPlan, is_new bool) {
	if is_new {
		plan.Name = getInput("Plan name", "",
			func(planName string) (err error) {
				if planName == "" {
					err = fmt.Errorf("Plan name can not be empty")
				} else if core.IsBackupPlanExists(planName) {
					err = fmt.Errorf("Plan with specified name already exists")
				}
				return
			})
	}

	defaultChunkSizeMB := core.DefaultChunkSizeMB
	if !is_new {
		defaultChunkSizeMB = plan.ChunkSize / 1024 / 1024
	}
	chunkSizeMB, _ := strconv.ParseInt(getInput("Limit size of one archive (MB)", strconv.FormatInt(defaultChunkSizeMB, 10),
		func(text string) error {
			return checkInt(text, 0, 100*1024)
		}), 10, 64)
	plan.ChunkSize = chunkSizeMB * 1024 * 1024

	defaultEncrypt := ""
	if !is_new {
		if plan.Encrypt {
			defaultEncrypt = "Yes"
		} else {
			defaultEncrypt = "No"
		}
	}
	plan.Encrypt, _ = parseCmdsBool(getInput("Encrypt data [Y/N]", defaultEncrypt,
		func(text string) error {
			return checkCmdsBool(text)
		}))
	plan.Encrypt_passphrase = getInput("Encryption/Decryption passphrase (24-32 symbols)", plan.Encrypt_passphrase,
		func(passphrase string) error {
			if plan.Encrypt && !(len(passphrase) >= 24 && len(passphrase) <= 40) {
				return fmt.Errorf("Passphrase length should be between 24 and 32 symbols")
			}
			return nil
		})

	editPathes := false
	if !is_new {
		fmt.Println("Currenct list of pathes to backup:")
		for _, path := range plan.NodesToArchive {
			fmt.Printf("    %v\n", path)
		}
		editPathes, _ = parseCmdsBool(getInput("Do you want set up new list of pathes to backup? [Y/N]", "",
			func(text string) error {
				return checkCmdsBool(text)
			}))
	}

	if is_new || editPathes {
		plan.NodesToArchive = getInputList("Provide pathes you want to backup", "one more path", false,
			func(path string) error {
				if path != "" {
					_, err := os.Stat(path)
					if err != nil || !filepath.IsAbs(path) {
						return fmt.Errorf("Path should be absolute path to existing directory or file")
					}
				}
				return nil
			})
	}

	defaultStorageType := ""
	if !is_new && plan.Storage != nil {
		defaultStorageType = plan.Storage.GetType()
	}
	storageType := getInput("Storage type ["+strings.Join(storage.GetStorageTypes(), "/")+"]", defaultStorageType,
		func(stype string) error {
			if !storage.GetStorageTypesMap()[stype] {
				return fmt.Errorf("Storage type is not supported")
			}
			return nil
		})
	storageOldConfig := make(map[string]string)
	if !is_new && plan.Storage != nil && storageType == plan.Storage.GetType() {
		storageOldConfig = plan.Storage.GetStorageConfig()
	}
	storageFields, err := storage.GetStorageConfigFields(storageType)
	if err != nil {
		base.LogErr.Fatalln(err)
	}
	storageConfig := make(map[string]string)
	storageConfig["type"] = storageType
	for _, cf := range storageFields {
		storageConfig[cf.Name] = getInput(cf.Title, storageOldConfig[cf.Name],
			func(value string) error {
				return nil
			})
	}
	if plan.Storage, err = storage.NewStorage(storageConfig); err != nil {
		base.LogErr.Fatalln(err)
	}

	if err = plan.SavePlan(!is_new); err != nil {
		base.LogErr.Fatalln(err)
	}

	if is_new {
		fmt.Println("\nPlan successfully created\n")
	} else {
		fmt.Println("\nPlan successfully edited\n")
	}

	// name               string
	// baseDir            string
	// tmpDir             string
	// chunkSize          int64
	// encrypt            bool
	// encrypt_passphrase string
	// nodesToArchive     []string
	// storage            storage.GenericStorage

}

// func delete(plan core.BackupPlan) {

// }

// 	выводим все поля плана
func View(plan core.BackupPlan) {
	fmt.Printf("Plan name: %v\n", plan.Name)
	fmt.Printf("Limit size of one archive (MB): %v\n", plan.ChunkSize/1024/1024)

	fmt.Print("Encrypt data: ")
	if plan.Encrypt {
		fmt.Println("Yes")
		fmt.Printf("Encryption/Decryption passphrase: %v\n", plan.Encrypt_passphrase)
	} else {
		fmt.Println("No")
	}

	fmt.Println("Pathes to backup:")
	for _, path := range plan.NodesToArchive {
		fmt.Printf("    %v\n", path)
	}

	fmt.Printf("\nStorage type: %v\n", plan.Storage.GetType())

	storageFields, err := storage.GetStorageConfigFields(plan.Storage.GetType())
	if err != nil {
		base.LogErr.Fatalln(err)
	}
	storageConfig := plan.Storage.GetStorageConfig()
	for _, cf := range storageFields {
		fmt.Printf("%v: %v\n", cf.Title, storageConfig[cf.Name])
	}

	fmt.Println("")
}

// 	выводим текущую выполняемую планом команду
func Status(plan core.BackupPlan) {
	if plan.CheckOpLocked("restore") {
		fmt.Println("Data restoring is in progress")
	} else if plan.CheckOpLocked("sync") {
		fmt.Println("Synchronizing of metadata with storage is in progress")
	} else {
		fmt.Println("No operations in progress")
	}
}

// 	запускаем процесс бекапа согласно настроек плана
func Backup(plan core.BackupPlan) {
	err := plan.DoBackup()
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
	}
}

func Restore(plan core.BackupPlan) {
	if !plan.CheckOpLocked("restore") {
		pathList := getInputList("Provide pathes you want to restore", "one more path", true,
			func(path string) error {
				if path != "" && !filepath.IsAbs(path) {
					return fmt.Errorf("Path should be absolute path to directory or file")
				}
				return nil
			})
		restorePoints := plan.GetRestorePoints(pathList)
		if len(restorePoints) == 0 {
			fmt.Printf("[ERROR] There are no restore points available for selected pathes\n")
			return
		}

		fmt.Println("Restore points available for selected pathes:")
		for i, rp := range restorePoints {
			fmt.Printf("%v) %v\n", i+1, rp.GetMetaFileCreateDate().Format("2006-01-02 15:04:05"))
		}

		restorePointInd, _ := strconv.ParseInt(getInput("Select one of restore point", "",
			func(text string) error {
				return checkInt(text, 1, int64(len(restorePoints)))
			}), 10, 64)

		targetPath := getInput("Provide target path to restore ("+core.OriginTargetPath+" for restore to each file/directory origin path)", "",
			func(path string) error {
				if path != core.OriginTargetPath {
					fi, err := os.Stat(path)
					if err != nil || !filepath.IsAbs(path) || !fi.IsDir() {
						return fmt.Errorf("Path should be absolute path to existing directory or %v", core.OriginTargetPath)
					}
				}
				return nil
			})

		if err := plan.InitRestore(pathList, &restorePoints[restorePointInd-1], targetPath); err != nil {
			fmt.Printf("[ERROR] %v\n", err)
			return
		}
	}

	for {
		err := plan.DoRestore()
		if err != nil {
			if err == base.ErrStorageRequestInProgress {
				fmt.Printf("Request to storage is in progress. Waiting for %v seconds...\n", base.StorageRequestInProgressRetrySeconds)
				time.Sleep(time.Duration(base.StorageRequestInProgressRetrySeconds) * time.Second)
			} else {
				fmt.Printf("[ERROR] %v\n", err)
				return
			}
		} else {
			break
		}
	}

}

func Sync(plan core.BackupPlan) {
	deleteLocalMetafiles := false
	for {
		err := plan.SyncMeta(deleteLocalMetafiles)
		if err != nil {
			if err == base.ErrStorageRequestInProgress {
				fmt.Printf("Request to storage is in progress. Waiting for %v seconds...\n", base.StorageRequestInProgressRetrySeconds)
				time.Sleep(time.Duration(base.StorageRequestInProgressRetrySeconds) * time.Second)
			} else if err == base.ErrLocalMetaExists {
				fmt.Printf("[ERROR] %v\n", err)
				deleteLocalMetafiles, _ = parseCmdsBool(getInput("Do you want to delete local metafiles and completely get them from storage [Y/N]", "",
					func(text string) error {
						return checkCmdsBool(text)
					}))
				if !deleteLocalMetafiles {
					return
				}
			} else {
				fmt.Printf("[ERROR] %v\n", err)
				return
			}
		} else {
			break
		}
	}
}

func WebUI(plan core.BackupPlan) {
	webui.Init(plan.Name)
}

func getInput(title string, defaultValue string, checkFunc func(string) error) string {
	input := ""
	for {
		if defaultValue == "" {
			fmt.Print(title + ": ")
		} else {
			fmt.Printf("%v (default: %v): ", title, defaultValue)
		}
		_, err := fmt.Scanln(&input)
		if err != nil && err.Error() != "unexpected newline" {
			panic(err)
		}

		if input == "" && defaultValue != "" {
			return defaultValue
		} else if err = checkFunc(input); err != nil {
			fmt.Printf("[ERROR] %v\n", err)
		} else {
			return input
		}
	}
}

func getInputList(title string, oneItemTitle string, notEmpty bool, checkFunc func(string) error) []string {
	list := make([]string, 0)

	fmt.Println(title + ": ")
	for {
		itemText := getInput("    "+oneItemTitle, "", checkFunc)
		if itemText != "" {
			list = append(list, itemText)
		} else {
			if notEmpty && len(list) == 0 {
				fmt.Printf("[ERROR] Values list should not be empty\n")
			} else {
				break
			}
		}
	}

	return list
}

func checkInt(text string, min, max int64) error {
	number, err := strconv.ParseInt(text, 10, 64)
	if err != nil || !(number >= min && number <= max) {
		return fmt.Errorf("The value should be integer number between %v and %v", min, max)
	}

	return nil
}

func parseCmdsBool(text string) (bool, error) {
	true_values := []string{"Y", "y", "Yes", "yes"}
	false_values := []string{"N", "n", "No", "no"}

	if regexp.MustCompile("^(" + strings.Join(true_values, "|") + ")$").MatchString(text) {
		return true, nil
	} else if regexp.MustCompile("^(" + strings.Join(false_values, "|") + ")$").MatchString(text) {
		return false, nil
	} else {
		return false, fmt.Errorf("Value shoud be one of: %v, %v", strings.Join(true_values, ", "), strings.Join(false_values, ", "))
	}
}

func checkCmdsBool(text string) error {
	_, err := parseCmdsBool(text)
	return err
}
