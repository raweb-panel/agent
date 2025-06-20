package user

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func CreateHome(serverName, username string) error {
	userHome := "/home/" + username
	domainDir := filepath.Join(userHome, serverName)

	// 1. Create system user if it doesn't exist, with shell disabled
	if !userExists(username) {
		cmd := exec.Command("useradd", "-m", "-d", userHome, "-s", "/usr/sbin/nologin", username)
		if err := cmd.Run(); err != nil {
			return errors.New("failed to create user: " + err.Error())
		}
	}

	// 2. Ensure user's home folder exists and set ownership
	if _, err := os.Stat(userHome); os.IsNotExist(err) {
		if err := os.MkdirAll(userHome, 0755); err != nil {
			return errors.New("failed to create home directory: " + userHome)
		}
	}
	exec.Command("chown", "-R", username+":"+username, userHome).Run()

	// 3. Create domain folder and subfolders
	subdirs := []string{"public_html", "logs", "tmp"}
	if _, err := os.Stat(domainDir); os.IsNotExist(err) {
		if err := os.MkdirAll(domainDir, 0755); err != nil {
			return errors.New("failed to create domain directory: " + domainDir)
		}
	}
	for _, subdir := range subdirs {
		path := filepath.Join(domainDir, subdir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0755); err != nil {
				return errors.New("failed to create directory: " + path)
			}
		}
	}
	exec.Command("chown", "-R", username+":"+username, domainDir).Run()

	// 4. Copy default config directory to /home/configs/$username/$serverName/config
	defaultConfig := "/raweb/web/panel/app/Helpers/defaults/config"
	targetConfigDir := filepath.Join("/home/configs", username, serverName, "config")
	if err := os.MkdirAll(targetConfigDir, 0755); err != nil {
		return errors.New("failed to create config directory: " + targetConfigDir)
	}
	if err := copyDir(defaultConfig, targetConfigDir); err != nil {
		return errors.New("failed to copy config directory: " + err.Error())
	}

	return nil
}

func userExists(username string) bool {
	cmd := exec.Command("id", username)
	err := cmd.Run()
	return err == nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		return copyFile(path, targetPath)
	})
}
