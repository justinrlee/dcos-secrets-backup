// stolen from https://gist.github.com/josephspurrier/12cc5ed76d2228a41ceb

package cmd

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"fmt"
	"io"
	"os"
	// "path/filepath"

	"archive/tar"
	"bytes"
)

func decrypt(cipherstring string, keystring string) string {
	// Byte array of the string
	ciphertext := []byte(cipherstring)

	// Key
	key := []byte(keystring)

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// Before even testing the decryption,
	// if the text is too small, then it is incorrect
	if len(ciphertext) < aes.BlockSize {
		panic("Text is too short")
	}

	// Get the 16 byte IV
	iv := ciphertext[:aes.BlockSize]

	// Remove the IV from the ciphertext
	ciphertext = ciphertext[aes.BlockSize:]

	// Return a decrypted stream
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt bytes from ciphertext
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext)
}

func encrypt(plainstring, keystring string) string {
	// Byte array of the string
	plaintext := []byte(plainstring)

	// Key
	key := []byte(keystring)

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// Empty array of 16 + plaintext length
	// Include the IV at the beginning
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	// Slice of first 16 bytes
	iv := ciphertext[:aes.BlockSize]

	// Write 16 rand bytes to fill iv
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	// Return an encrypted stream
	stream := cipher.NewCFBEncrypter(block, iv)

	// Encrypt bytes from plaintext to ciphertext
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return string(ciphertext)
}

// Unused, for now.
// func readline() string {
// 	bio := bufio.NewReader(os.Stdin)
// 	line, _, err := bio.ReadLine()
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return string(line)
// }
// 
// func writeToFile(data, file string) {
// 	ioutil.WriteFile(file, []byte(data), 777)
// }

// func readFromFile(file string) ([]byte, error) {
// 	data, err := ioutil.ReadFile(file)
// 	return data, err
// }

func writeTar(files []Secret, filename string){
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Create a new tar archive.
	tw := tar.NewWriter(f)

	for _, file := range files {
		hdr := &tar.Header{
			Name: file.ID,
			Mode: 0600,
			Size: int64(len(file.EncryptedJSON)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			fmt.Println("Error writing header")
			// log.Fatalln(err)
		}
		if _, err := tw.Write([]byte(file.EncryptedJSON)); err != nil {
			fmt.Println("Error writing content")
			// log.Fatalln(err)
		}
	}
	// Make sure to check the error on Close.
	if err := tw.Close(); err != nil {
		fmt.Println("Error closing")
		// log.Fatalln(err)
	}
}

func readTar(filename string) (files []Secret){
	files = []Secret{}
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	tr := tar.NewReader(f)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			fmt.Println("Error advancing")
			// log.Fatalln(err)
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(tr)
		s := buf.String()
		files = append(files, Secret{ID: hdr.Name, EncryptedJSON: s})
	}
	return files
}

// func createDirFor(path string) {
// 	dir := filepath.Dir(path)
// 	// fmt.Println(dir)
// 	os.MkdirAll(dir, os.ModePerm)
// }

func validateCipher() {
	if cipherkey == "" {
		cipherkey = "ThisIsAMagicKeyString12345667890"
	} else if len(cipherkey)%32 != 0 {
		fmt.Printf("'cipherkey' has a length of %d characters\n", len(cipherkey))
		fmt.Println("It must be a multiple of 32 characters long")
		os.Exit(1)
	}
}