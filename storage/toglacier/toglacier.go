package toglacier

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/glacier"

	"github.com/n-boy/backuper/base"
	"github.com/n-boy/backuper/crypter"
)

type GlacierStorage struct {
	region                string `name:"region" title:"AWS Region"`
	vault_name            string `name:"vault_name" title:"Vault Name"`
	aws_access_key_id     string `name:"aws_access_key_id" title:"Access Key ID"`
	aws_secret_access_key string `name:"aws_secret_access_key" title:"Secret Access Key"`
}

type GlacierFileInfo struct {
	ArchiveId          string
	ArchiveDescription string
	CreationDate       string
	Size               int64
	SHA256TreeHash     string
}

func NewStorage(config map[string]string) (GlacierStorage, error) {
	var gs GlacierStorage

	gs.region = config["region"]
	gs.vault_name = config["vault_name"]
	gs.aws_access_key_id = config["aws_access_key_id"]
	gs.aws_secret_access_key = config["aws_secret_access_key"]

	return gs, nil
}

func GetEmptyStorage() GlacierStorage {
	return GlacierStorage{}
}

func (gs GlacierStorage) GetType() string {
	return "glacier"
}

func (gs GlacierStorage) GetStorageConfig() map[string]string {
	config := make(map[string]string)

	config["region"] = gs.region
	config["vault_name"] = gs.vault_name
	config["aws_access_key_id"] = gs.aws_access_key_id
	config["aws_secret_access_key"] = gs.aws_secret_access_key

	return config
}

func (gs GlacierStorage) UploadFile(filePath string, encrypter *crypter.Encrypter, remoteFileName string) (map[string]string, error) {
	result := make(map[string]string)

	fileReader, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer fileReader.Close()

	filename := filepath.Base(filePath)
	if remoteFileName != "" {
		filename = filepath.Base(remoteFileName)
	}

	params := &glacier.UploadArchiveInput{
		AccountId:          aws.String("-"),
		VaultName:          aws.String(gs.vault_name),
		ArchiveDescription: aws.String(filename),
		Body:               fileReader}

	uploadResult, err := gs.getStorageClient().UploadArchive(params)
	if err != nil {
		return result, err
	}
	err = fileReader.Close()
	if err != nil {
		return result, err
	}

	result["ArchiveId"] = aws.StringValue(uploadResult.ArchiveId)
	result["Checksum"] = aws.StringValue(uploadResult.Checksum)
	result["Location"] = aws.StringValue(uploadResult.Location)
	return result, nil
}

func (gs GlacierStorage) DownloadFile(fileStorageId map[string]string, localFilePath string,
	decrypter *crypter.Decrypter) error {
	activeJob, err := gs.findJob(glacier.ActionCodeArchiveRetrieval, fileStorageId["ArchiveId"])
	if err != nil {
		return err
	}

	if activeJob == nil {
		params := &glacier.InitiateJobInput{
			AccountId: aws.String("-"),
			VaultName: aws.String(gs.vault_name),
			JobParameters: &glacier.JobParameters{
				ArchiveId: aws.String(fileStorageId["ArchiveId"]),
				Type:      aws.String("archive-retrieval"),
			},
		}
		result, err := gs.getStorageClient().InitiateJob(params)
		if err != nil {
			return err
		}
		activeJob, err = gs.getJob(result.JobId)
		if err != nil {
			return err
		}
	}

	activeJob, err = gs.waitJobComplete(activeJob, 0, 1)
	if err != nil {
		return err
	}

	params := &glacier.GetJobOutputInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(gs.vault_name),
		JobId:     activeJob.JobId}
	result, err := gs.getStorageClient().GetJobOutput(params)
	if err != nil {
		return err
	}

	localFilePathShadow := fmt.Sprint(localFilePath, "~")

	localFileWriter, err := os.Create(localFilePathShadow)
	if err != nil {
		return err
	}
	defer localFileWriter.Close()

	bufWriter := bufio.NewWriterSize(localFileWriter, 16*1024*1024)

	buf := make([]byte, 4*1024*1024)
	for {
		n, err := result.Body.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		_, err = bufWriter.Write(buf[:n])
		if err != nil {
			return err
		}
	}
	if err = bufWriter.Flush(); err != nil {
		return err
	} else if err = localFileWriter.Close(); err != nil {
		return err
	}
	err = os.Rename(localFilePathShadow, localFilePath)
	return err
}

func (gs GlacierStorage) DeleteFile(fileStorageInfo map[string]string) error {
	params := &glacier.DeleteArchiveInput{
		AccountId: aws.String("-"),
		ArchiveId: aws.String(fileStorageInfo["ArchiveId"]),
		VaultName: aws.String(gs.vault_name)}

	_, err := gs.getStorageClient().DeleteArchive(params)
	return err
}

func (gs GlacierStorage) GetFilesList() ([]base.GenericStorageFileInfo, error) {
	var filesList []base.GenericStorageFileInfo
	activeJob, err := gs.findJob(glacier.ActionCodeInventoryRetrieval, "")
	if err != nil {
		return filesList, err
	}

	if activeJob == nil {
		params := &glacier.InitiateJobInput{
			AccountId: aws.String("-"),
			VaultName: aws.String(gs.vault_name),
			JobParameters: &glacier.JobParameters{
				Type: aws.String("inventory-retrieval"),
			},
		}
		result, err := gs.getStorageClient().InitiateJob(params)
		if err != nil {
			return filesList, err
		}
		activeJob, err = gs.getJob(result.JobId)
		if err != nil {
			return filesList, err
		}
	}

	activeJob, err = gs.waitJobComplete(activeJob, 0, 1)
	if err != nil {
		return filesList, err
	}

	params := &glacier.GetJobOutputInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(gs.vault_name),
		JobId:     activeJob.JobId}
	result, err := gs.getStorageClient().GetJobOutput(params)
	if err != nil {
		return filesList, err
	}

	type jsonJobResp struct {
		ArchiveList []GlacierFileInfo
	}
	var jobResp jsonJobResp
	jsonResp, err := ioutil.ReadAll(result.Body)
	if err == nil {
		err = json.Unmarshal(jsonResp, &jobResp)
	}
	if err != nil {
		return filesList, err
	}
	for _, gfi := range jobResp.ArchiveList {
		filesList = append(filesList, base.GenericStorageFileInfo(gfi))
	}
	return filesList, nil
}

func (gs GlacierStorage) findJob(jobAction string, archiveId string) (*glacier.JobDescription, error) {
	params := &glacier.ListJobsInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(gs.vault_name),
		Limit:     aws.String("1000000")}
	jobsList, err := gs.getStorageClient().ListJobs(params)
	if err != nil {
		return nil, err
	}
	var activeJob *glacier.JobDescription
	for _, job := range jobsList.JobList {
		if *job.StatusCode == glacier.StatusCodeFailed ||
			*job.Action != jobAction {
			continue
		}

		switch {
		case jobAction == glacier.ActionCodeInventoryRetrieval:
			activeJob = job
		case jobAction == glacier.ActionCodeArchiveRetrieval:
			if *job.ArchiveId == archiveId {
				activeJob = job
			}
		}
	}
	return activeJob, nil
}

func (gs GlacierStorage) getJob(jobId *string) (*glacier.JobDescription, error) {
	params := &glacier.DescribeJobInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(gs.vault_name),
		JobId:     jobId,
	}
	result, err := gs.getStorageClient().DescribeJob(params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (gs GlacierStorage) waitJobComplete(job *glacier.JobDescription, timeout int, repeatPause int) (*glacier.JobDescription, error) {
	var err error
	timeStart := time.Now()
	for *job.Completed == false {
		time.Sleep(time.Duration(repeatPause) * time.Second)
		if time.Since(timeStart).Seconds() > float64(timeout) {
			break
		}
		job, err = gs.getJob(job.JobId)
		if err != nil {
			return nil, err
		}
	}

	if *job.Completed == true {
		if *job.StatusCode == glacier.StatusCodeFailed {
			return job, errors.New(*job.StatusMessage)
		}
		return job, nil
	}
	return job, base.ErrStorageRequestInProgress

}

// func (gs GlacierStorage) syncMetaInfo() error {

// }

func (gs GlacierStorage) getStorageClient() *glacier.Glacier {
	creds := credentials.NewStaticCredentials(gs.aws_access_key_id, gs.aws_secret_access_key, "")

	return glacier.New(&aws.Config{
		Region:      aws.String(gs.region),
		Credentials: creds,
		LogLevel:    aws.LogLevel(1),
	})
}

// func (gs GlacierStorage) GetFileIface(ffm map[string]string) base.GenericStorageFileInfo {
// 	gfi := GlacierFileInfo{
// 		ArchiveId:          ffm["ArchiveId"],
// 		ArchiveDescription: ffm["ArchiveDescription"],
// 		CreationDate:       ffm["CreationDate"],
// 		SHA256TreeHash:     ffm["SHA256TreeHash"],
// 	}
// 	var err error
// 	gfi.Size, err = strconv.ParseInt(ffm["Size"], 10, 64)
// 	if err != nil {
// 		base.LogErr.Fatalln(err)
// 	}
// 	return gfi
// }

// func (gfi GlacierFileInfo) GetFlatMap() map[string]string {
// 	ffm := make(map[string]string)
// 	ffm["ArchiveId"] = gfi.ArchiveId
// 	ffm["ArchiveDescription"] = gfi.ArchiveDescription
// 	ffm["CreationDate"] = gfi.CreationDate
// 	ffm["SHA256TreeHash"] = gfi.SHA256TreeHash
// 	ffm["Size"] = strconv.FormatInt(gfi.Size, 10)
// 	return ffm
// }

func (gfi GlacierFileInfo) GetFilename() string {
	return gfi.ArchiveDescription
}

func (gfi GlacierFileInfo) GetFileStorageId() map[string]string {
	return map[string]string{"ArchiveId": gfi.ArchiveId}
}
