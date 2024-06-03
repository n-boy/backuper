package base

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nightlyone/lockfile"
)

type AppConfig struct {
	AppDir         string
	LogToStdout    bool
	LogErrToStderr bool
}

var DefaultAppConfig AppConfig = AppConfig{
	LogToStdout:    true,
	LogErrToStderr: true,
}

var (
	Log    *log.Logger
	LogErr *log.Logger

	appLock lockfile.Lockfile

	appConfig AppConfig
)
var ErrStorageRequestInProgress = errors.New("Request to storage is in progress")
var ErrLocalMetaExists = errors.New("Can't start synchronizing metadata. Local metafiles exists in plan directory")

var StorageRequestInProgressRetrySeconds int64 = 30 * 60

type GenericStorageFileInfo interface {
	GetFilename() string
	GetFileStorageId() map[string]string
}

func InitApp(config AppConfig) {
	appConfig = config
	if appConfig.AppDir == "" {
		appConfig.AppDir = createAppDir()
	}

	if _, err := os.Stat(appConfig.AppDir); os.IsNotExist(err) {
		panic(err)
	}

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

func SetAppConfig(config AppConfig) {
	appConfig = config
}

func InitLogToDestination(dwref *io.Writer) {
	var dw io.Writer
	if dwref == nil {
		filePath := filepath.Join(GetAppDir(), "history.log")

		fh, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}

		dw = (io.Writer)(fh)
	} else {
		dw = *dwref
	}

	w := io.MultiWriter(dw)

	if appConfig.LogToStdout {
		w = io.MultiWriter(dw, os.Stdout)
	}
	Log = log.New(w, "[INFO] ", log.LstdFlags)

	w = io.MultiWriter(dw)
	if appConfig.LogToStdout {
		w = io.MultiWriter(dw, os.Stderr)
	}
	LogErr = log.New(w, "[ERROR] ", log.LstdFlags)
}

func initLog() {
	InitLogToDestination(nil)
}

func GetAppDir() string {
	if _, err := os.Stat(appConfig.AppDir); appConfig.AppDir == "" || os.IsNotExist(err) {
		panic("AppDir is not initialized")
	}
	return appConfig.AppDir
}

func createAppDir() string {
	var basePath = ""

	if basePath == "" {
		switch runtime.GOOS {
		case "windows":
			basePath = os.Getenv("LOCALAPPDATA")
		case "darwin":
			basePath = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
		case "linux":
			// basePath = filepath.Join(os.Getenv("HOME"))
			basePath = "/nfs/Public/Backup"
		}
	}
	appDir := filepath.Join(basePath, "Backuper")

	err := os.Mkdir(appDir, 0770)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	return appDir
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

func IsPathInBasePath(basePath, path string) bool {
	r, err := filepath.Rel(basePath, path)
	if err == nil && strings.Split(r, string(filepath.Separator))[0] != ".." {
		return true
	}
	return false
}

func IsSubPathToBasePath(basePath, path string) bool {
	if IsPathInBasePath(basePath, path) {
		r, _ := filepath.Rel(basePath, path)
		if r != "" && r != "." {
			return true
		}
	}
	return false
}

func GetFirstLevelPath(basePath, path string) string {
	if path == "" || (basePath != "" && !IsSubPathToBasePath(basePath, path)) {
		return ""
	} else if basePath == "" {
		return GetPathFirstPart(path)
	}

	r, err := filepath.Rel(basePath, path)
	if err != nil {
		return ""
	}
	return filepath.Join(basePath, GetPathFirstPart(r))
}

func GetPathFirstPart(path string) string {
	parts := strings.Split(path, string(filepath.Separator))

	pfp := ""
	if len(parts) == 0 {
		return pfp
	} else if parts[0] == "" && len(parts) > 1 {
		pfp = strings.Join(parts[0:2], string(filepath.Separator))
	} else {
		pfp = parts[0]
	}
	return filepath.Clean(pfp + string(filepath.Separator))
}
