package crypter_test

import (
	"github.com/n-boy/backuper/crypter"

	"github.com/n-boy/backuper/ut/testutils"

	"bufio"
	"bytes"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const passphrase = "my pretty long passphrase"
const passphrase2 = "my pretty long passphrase 2"
const plainText = `To be, or not to be: that is the question:
Whether 'tis nobler in the mind to suffer
The slings and arrows of outrageous fortune,
Or to take arms against a sea of troubles,
And by opposing end them? To die: to sleep;`

type DataCoincidePlace int

const (
	NONE DataCoincidePlace = iota
	BEGIN
	MIDDLE
	END
)

var places = [...]string{
	"none",
	"begin",
	"middle",
	"end",
}

func TestEncrypt(t *testing.T) {
	testName := "Encrypt"
	encText, err := encryptText(plainText, passphrase)
	if err != nil {
		t.Fatalf("Test died. Name: %v, error while encrypt: %v\n",
			testName, err)
	}

	if len(encText) < len(plainText) {
		t.Errorf("Test failed. Name: %v, encrypted text length expected: >= %v, got: %v\n",
			testName, len(plainText), len(encText))
	}

	coincidePlace := whereDataPartlyCoincide([]byte(plainText), encText)
	if coincidePlace != NONE {
		t.Errorf("Test failed. Name: %v, chunk from plain text on the %v contains in encrypted text\n",
			testName, places[coincidePlace])
	}
}

func TestEncryptDecryptPassphraseCorrect(t *testing.T) {
	testName := "EncryptDecryptPassphraseCorrect"
	encText, err := encryptText(plainText, passphrase)
	if err != nil {
		t.Fatalf("Test died. Name: %v, error while encrypt: %v\n",
			testName, err)
	}

	decryptedText, err := decryptText(encText, passphrase)
	if err != nil {
		t.Fatalf("Test died. Name: %v, error while decrypt: %v\n",
			testName, err)
	}
	if plainText != decryptedText {
		t.Errorf("Test failed. Name: %v, decrypted text not equals to source plain text, expected:\n %v, \n\r\ngot: %v\n",
			testName, plainText, decryptedText)
	}

}

func TestEncryptDecryptPassphraseIncorrect(t *testing.T) {
	testName := "EncryptDecryptPassphraseIncorrect"
	encText, err := encryptText(plainText, passphrase)
	if err != nil {
		t.Fatalf("Test died. Name: %v, error while encrypt: %v\n",
			testName, err)
	}

	decryptedText, err := decryptText(encText, passphrase2)
	if err != nil {
		t.Fatalf("Test died. Name: %v, error while decrypt: %v\n",
			testName, err)
	}

	if plainText == decryptedText {
		t.Errorf("Test failed. Name: %v, decrypted text equals to source plain text\n", testName)
	}

}

func TestEncryptDecryptUseWriter(t *testing.T) {
	testName := "EncryptDecryptUseWriter"
	encText, err := encryptText(plainText, passphrase)
	if err != nil {
		t.Fatalf("Test died. Name: %v, error while encrypt: %v\n",
			testName, err)
	}

	decryptedText, err := decryptTextUseWriter(encText, passphrase, false)
	if err != nil {
		t.Fatalf("Test died. Name: %v, error while decrypt: %v\n",
			testName, err)
	}
	if plainText != decryptedText {
		t.Errorf("Test failed. Name: %v, decrypted text not equals to source plain text, expected:\n %v, \n\r\ngot: %v\n",
			testName, plainText, decryptedText)
	}

}

// checks partially writing of IV cipher part
func TestEncryptDecryptUseWriterSmallChunks(t *testing.T) {
	testName := "EncryptDecryptUseWriterSmallChunks"
	encText, err := encryptText(plainText, passphrase)
	if err != nil {
		t.Fatalf("Test died. Name: %v, error while encrypt: %v\n",
			testName, err)
	}

	decryptedText, err := decryptTextUseWriter(encText, passphrase, true)
	if err != nil {
		t.Fatalf("Test died. Name: %v, error while decrypt: %v\n",
			testName, err)
	}
	if plainText != decryptedText {
		t.Errorf("Test failed. Name: %v, decrypted text not equals to source plain text, expected:\n %v, \n\r\ngot: %v\n",
			testName, plainText, decryptedText)
	}

}

func encryptText(plainText string, passphrase string) (encText []byte, err error) {
	plainTextReader := strings.NewReader(plainText)

	var encTextBuf bytes.Buffer
	encTextWriter := bufio.NewWriter(&encTextBuf)

	encrypter := crypter.GetEncrypter(passphrase)
	encrypter.InitWriter(encTextWriter)

	_, err = io.Copy(encrypter, plainTextReader)
	if err != nil {
		return
	}
	encTextWriter.Flush()
	encText = encTextBuf.Bytes()

	return
}

func decryptText(encText []byte, passphrase string) (plainText string, err error) {
	encTextReader := bytes.NewReader(encText)

	var plainTextBuf bytes.Buffer
	plainTextWriter := bufio.NewWriter(&plainTextBuf)

	decrypter := crypter.GetDecrypter(passphrase)
	decrypter.InitReader(encTextReader)

	_, err = io.Copy(plainTextWriter, decrypter)
	if err != nil {
		return
	}
	plainTextWriter.Flush()
	plainText = string(plainTextBuf.Bytes()[:])

	return
}

func decryptTextUseWriter(encText []byte, passphrase string, smallChunks bool) (plainText string, err error) {
	encTextReader := bytes.NewReader(encText)

	var plainTextBuf bytes.Buffer
	plainTextWriter := bufio.NewWriter(&plainTextBuf)

	decrypter := crypter.GetDecrypter(passphrase)
	decrypter.InitWriter(plainTextWriter)

	if smallChunks {
		buf := make([]byte, 5)
		for {
			n, err1 := encTextReader.Read(buf)
			if err1 != nil && err1 != io.EOF {
				return "", err1
			}

			if n == 0 {
				break
			}

			_, err = decrypter.Write(buf[:n])
			if err != nil {
				return
			}
		}

	} else {
		_, err = io.Copy(decrypter, encTextReader)
		if err != nil {
			return
		}
	}

	plainTextWriter.Flush()
	plainText = string(plainTextBuf.Bytes()[:])

	return
}

func whereDataPartlyCoincide(source []byte, target []byte) DataCoincidePlace {
	chunkSize := 32
	if len(source) < chunkSize {
		chunkSize = len(source)
	}
	if chunkIsContained(source[0:chunkSize], target) {
		return BEGIN
	}

	middle_idx := len(source) / 2
	if len(source)-middle_idx < chunkSize {
		middle_idx -= chunkSize - (len(source) - middle_idx)
	}
	if chunkIsContained(source[middle_idx:middle_idx+chunkSize], target) {
		return MIDDLE
	}

	if chunkIsContained(source[len(source)-chunkSize:len(source)], target) {
		return END
	}

	return NONE
}

func chunkIsContained(chunk []byte, target []byte) bool {
	if len(target) < len(chunk) {
		return false
	}

	chunkHex := hex.EncodeToString(chunk)
	for i := 0; i < len(target)-len(chunk); i++ {
		if chunkHex == hex.EncodeToString(target[i:i+len(chunk)]) {
			return true
		}
	}

	return false
}

func TestEncryptDecryptFile(t *testing.T) {
	testName := "EncryptDecryptFile"

	srcFilePath := filepath.Join(testutils.TmpDir(), testutils.RandString(20)+".bin")
	t.Logf("Creating temporary file %s", srcFilePath)
	err := ioutil.WriteFile(srcFilePath, []byte(plainText), 0600)
	defer os.Remove(srcFilePath)
	if err != nil {
		t.Fatalf("Test died. Error while creating plain text file: %v", err)
	}

	srcFileMd5, err1 := testutils.CalcFileMD5(srcFilePath)
	if err1 != nil {
		t.Fatalf("Test died. Name: %v. Error while calc plain text file md5: %v", testName, err1)
	}

	encFilePath := filepath.Join(testutils.TmpDir(), testutils.RandString(20)+".bin")
	err = crypter.EncryptFile(passphrase, srcFilePath, encFilePath)
	defer os.Remove(encFilePath)
	if err != nil {
		t.Fatalf("Test died. Name: %v. Error while encrypting file: %v", testName, err)
	}

	newSrcFileMd5, err2 := testutils.CalcFileMD5(srcFilePath)
	if err2 != nil {
		t.Fatalf("Test died. Name: %v. Error while calc plain text file md5 after encryption: %v", testName, err2)
	}
	if newSrcFileMd5 != srcFileMd5 {
		t.Errorf("Test failed. Name: %v. Plain text md5 changed after encryption", testName)
	}

	encFileMd5, err3 := testutils.CalcFileMD5(encFilePath)
	if err3 != nil {
		t.Fatalf("Test died. Name: %v. Error while calc encrypted file md5: %v", testName, err3)
	}
	if encFileMd5 == srcFileMd5 {
		t.Errorf("Test failed. Name: %v. Plain text md5 equals to encrypted file md5", testName)
	}

	decryptedFilePath := filepath.Join(testutils.TmpDir(), testutils.RandString(20)+".bin")
	err = crypter.DecryptFile(passphrase, encFilePath, decryptedFilePath)
	defer os.Remove(decryptedFilePath)
	if err != nil {
		t.Fatalf("Test died. Name: %v. Error while decrypting file: %v", testName, err)
	}

	newEncFileMd5, err4 := testutils.CalcFileMD5(encFilePath)
	if err4 != nil {
		t.Fatalf("Test died. Name: %v. Error while calc encrypted text file md5 after decryption: %v", testName, err4)
	}
	if newEncFileMd5 != encFileMd5 {
		t.Errorf("Test failed. Name: %v. Encrypted text md5 changed after decryption", testName)
	}

	decryptedFileMd5, err5 := testutils.CalcFileMD5(decryptedFilePath)
	if err5 != nil {
		t.Fatalf("Test died. Name: %v. Error while calc decrypted text file md5: %v", testName, err5)
	}
	if decryptedFileMd5 != srcFileMd5 {
		t.Errorf("Test failed. Name: %v. Decrypted text md5 not equals source plain text file md5", testName)
	}

}
