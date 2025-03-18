package unzip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func UnzipFile(zipPath, destPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
		return err
	}

	for _, file := range reader.File {
		filePath := filepath.Join(destPath, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		zipFile, err := file.Open()
		if err != nil {
			return err
		}
		defer zipFile.Close()

		destFile, err := os.Create(filePath)
		if err != nil {
			zipFile.Close()
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, zipFile)
		if err != nil {
			return err
		}
	}

	return nil
}
