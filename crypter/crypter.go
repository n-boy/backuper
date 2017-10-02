package crypter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

type Encrypter struct {
	passphrase   string
	streamWriter *cipher.StreamWriter
}

func GetEncrypter(passphrase string) *Encrypter {
	return &Encrypter{passphrase: passphrase}
}

func (e *Encrypter) InitWriter(fileWriter io.Writer) (io.Writer, error) {
	if e.streamWriter != nil {
		return io.Writer(e), fmt.Errorf("Encrypt writer already inited")
	}
	encStream, iv, err := getEncryptStream(e.passphrase)
	if err != nil {
		return io.Writer(e), err
	}
	if _, err := fileWriter.Write(iv); err != nil {
		return io.Writer(e), err
	}
	e.streamWriter = &cipher.StreamWriter{S: encStream, W: fileWriter}
	return io.Writer(e), nil
}

func (e *Encrypter) Write(p []byte) (n int, err error) {
	if e.streamWriter == nil {
		return 0, fmt.Errorf("Encrypt writer should be inited before use")
	}
	return e.streamWriter.Write(p)
}

type Decrypter struct {
	passphrase   string
	iv           []byte
	streamReader *cipher.StreamReader
	streamWriter *cipher.StreamWriter
	fileWriter   *io.Writer
}

func GetDecrypter(passphrase string) *Decrypter {
	return &Decrypter{passphrase: passphrase}
}

func (d *Decrypter) InitReader(fileReader io.Reader) (io.Reader, error) {
	iv := make([]byte, aes.BlockSize)
	n, err := fileReader.Read(iv)
	if err != nil {
		return io.Reader(d), err
	} else if n != aes.BlockSize {
		return io.Reader(d), fmt.Errorf("Encrypted file is too small")
	}

	decStream, err := getDecryptStream(d.passphrase, iv)
	if err != nil {
		return io.Reader(d), err
	}
	d.streamReader = &cipher.StreamReader{S: decStream, R: fileReader}
	return io.Reader(d), nil
}

func (d *Decrypter) Read(p []byte) (n int, err error) {
	if d.streamReader == nil {
		return 0, fmt.Errorf("Encrypt reader should be inited before use")
	}
	return d.streamReader.Read(p)
}

func (d *Decrypter) InitWriter(fileWriter io.Writer) (io.Writer, error) {
	d.iv = make([]byte, 0)
	d.fileWriter = &fileWriter

	return io.Writer(d), nil
}

func (d *Decrypter) finishInitWriter() error {
	decStream, err := getDecryptStream(d.passphrase, d.iv)
	if err != nil {
		return err
	}
	d.streamWriter = &cipher.StreamWriter{S: decStream, W: *d.fileWriter}
	return nil
}

func (d *Decrypter) Write(p []byte) (n int, err error) {
	if d.fileWriter == nil {
		return 0, fmt.Errorf("Decrypt writer should be inited before use")
	}

	readPos := 0
	if d.streamWriter == nil {
		leftBytesForIV := aes.BlockSize - len(d.iv)
		if leftBytesForIV > 0 {
			bytesToRead := leftBytesForIV
			if bytesToRead > len(p) {
				bytesToRead = len(p)
			}
			d.iv = append(d.iv, p[0:bytesToRead]...)
			readPos += bytesToRead
		}

		if len(d.iv) == aes.BlockSize {
			err := d.finishInitWriter()
			if err != nil {
				return 0, err
			}
		}
	}

	if d.streamWriter != nil {
		n, err = d.streamWriter.Write(p[readPos:])
		return n + readPos, err
	} else {
		return readPos, nil
	}
}

func getEncryptStream(passphrase string) (stream cipher.Stream, iv []byte, err error) {
	enc_block, err := aes.NewCipher(getEncryptKey(passphrase))
	if err != nil {
		return
	}
	iv = make([]byte, aes.BlockSize)
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return
	}
	stream = cipher.NewOFB(enc_block, iv)
	return
}

func getDecryptStream(passphrase string, iv []byte) (stream cipher.Stream, err error) {
	enc_block, err := aes.NewCipher(getEncryptKey(passphrase))
	if err != nil {
		return
	}
	stream = cipher.NewOFB(enc_block, iv)
	return
}

func getEncryptKey(passphrase string) []byte {
	key_size := 32
	key := []byte(passphrase)
	for len(key) < key_size {
		key = append(key, 0)
	}
	return key[0:key_size]
}

func EncryptFile(passphrase string, srcFilePath string, targetFilePath string) error {
	convertAction := func(r *os.File, w *os.File) error {
		encrypter := GetEncrypter(passphrase)
		encrypter.InitWriter(w)
		_, err := io.Copy(encrypter, r)
		return err
	}

	return convertFileData(convertAction, srcFilePath, targetFilePath)
}

func DecryptFile(passphrase string, srcFilePath string, targetFilePath string) error {
	convertAction := func(r *os.File, w *os.File) error {
		decrypter := GetDecrypter(passphrase)
		decrypter.InitReader(r)
		_, err := io.Copy(w, decrypter)
		return err
	}

	return convertFileData(convertAction, srcFilePath, targetFilePath)
}

func convertFileData(convertAction func(r *os.File, w *os.File) error,
	srcFilePath string, targetFilePath string) error {

	srcReader, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer srcReader.Close()

	targetWriter, err2 := os.OpenFile(targetFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err2 != nil {
		return err2
	}
	defer targetWriter.Close()

	err = convertAction(srcReader, targetWriter)
	if err != nil {
		return err
	}

	return targetWriter.Close()
}
