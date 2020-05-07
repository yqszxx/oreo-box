package image

import (
	"fmt"
	"github.com/yqszxx/oreo-box/config"
	"golang.org/x/crypto/openpgp"
	"os"
	"os/exec"
	"path"
	"strings"
)

func Import(imageName, imageFilePath string) error {
	imageTempDir := path.Join(config.ImageTempPath, imageName)
	if err := os.MkdirAll(imageTempDir, 0755); err != nil {
		return fmt.Errorf("cannot create temp image dir `%s`: %v", imageTempDir, err)
	}
	defer func() {
		if err := os.RemoveAll(imageTempDir); err != nil {
			panic(err)
		}
	}()

	if _, err := exec.Command("tar", "-xvf", imageFilePath, "-C", imageTempDir).CombinedOutput(); err != nil {
		return fmt.Errorf("fail to untar `%s` to `%s`: %v", imageFilePath, imageTempDir, err)
	}

	keyRingReader := strings.NewReader(config.PublicKey)

	signatureFilePath := path.Join(config.ImageTempPath, imageName, config.SignatureFileName)
	signatureFile, err := os.Open(signatureFilePath)
	if err != nil {
		return fmt.Errorf("cannot open signature file `%s`: %v", signatureFilePath, err)
	}
	defer func() {
		if err := signatureFile.Close(); err != nil {
			panic(err)
		}
	}()

	imageDataFilePath := path.Join(config.ImageTempPath, imageName, config.ImageDataFileName)
	imageDataFile, err := os.Open(imageDataFilePath)
	if err != nil {
		return fmt.Errorf("cannot open image data file `%s`: %v", imageDataFilePath, err)
	}
	defer func() {
		if err := imageDataFile.Close(); err != nil {
			panic(err)
		}
	}()

	keyring, err := openpgp.ReadArmoredKeyRing(keyRingReader)
	if err != nil {
		return fmt.Errorf("cannot read public key ring: %v", err)
	}

	if _, err := openpgp.CheckArmoredDetachedSignature(keyring, imageDataFile, signatureFile); err != nil {
		return fmt.Errorf("cannot verify image signature: %v", err)
	}

	// signature valid, untar image to image store
	imageDataDir := path.Join(config.ImagePath, imageName)
	if err := os.MkdirAll(imageDataDir, 0755); err != nil {
		return fmt.Errorf("cannot create image storage dir `%s`: %v", imageDataDir, err)
	}

	if _, err := exec.Command("tar", "-xvf", imageDataFilePath, "-C", imageDataDir).CombinedOutput(); err != nil {
		return fmt.Errorf("fail to untar `%s` to `%s`: %v", imageDataFilePath, imageDataDir, err)
	}

	// import success
	fmt.Println(imageName)

	return nil
}
