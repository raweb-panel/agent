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

	// 1. Create user if missing
	if !userExists(username) {
		cmd := exec.Command("useradd", "-m", "-d", userHome, "-s", "/usr/sbin/nologin", username)
		if err := cmd.Run(); err != nil {
			return errors.New("failed to create user: " + err.Error())
		}
	}

	// 2. Create home dir if missing
	if _, err := os.Stat(userHome); os.IsNotExist(err) {
		if err := os.MkdirAll(userHome, 0755); err != nil {
			return errors.New("failed to create home directory: " + userHome)
		}
	}
	exec.Command("chown", "-R", username+":"+username, userHome).Run()

	// 3. If domain dir exists, return error
	if _, err := os.Stat(domainDir); err == nil {
		return errors.New("domain_exists")
	}

	// 4. Create domain dir and subdirs
	subdirs := []string{"public_html", "logs", "tmp"}
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		return errors.New("failed to create domain directory: " + domainDir)
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

	// 5. Copy config
	defaultConfig := "/raweb/apps/raweb/panel/app/Helpers/defaults/config"
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
