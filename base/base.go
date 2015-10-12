package base

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/nightlyone/lockfile"
)

var (
	Log    *log.Logger
	LogErr *log.Logger

	appLock lockfile.Lockfile
)
var ErrStorageRequestInProgress = errors.New("Request to storage is in progress")
var ErrLocalMetaExists = errors.New("Can't start synchronizing metadata. Local metafiles exists in plan directory")

var StorageRequestInProgressRetrySeconds int64 = 30 * 60

type GenericStorageFileInfo interface {
	GetFilename() string
	GetFileStorageId() map[string]string
}

func InitApp() {
	createAppDir()
	initLog()
	if runtime.GOOS != "windows" {
		getAppLock()
	}

	// lockch := make(chan lockfile.Lockfile, 1)
	// go func() {
	// 	s := <-lockch
	// 	if s != nil {
	// 		FinishApp()
	// 		os.Exit(1)
	// 	}
	// }()
}

func FinishApp() {
	if runtime.GOOS != "windows" {
		releaseAppLock()
	}
}

func initLog() {
	filePath := filepath.Join(GetAppDir(), "history.log")

	fh, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	Log = log.New(io.MultiWriter(fh, os.Stdout), "[INFO] ", log.LstdFlags)
	LogErr = log.New(io.MultiWriter(fh, os.Stderr), "[ERROR] ", log.LstdFlags)
}

func GetAppDir() string {
	var basePath string

	switch runtime.GOOS {
	case "windows":
		basePath = os.Getenv("LOCALAPPDATA")
	case "darwin":
		basePath = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	}
	return filepath.Join(basePath, "Backuper")
}

func createAppDir() {
	err := os.Mkdir(GetAppDir(), 0770)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
}

func getAppLock() {
	var err error
	appLock, err = lockfile.New(filepath.Join(GetAppDir(), "run.lock"))
	if err != nil {
		LogErr.Fatalf("Cannot init app lock: %v\n", err)
	}

	err = appLock.TryLock()
	if err != nil {
		LogErr.Fatalf("Cannot obtain app lock: %v\n", err)
	}
}

func releaseAppLock() {
	err := appLock.Unlock()
	if err != nil {
		LogErr.Printf("Cannot release app lock: %v\n", err)
	}
}
