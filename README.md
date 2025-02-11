# sysproxy-go

System Proxy Agent for Windows, implemented with go.
It has the same behavior as https://github.com/shadowsocks/sysproxy, it is also a go lib for get/set system proxy, only
windows supported.

## Install

```shell
go install github.com/black-06/sysproxy-go/cmd@latest
```

Or download from release page.

## Usage

- global
  ```shell
  sysproxy global <proxy-server> [<bypass-list>]
  ```
  bypass list is a string like: `localhost;127.*;10.*` without trailing semicolon.
- pac
  ```shell
  sysproxy pac <pac-url>
  ```
- query
  ```shell
  sysproxy query
  ```
  output: `<flags> \n <proxy-server> \n <bypass-list> \n <pac-url>`, 
- set
  ```shell
  sysproxy set <flags> [<proxy-server> [<bypass-list> [<pac-url>]]]
  ```
  flags is bitwise combination of INTERNET_PER_CONN_FLAGS.
  "-" is a placeholder to keep the original value.
