package sysproxy_go

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	modWininet = syscall.NewLazyDLL("wininet.dll")
	modRAS     = syscall.NewLazyDLL("rasapi32.dll")

	// https://learn.microsoft.com/zh-cn/windows/win32/api/wininet/nf-wininet-internetsetoptionw
	procInternetSetOptionW = modWininet.NewProc("InternetSetOptionW")
	// https://learn.microsoft.com/zh-cn/windows/win32/api/wininet/nf-wininet-internetqueryoptionw
	procInternetQueryOptionW = modWininet.NewProc("InternetQueryOptionW")
	// https://learn.microsoft.com/zh-cn/windows/win32/api/ras/nf-ras-rasenumentriesw
	procRasEnumEntriesW = modRAS.NewProc("RasEnumEntriesW")
)

func InternetSet(status InternetStatus) error {
	list, err := internetStatusToInternetPerConnOptionListW(status)
	if err != nil {
		return err
	}

	// Set LAN
	list.pszConnection = 0
	if err := internetSet(list); err != nil {
		return err
	}

	// Find connections and apply proxy settings
	var names []rasEntryNameW
	bufSize, nameSize := 0, 0
	rst, _, err := syscall.SyscallN(procRasEnumEntriesW.Addr(), 0, 0, 0, uintptr(unsafe.Pointer(&bufSize)), uintptr(unsafe.Pointer(&nameSize)))
	if rst == errorBufferTooSmall {
		names = make([]rasEntryNameW, nameSize)
		names[0].dwSize = sizeRASEntryNameW
		rst, _, err = syscall.SyscallN(procRasEnumEntriesW.Addr(), 0, 0, uintptr(unsafe.Pointer(&names[0])), uintptr(unsafe.Pointer(&bufSize)), uintptr(unsafe.Pointer(&nameSize)))
	}
	if rst != 0 {
		return os.NewSyscallError("RasEnumEntriesW", err)
	}
	for _, name := range names {
		list.pszConnection = uintptr(unsafe.Pointer(&name.szEntryName[0]))
		if setErr := internetSet(list); setErr != nil {
			return err
		}
	}
	return nil
}

func internetSet(list *internetPerConnOptionListW) error {
	rst, _, err := syscall.SyscallN(procInternetSetOptionW.Addr(), 0, internetOptionPerConnectionOption, uintptr(unsafe.Pointer(list)), uintptr(unsafe.Pointer(&list.dwSize)))
	if rst == 0 {
		return os.NewSyscallError("InternetSetOptionW(PER_CONNECTION_OPTION)", err)
	}
	rst, _, err = syscall.SyscallN(procInternetSetOptionW.Addr(), 0, internetOptionProxySettingsChanged, 0, 0)
	if rst == 0 {
		return os.NewSyscallError("InternetSetOptionW(PROXY_SETTINGS_CHANGED)", err)
	}
	rst, _, err = syscall.SyscallN(procInternetSetOptionW.Addr(), 0, internetOptionRefresh, 0, 0)
	if rst == 0 {
		return os.NewSyscallError("InternetSetOptionW(REFRESH)", err)
	}
	return nil
}

func InternetQuery() (*InternetStatus, error) {
	options := []internetPerConnOptionW{
		// On Windows 7 or above (IE8+), query with INTERNET_PER_CONN_FLAGS_UI is recommended.
		// See https://learn.microsoft.com/zh-cn/windows/win32/api/wininet/ns-wininet-internet_per_conn_optiona
		{dwOption: internetPerConnFlagsUi},
		{dwOption: internetPerConnProxyServer},
		{dwOption: internetPerConnProxyBypass},
		{dwOption: internetPerConnAutoconfigUrl},
	}
	list := internetPerConnOptionListW{
		dwSize:        sizeInternetPerConnOptionListW,
		dwOptionCount: uint32(len(options)),
		pOptions:      uintptr(unsafe.Pointer(&options[0])),
	}
	rst, _, err := syscall.SyscallN(procInternetQueryOptionW.Addr(), 0, internetOptionPerConnectionOption, uintptr(unsafe.Pointer(&list)), uintptr(unsafe.Pointer(&list.dwSize)))
	if rst == 0 {
		// Set option to INTERNET_PER_CONN_FLAGS and try again to compatible with older versions of Windows.
		options[0].dwOption = internetPerConnFlags
		rst, _, err = syscall.SyscallN(procInternetQueryOptionW.Addr(), 0, internetOptionPerConnectionOption, uintptr(unsafe.Pointer(&list)), uintptr(unsafe.Pointer(&list.dwSize)))
		if rst == 0 {
			return nil, os.NewSyscallError("InternetQueryOptionW", err)
		}
	}
	return &InternetStatus{
		ProxyType:   ProxyType(options[0].value),
		ProxyServer: utf16PtrToString(options[1].value),
		ProxyBypass: utf16PtrToString(options[2].value),
		ConfigUrl:   utf16PtrToString(options[3].value),
	}, nil
}

func internetStatusToInternetPerConnOptionListW(status InternetStatus) (*internetPerConnOptionListW, error) {
	if status.ProxyType > 0x0F || status.ProxyType < 1 {
		return nil, errors.New("invalid proxy type")
	}
	options := make([]internetPerConnOptionW, 0, 4)
	options = append(options, internetPerConnOptionW{
		dwOption: internetPerConnFlags, value: uintptr(status.ProxyType),
	})
	if status.ProxyServer != "" {
		value, err := syscall.UTF16PtrFromString(status.ProxyServer)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy server value: %w", err)
		}
		options = append(options, internetPerConnOptionW{
			dwOption: internetPerConnProxyServer,
			value:    uintptr(unsafe.Pointer(value)),
		})
	}
	if status.ProxyBypass != "" {
		value, err := syscall.UTF16PtrFromString(status.ProxyBypass)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy bypass value: %w", err)
		}
		options = append(options, internetPerConnOptionW{
			dwOption: internetPerConnProxyBypass,
			value:    uintptr(unsafe.Pointer(value)),
		})
	}
	if status.ConfigUrl != "" {
		value, err := syscall.UTF16PtrFromString(status.ConfigUrl)
		if err != nil {
			return nil, fmt.Errorf("invalid autoconfig url value: %w", err)
		}
		options = append(options, internetPerConnOptionW{
			dwOption: internetPerConnAutoconfigUrl,
			value:    uintptr(unsafe.Pointer(value)),
		})
	}
	return &internetPerConnOptionListW{
		dwSize:        sizeInternetPerConnOptionListW,
		dwOptionCount: uint32(len(options)),
		pOptions:      uintptr(unsafe.Pointer(&options[0])),
	}, nil
}

// utf16PtrToString takes a pointer to a UTF-16 sequence and returns the corresponding UTF-8 encoded string.
func utf16PtrToString(ptr uintptr) string {
	p := (*uint16)(unsafe.Pointer(ptr))
	if p == nil || *p == 0 {
		return ""
	}
	n := 0
	for size := unsafe.Sizeof(uint16(0)); *(*uint16)(unsafe.Pointer(ptr)) != 0; n++ {
		ptr += size
	}
	return syscall.UTF16ToString(unsafe.Slice(p, n))
}
