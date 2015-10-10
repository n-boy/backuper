package crypter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
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
	streamReader *cipher.StreamReader
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
