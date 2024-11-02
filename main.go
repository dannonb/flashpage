package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	// "path"
)

func createNextApp(projectName string) (string, error) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "nextjs-"+projectName)
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Run `npx create-next-app` to initialize the app
	cmd := exec.Command("npx", "create-next-app", "--yes", projectName, "--use-npm") // Adjust flags as needed
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to initialize Next.js app: %w", err)
	}

	return tmpDir, nil
}

// zipProject compresses the generated Next.js project into a zip file.
func zipProject(sourceDir, outputFile string) error {
	zipFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name, _ = filepath.Rel(sourceDir, path)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}

		return err
	})

	return err
}


func CopyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Determine the new path in the destination directory
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		// If it's a directory, create it in the destination
		if info.IsDir() {
			if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
				return err
			}
		} else {
			// If it's a file, copy it to the destination
			if err := copyFile(path, dstPath); err != nil {
				return err
			}
		}

		return nil
	})
}

// copyFile copies a single file from src to dst
func copyFile(srcFile, dstFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	// Copy file permissions as well
	srcInfo, err := os.Stat(srcFile)
	if err != nil {
		return err
	}
	return os.Chmod(dstFile, srcInfo.Mode())
}

func main() {
	projectName := "user-nextjs-app"
	outputZip := projectName + ".zip"

	// Step 1: Create Next.js app in a temporary directory
	// projectPath, err := createNextApp(projectName)
	projectPath, err := os.MkdirTemp("", "temp")
	if err != nil {
		fmt.Println("Error making temp:", err)
		return
	}

	println(projectPath)

	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting working:", err)
		return
	}

	srcpath := filepath.Join(wd + "/templates/base")

	err = CopyDir(srcpath, projectPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer os.RemoveAll(projectPath) // Clean up after

	// Step 2: Zip the project for download
	if err := zipProject(projectPath, outputZip); err != nil {
		fmt.Println("Error zipping project:", err)
		return
	}

	fmt.Println("Project created and zipped successfully:", outputZip)
}