# ksctl 远程命令工具

此工具用于向远端服务器发送下载文件和shell命令操作，只需配置yaml文件即可

#### yaml文件实例:

```yaml
server:
  ip: xx.xxx.xxx.xxx
  port: 22
  account: root
  password: xxxxxx
commands:
  - ls
  - ps
uploads:
  -
    src: ./ksctl.yaml
    dst: /root/ksctl.yaml
  -
    src: ./xxxx
    dst: /root/xxxx
downloads:
  -
    src: /root/ksctl.yaml
    dst: ./ksctl.yaml
  -
    src: ./xxxx
    dst: /root/xxxx
exec:
	name: /xxx/xxx
	exit: true
```

server为服务器配置

commands命令将会依次运行，每次条指令运行后将返回结果再运行吓一条命令

uploads和downloads为上传和下载文件

exec为想要运行的应用程序，name是绝对路径，exit代表当前程序退出时，远程程序是否退出

#### 使用：

```powershell
git clone https://github.com/hxoreyer/ksctl
go run ksctl.go -f [yaml文件地址(默认为./ksctl.yaml)]
```

##### Release:

```powershell
ksctl -f [yaml文件地址(默认为./ksctl.yaml)]
```

> 可将二进制程序放入环境变量path中

