package main

import (
	"fmt"
	"os"
	"strconv"

	sysproxy "github.com/black-06/sysproxy-go"
)

func printUsageThenExit() {
	fmt.Println(`Usage: 
sysproxy global <proxy-server> [<bypass-list>]
      bypass list is a string like: localhost;127.*;10.* without trailing semicolon.
sysproxy pac <pac-url>
sysproxy query
sysproxy set <flags> [<proxy-server> [<bypass-list> [<pac-url>]]]
      <flags> is bitwise combination of INTERNET_PER_CONN_FLAGS.
      "-" is a placeholder to keep the original value.`)
	os.Exit(1)
}

func getArgs(idx int) string {
	if idx < len(os.Args) {
		arg := os.Args[idx]
		// "-" is a placeholder to keep the original value.
		if arg == "-" {
			return ""
		}
		return arg
	}
	return ""
}

func main() {
	argc := len(os.Args)
	if argc < 2 {
		printUsageThenExit()
	}
	switch os.Args[1] {
	case "global":
		if argc < 3 {
			printUsageThenExit()
		}
		if argc > 4 {
			_, _ = fmt.Fprintln(os.Stderr, "Error: bypass list shouldn't contain spaces, please check parameters")
			printUsageThenExit()
		}
		status := sysproxy.InternetStatus{
			ProxyType:   sysproxy.ProxyTypeDirect | sysproxy.ProxyTypeProxy,
			ProxyServer: os.Args[2],
		}
		if argc == 4 {
			status.ProxyBypass = os.Args[3]
		} else {
			status.ProxyBypass = "<local>"
		}
		if err := sysproxy.InternetSet(status); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "pac":
		if argc != 3 {
			printUsageThenExit()
		}
		if err := sysproxy.InternetSet(sysproxy.InternetStatus{
			ProxyType: sysproxy.ProxyTypeDirect | sysproxy.ProxyTypeAutoProxyUrl,
			ConfigUrl: os.Args[2],
		}); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "query":
		status, err := sysproxy.InternetQuery()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(status.ProxyType)
		fmt.Println(status.ProxyServer)
		fmt.Println(status.ProxyBypass)
		fmt.Println(status.ConfigUrl)
	case "set":
		if argc < 3 || argc > 6 {
			printUsageThenExit()
		}
		flags, err := strconv.Atoi(os.Args[2])
		if err != nil || flags > 0x0F || flags < 1 {
			_, _ = fmt.Fprintln(os.Stderr, "Error: flags is not accepted")
			printUsageThenExit()
		}
		if err = sysproxy.InternetSet(sysproxy.InternetStatus{
			ProxyType:   sysproxy.ProxyType(flags),
			ProxyServer: getArgs(3),
			ProxyBypass: getArgs(4),
			ConfigUrl:   getArgs(5),
		}); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		printUsageThenExit()
	}
}
