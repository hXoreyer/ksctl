package ctl

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
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
	srcFile, _ := os.Open(srcPath)
	dstFile, _ := cf.sftpClient.Create(dstPath)

	defer func() {
		_ = srcFile.Close()
		_ = dstFile.Close()
	}()

	buffer := bytes.Buffer{}
	buf := make([]byte, 1024)
	for {
		n, err := srcFile.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Fatalf("read file error: %s", err.Error())
			} else {
				break
			}
		}
		buffer.Write(buf[:n])
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		str := "Uploading: ="
		for {
			stat, _ := dstFile.Stat()
			state2, _ := srcFile.Stat()
			if stat.Size() <= state2.Size() {
				var ft float64
				ft = float64(stat.Size()) / float64(state2.Size())
				for i := 0; i < int(ft*100)/2; i++ {
					str += "="
				}
				str += ">"
				fmt.Printf("%s %d%%\n", str, int64(ft*100))
				str = "Upload: ="
				if stat.Size() == state2.Size() {
					fmt.Println("Upload complete")
					break
				}
			} else {
				break
			}
			time.Sleep(time.Millisecond * 200)
		}
		wg.Done()
	}()
	_, _ = dstFile.Write(buffer.Bytes())
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

func (cf *ClientConfig) waitForQuit(exitCode string) {
	var s string
	for {
		fmt.Scanf("%s", &s)
		if s == exitCode {
			break
		} else {
			fmt.Println(s)
		}
	}
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

func (cf *ClientConfig) Run(exitCode, procName string) {
	defer func() {
		pid := cf.RunShell(fmt.Sprintf("pgrep %s", procName))
		cf.RunShell(fmt.Sprintf("kill %s", pid))
	}()
	cf.waitForQuit(exitCode)
}
