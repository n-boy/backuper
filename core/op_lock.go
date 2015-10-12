package core

import (
	"github.com/n-boy/backuper/base"

	"fmt"
	"os"
	"path/filepath"
)

var lockOperations = [...]string{"backup", "sync", "restore"}

func (plan BackupPlan) CheckOpLocked(op string) bool {
	CheckLockOperation(op)
	opLocked := false
	if _, err := os.Stat(plan.GetOpLockFile(op)); err == nil || !os.IsNotExist(err) {
		opLocked = true
	}
	return opLocked
}

func (plan BackupPlan) CreateOpLock(op string) error {
	err := plan.CheckOpLockAllowed(op)
	if err != nil {
		return err
	}
	opLock, err := os.Create(plan.GetOpLockFile(op))
	if !os.IsExist(err) {
		if err == nil {
			err = opLock.Close()
		}
		if err != nil {
			return fmt.Errorf("Can't create %v lock file: %v", op, err)
		}
	}
	return nil
}

func (plan BackupPlan) RemoveOpLock(op string) error {
	CheckLockOperation(op)
	err := os.Remove(plan.GetOpLockFile(op))
	if err != nil {
		return fmt.Errorf("Can't remove %v lock file: %v", op, err)
	}
	return nil
}

func (plan BackupPlan) GetOpLockFile(op string) string {
	return filepath.Join(plan.BaseDir, fmt.Sprint(op, ".lock"))
}

func (plan BackupPlan) CheckOpLockAllowed(op string) error {
	CheckLockOperation(op)
	for _, ok_op := range lockOperations {
		if ok_op != op && plan.CheckOpLocked(ok_op) {
			return fmt.Errorf("Operation '%v' is locked by operation '%v'", op, ok_op)
		}
	}
	return nil
}

func CheckLockOperation(op string) {
	ok := false
	for _, ok_op := range lockOperations {
		if ok_op == op {
			ok = true
			break
		}
	}
	if !ok {
		base.LogErr.Fatalf("Unsupported lock operation declared: %v\n", op)
	}
}
