package windowsMutex

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func WindowsMutex(appName string) (mutex windows.Handle, err error) {
	muName := windows.StringToUTF16Ptr(appName)
	mutex, err = windows.OpenMutex(windows.MUTEX_ALL_ACCESS, false, muName)

	if err == nil {
		windows.CloseHandle(mutex)
		err = fmt.Errorf(appName + " is already running")
		return
	}

	mutex, err = windows.CreateMutex(nil, false, muName)
	if err != nil {
		err = fmt.Errorf(appName+" is already running err : %w", err)
		return
	}

	event, err := windows.WaitForSingleObject(mutex, windows.INFINITE)
	if err != nil {
		err = fmt.Errorf(appName+" wait for mutex error: %w", err)
		return
	}

	switch event {
	case windows.WAIT_OBJECT_0, windows.WAIT_ABANDONED:
	default:
		err = fmt.Errorf(appName+" wait for mutex event id error: %v", event)
		return

	}

	return
}
