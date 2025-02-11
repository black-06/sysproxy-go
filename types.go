package sysproxy_go

import "unsafe"

type (
	ProxyType      uintptr
	InternetStatus struct {
		ProxyType   ProxyType
		ProxyServer string
		ProxyBypass string
		ConfigUrl   string
	}
)

const (
	ProxyTypeDirect       ProxyType = 1
	ProxyTypeProxy        ProxyType = 2
	ProxyTypeAutoProxyUrl ProxyType = 4
	ProxyTypeAutoDetect   ProxyType = 8
)

type (
	internetPerConn uint32

	// internetPerConnOptionW is INTERNET_PER_CONN_OPTIONW,
	// see https://learn.microsoft.com/zh-cn/windows/win32/api/wininet/ns-wininet-internet_per_conn_optionw
	internetPerConnOptionW struct {
		dwOption internetPerConn
		value    uintptr // dwValue(ProxyType)/pszValue(LPWSTR)/ftValue(syscall.Filetime) union value.
	}

	// internetPerConnOptionListW is INTERNET_PER_CONN_OPTION_LISTW,
	// see https://learn.microsoft.com/zh-cn/windows/win32/api/wininet/ns-wininet-internet_per_conn_option_listw
	internetPerConnOptionListW struct {
		dwSize        uintptr
		pszConnection uintptr
		dwOptionCount uint32
		dwOptionError uint32
		pOptions      uintptr
	}
)

const (
	internetPerConnFlags                     internetPerConn = 1
	internetPerConnProxyServer               internetPerConn = 2
	internetPerConnProxyBypass               internetPerConn = 3
	internetPerConnAutoconfigUrl             internetPerConn = 4
	internetPerConnAutodiscoveryFlags        internetPerConn = 5
	internetPerConnAutoconfigSecondaryUrl    internetPerConn = 6
	internetPerConnAutoconfigReloadDelayMins internetPerConn = 7
	internetPerConnAutoconfigLastDetectTime  internetPerConn = 8
	internetPerConnAutoconfigLastDetectUrl   internetPerConn = 9
	internetPerConnFlagsUi                   internetPerConn = 10
)

// https://learn.microsoft.com/zh-cn/windows/win32/wininet/option-flags
const (
	internetOptionPerConnectionOption  = 75
	internetOptionProxySettingsChanged = 95
	internetOptionRefresh              = 37
)

// rasEntryNameW https://learn.microsoft.com/zh-cn/previous-versions/windows/desktop/legacy/aa377267(v=vs.85)
type rasEntryNameW struct {
	dwSize          uintptr
	szEntryName     [257]uint16 // RAS_MaxEntryName is 256
	dwFlags         uint32
	szPhonebookPath [261]uint16 // MAX_PATH is 260
}

const (
	errorBufferTooSmall uintptr = 603
)

var (
	sizeRASEntryNameW              = unsafe.Sizeof(rasEntryNameW{})
	sizeInternetPerConnOptionListW = unsafe.Sizeof(internetPerConnOptionListW{})
)
