package toglacier

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glacier"

	"github.com/n-boy/backuper/base"
	storageutils "github.com/n-boy/backuper/storage/utils"
)

// must be power of two
const MultipartUploadPartSize = 32 * 1024 * 1024

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

func (gs GlacierStorage) UploadFile(filePath string, remoteFileName string) (map[string]string, error) {
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

	fileInfo, err := fileReader.Stat()
	if err != nil {
		return result, err
	}

	checksum := hex.EncodeToString(glacier.ComputeHashes(fileReader).TreeHash)

	// initiate multipart upload
	initUploadParams := &glacier.InitiateMultipartUploadInput{
		AccountId:          aws.String("-"),
		VaultName:          aws.String(gs.vault_name),
		ArchiveDescription: aws.String(filename),
		PartSize:           aws.String(strconv.FormatInt(MultipartUploadPartSize, 10))}
	// panic(fmt.Errorf("%v", initUploadParams).Error())
	initUploadResult, err := gs.getStorageClient().InitiateMultipartUpload(initUploadParams)
	if err != nil {
		return result, err
	}

	abortUploadParams := &glacier.AbortMultipartUploadInput{
		AccountId: initUploadParams.AccountId,
		UploadId:  initUploadResult.UploadId,
		VaultName: initUploadParams.VaultName,
	}

	// upload archive parts, one-by-one
	numOfParts := int64(math.Ceil(float64(fileInfo.Size()) / float64(MultipartUploadPartSize)))

	for i := int64(0); i < numOfParts; i++ {
		rangeStart := i * MultipartUploadPartSize
		rangeFinish := (i+1)*MultipartUploadPartSize - 1
		if rangeFinish >= fileInfo.Size() {
			rangeFinish = fileInfo.Size() - 1
		}

		fileSectionReader := io.NewSectionReader(fileReader, rangeStart, rangeFinish-rangeStart+1)

		uploadPartParams := &glacier.UploadMultipartPartInput{
			AccountId: initUploadParams.AccountId,
			UploadId:  initUploadResult.UploadId,
			VaultName: initUploadParams.VaultName,
			Body:      fileSectionReader,
			Range:     aws.String(fmt.Sprintf("bytes %d-%d/*", rangeStart, rangeFinish)),
		}

		base.Log.Println(fmt.Sprintf("Start uploading of part %d, range: %d - %d", i+1, rangeStart, rangeFinish))
		t0 := time.Now().Unix()
		_, err := gs.getStorageClient().UploadMultipartPart(uploadPartParams)
		if err != nil {
			// errors ignoring upload aborting
			_, err2 := gs.getStorageClient().AbortMultipartUpload(abortUploadParams)
			if err2 != nil {
				base.LogErr.Printf("Error while aborting multipart upload: %v", err2)
			}
			return result, err
		} else {
			speed := (rangeFinish - rangeStart) / (time.Now().Unix() - t0 + 1) / 1024 * 8
			base.Log.Println(fmt.Sprintf("Uploaded part %d of %d (%d KBit/s)", i+1, numOfParts, speed))
		}
	}

	// complete multipart upload
	completeUploadParams := &glacier.CompleteMultipartUploadInput{
		AccountId:   initUploadParams.AccountId,
		UploadId:    initUploadResult.UploadId,
		VaultName:   initUploadParams.VaultName,
		ArchiveSize: aws.String(strconv.FormatInt(fileInfo.Size(), 10)),
		Checksum:    aws.String(checksum),
	}

	uploadResult, err := gs.getStorageClient().CompleteMultipartUpload(completeUploadParams)
	if err != nil {
		// errors ignoring upload aborting
		_, err2 := gs.getStorageClient().AbortMultipartUpload(abortUploadParams)
		if err2 != nil {
			base.LogErr.Printf("Error while aborting multipart upload: %v", err2)
		}

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

func (gs GlacierStorage) DownloadFile(fileStorageId map[string]string, localFilePath string) error {
	downloadAction := func(pipe io.Writer) error {
		return gs.DownloadFileToPipe(fileStorageId, pipe)
	}
	return storageutils.DownloadFile(downloadAction, localFilePath)
}

func (gs GlacierStorage) DownloadFileToPipe(fileStorageId map[string]string, pipe io.Writer) error {

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

	bufWriter := bufio.NewWriterSize(pipe, 16*1024*1024)

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
	}
	return nil
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
		if *job.StatusCode == glacier.StatusCodeFailed &&
			!(jobAction == glacier.ActionCodeArchiveRetrieval && *job.StatusMessage == "ARCHIVE_DELETED") ||
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
	for *job.Completed == false && *job.StatusCode != glacier.StatusCodeFailed {
		time.Sleep(time.Duration(repeatPause) * time.Second)
		if time.Since(timeStart).Seconds() > float64(timeout) {
			break
		}
		job, err = gs.getJob(job.JobId)
		if err != nil {
			return nil, err
		}
	}

	if *job.Completed == true || *job.StatusCode == glacier.StatusCodeFailed {
		if *job.StatusCode == glacier.StatusCodeFailed {
			return job, errors.New(*job.StatusMessage)
		}
		return job, nil
	}
	return job, base.ErrStorageRequestInProgress

}

func (gs GlacierStorage) getStorageClient() *glacier.Glacier {
	creds := credentials.NewStaticCredentials(gs.aws_access_key_id, gs.aws_secret_access_key, "")

	return glacier.New(session.Must(session.NewSession()),
		&aws.Config{
			Region:      aws.String(gs.region),
			Credentials: creds,
			LogLevel:    aws.LogLevel(1),
		})
}

func (gfi GlacierFileInfo) GetFilename() string {
	return gfi.ArchiveDescription
}

func (gfi GlacierFileInfo) GetFileStorageId() map[string]string {
	return map[string]string{"ArchiveId": gfi.ArchiveId}
}
