package ctl

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	Windows = "windows"
	Linux   = "linux"
	Macos   = "darwin"
)

type ClientConfig struct {
	Host       string
	Port       string
	Username   string
	Password   string
	sshClient  *ssh.Client
	sftpClient *sftp.Client
	LastResult string
}

func CreateClient(host string, port string, username, password string) *ClientConfig {
	cf := &ClientConfig{}
	var (
		sshClient  *ssh.Client
		sftpClient *sftp.Client
		err        error
	)

	cf.Host = host
	cf.Port = port
	cf.Username = username
	cf.Password = password

	config := ssh.ClientConfig{
		User: cf.Username,
		Auth: []ssh.AuthMethod{ssh.Password(cf.Password)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: 10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%s", cf.Host, cf.Port)

	if sshClient, err = ssh.Dial("tcp", addr, &config); err != nil {
		log.Fatalf("connect ssh error: %s", err.Error())
		return nil
	}

	cf.sshClient = sshClient

	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		log.Fatalf("connect sftp error: %s", err.Error())
		return nil
	}

	cf.sftpClient = sftpClient
	return cf
}

func (cf *ClientConfig) RunShell(shell string) string {
	var (
		session *ssh.Session
		err     error
	)

	if session, err = cf.sshClient.NewSession(); err != nil {
		log.Fatalf("new session error: %s", err.Error())
	}
	defer session.Close()

	var output bytes.Buffer
	session.Stdout = &output
	go func() {
		var tmp string
		for {
			tmp = output.String()
			if cf.LastResult != tmp {
				cmd := exec.Command("cmd", "/C", "cls")
				cmd.Stdout = os.Stdout
				cmd.Run()
				fmt.Println(tmp)
				cf.LastResult = tmp
			}
		}
	}()
	if err := session.Run(shell); err != nil {
		fmt.Println(shell)
		log.Fatalf("run shell error: %s", err.Error())
	} else {
		cf.LastResult = output.String()
	}
	return cf.LastResult
}

func (cf *ClientConfig) Upload(srcPath, dstPath string) {
	fmt.Println("searching...")
	s, err := os.Stat(srcPath)
	if err != nil {
		panic(err)
	}
	if s.IsDir() {
		cf.uploadDirectory(srcPath, dstPath)
	} else {
		cf.uploadFile(srcPath, dstPath)
	}
}

func (cf *ClientConfig) uploadDirectory(srcPath, dstPath string) {

	localFiles, err := ioutil.ReadDir(srcPath)
	if err != nil {
		panic(err)
	}
	cf.sftpClient.Mkdir(dstPath)

	for _, backupDir := range localFiles {
		localFilePath := path.Join(srcPath, backupDir.Name())
		remoteFilePath := path.Join(dstPath, backupDir.Name())

		if backupDir.IsDir() {
			cf.sftpClient.Mkdir(remoteFilePath)
			cf.uploadDirectory(localFilePath, remoteFilePath)
		} else {
			cf.uploadFile(localFilePath, remoteFilePath)
		}
	}
}

func (cf *ClientConfig) uploadFile(srcPath, dstPath string) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		panic("os.Open error : " + err.Error() + srcPath)
	}

	defer srcFile.Close()

	dstFile, err := cf.sftpClient.Create(dstPath)
	if err != nil {
		panic("sftpClient.Create error : " + err.Error())

	}

	defer dstFile.Close()
	ff, err := ioutil.ReadAll(srcFile)
	if err != nil {
		panic("ReadAll error : " + err.Error())
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		stat, _ := dstFile.Stat()
		state2, _ := srcFile.Stat()
		bar := KBar{}
		bar.New(stat.Size(), state2.Size(), dstFile.Name())
		for {
			stat, _ := dstFile.Stat()
			if stat.Size() < state2.Size() {
				bar.Play(stat.Size())
				time.Sleep(time.Duration(100 * time.Millisecond))
			} else {
				bar.Play(stat.Size())
				bar.Finish()
				break
			}
		}
		wg.Done()
	}()
	dstFile.Write(ff)
	dstFile.Chmod(0777)
	wg.Wait()

}

func (cf *ClientConfig) Download(srcPath, dstPath string) {
	srcFile, _ := cf.sftpClient.Open(srcPath)
	dstFile, _ := os.Create(dstPath)

	defer func() {
		_ = srcFile.Close()
		_ = dstFile.Close()
	}()

	if _, err := srcFile.WriteTo(dstFile); err != nil {
		log.Fatalf("download erroe: %s", err.Error())
	}

	fmt.Println("download success")
}

func (cf *ClientConfig) waitForQuit() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}

func BuildExec(path, system, name string) {
	var cmdStr string
	if system == Windows {
		cmdStr = fmt.Sprintf("cd %s&SET CGO_ENABLED=0&SET GOOS=%s&SET GOARCH=amd64&go build %s", path, system, name)
	} else {
		cmdStr = fmt.Sprintf("cd %s&CGO_ENABLED=0&GOOS=%s&GOARCH=amd64&go build %s", path, system, name)
	}
	cmd := exec.Command("cmd", "/C", cmdStr)
	_ = cmd.Run()
}

func (cf *ClientConfig) Run(procName string) {
	defer func() {
		pid := cf.RunShell(fmt.Sprintf("pgrep %s", path.Base(procName)))
		cf.RunShell(fmt.Sprintf("kill %s", pid))
	}()
	go cf.RunShell(procName)
	cf.waitForQuit()
}
