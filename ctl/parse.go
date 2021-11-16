package ctl

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Yaml struct {
	Server        ServerAddr `yaml:"server"`
	UploadFiles   []FileCtl  `yaml:"uploads"`
	DownloadFiles []FileCtl  `yaml:"downloads"`
	Commands      []string   `yaml:"commands"`
	Exec          ExecShell  `yaml:"exec"`
}

type ServerAddr struct {
	IP       string `yaml:"ip"`
	Port     string `yaml:"port"`
	Account  string `yaml:"account"`
	Password string `yaml:"password"`
}

type FileCtl struct {
	Src string `yaml:"src"`
	Dst string `yaml:"dst"`
}

type ExecShell struct {
	ProcName string `yaml:"name"`
	Exit     bool   `yaml:"exit"`
}

func NewYaml(path string) *Yaml {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	yl := &Yaml{}
	err = yaml.Unmarshal(src, yl)
	if err != nil {
		panic(err)
	}
	return yl
}

func (yal Yaml) Run() {

	srv := CreateClient(yal.Server.IP, yal.Server.Port, yal.Server.Account, yal.Server.Password)
	for _, v := range yal.UploadFiles {
		srv.Upload(v.Src, v.Dst)
	}
	for _, v := range yal.DownloadFiles {
		srv.Download(v.Src, v.Dst)
	}
	for _, v := range yal.Commands {
		fmt.Println(srv.RunShell(v))
	}
	if yal.Exec.Exit {
		srv.Run(yal.Exec.ProcName)
	} else {
		srv.RunShell(yal.Exec.ProcName)
	}
}
