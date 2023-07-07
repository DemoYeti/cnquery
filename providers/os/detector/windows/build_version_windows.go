//go:build windows
// +build windows

package windows

import (
	"runtime"

	"golang.org/x/sys/windows/registry"
)

func GetWindowsOSBuild(conn connection.Connection) (*WindowsCurrentVersion, error) {
	// if we are running locally on windows, we want to avoid using powershell to be faster
	if ok && runtime.GOOS == "windows" {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
		if err != nil {
			return nil, err
		}

		currentBuild, _, err := k.GetStringValue("CurrentBuild")
		if err != nil && err != registry.ErrNotExist {
			return nil, err
		}

		ubr, _, err := k.GetIntegerValue("UBR")
		if err != nil && err != registry.ErrNotExist {
			return nil, err
		}

		edition, _, err := k.GetStringValue("EditionID")
		if err != nil && err != registry.ErrNotExist {
			return nil, err
		}
		defer k.Close()

		return &WindowsCurrentVersion{
			CurrentBuild: currentBuild,
			EditionID:    edition,
			UBR:          int(ubr),
		}, nil
	}

	// for all non-local checks use powershell
	return powershellGetWindowsOSBuild(p)
}
