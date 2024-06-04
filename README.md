# Backuper

Incremental "backuper" of local files to some storage.

Implemented **types of storage** (the list could be easily expanded):
- local filesystem
- Amazon Glacier

The application does NOT synchronize a local folder with a remote one (and vice versa). It **designed to work like this**:
- finds files/folders which have not been backed up (or have been changed) in guarded directories
- splits obtained list of files on chunks (by default, chunk size - 1GB)
- for each chunk
	- composes archive with files form chunk
	- writes list of files into meta info file
	- uploads the archive to storage
	- saves a copy of meta info files locally and uploads them to storage

Some **features**:
- supports many backup plans on one computer
- supports data encryption before uploading to storage
- supports files restoration in accordance with chosen recovery point
- by uploading only small number of large archives, supports a high speed of initial uploading
- could be terminated any time (this leads to reprocessing only one chunk/archive)
- command-line interface for managing backup plans
- web interface for analizing backed up files/folders, and used capacity

### Command-line interface
```
> backuper.exe
usage: D:\...\backuper.exe --create-plan
       D:\...\backuper.exe --plan my_plan_name --<command>
possible commands:
    --edit
    --view
    --status
    --backup
    --restore
    --sync
    --web-ui
```

After creation of backup plan (it is interactive), we can view created plan details:
```
> backuper.exe --plan backup_test --view
Plan name: backup_test
Limit size of one archive (MB): 100
Encrypt data: Yes
Encryption/Decryption passphrase: 12345
Pathes to backup:
    C:\Dir1
    C:\Dir2

Storage type: glacier
AWS Region: eu-central-1
Vault Name: backup-test
Access Key ID: qwerty12345
Secret Access Key: 12345qwerty
```

Common task to be run by daily/weekly schedule:
```
> backuper.exe --plan backup_test --backup
```

Command to use web interface:
```
> backuper.exe --plan backup_test --web-ui
[INFO] 2017/10/02 20:31:13 Indexing local filesystem...
[INFO] 2017/10/02 20:31:13 Starting web service on http://localhost:8080
```

Use `--sync` command to **restore metafiles** from remote storage (usually they are stored locally).
To **restore data files** use interactive command `--restore`.


### How to start using
```
go get -u github.com/n-boy/backuper
cd .../backuper
go build

backuper.exe --create-plan
...
```

**pre-build actions** (if html/css modified)
```
go get -u github.com/jteeuwen/go-bindata/...
go generate
```

**update dependencies in go.mod**
```
go mod init github.com/n-boy/backuper
go mod tidy
```


*Tested on Mac OS X, MS Windows 7*