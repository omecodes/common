package futils

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

func winHideFile(filename string) (string, error) {
	filenameW, err := syscall.UTF16PtrFromString(filename)
	if err != nil {
		return "", err
	}
	err = syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_HIDDEN)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func winDiskStatus(disk string) (DiskStatus, error) {
	kernel32, err := syscall.LoadLibrary("Kernel32.dll")
	if err != nil {
		return DiskStatus{}, err
	}
	defer func() {
		_ = syscall.FreeLibrary(kernel32)
	}()

	GetDiskFreeSpaceEx, err := syscall.GetProcAddress(syscall.Handle(kernel32), "GetDiskFreeSpaceExW")
	if err != nil {
		return DiskStatus{}, err
	}

	diskNamePtr, err := syscall.UTF16PtrFromString(disk)
	if err != nil {
		return DiskStatus{}, err
	}

	lpFreeBytesAvailable := int64(0)
	lpTotalNumberOfBytes := int64(0)
	lpTotalNumberOfFreeBytes := int64(0)
	_, _, e := syscall.Syscall6(uintptr(GetDiskFreeSpaceEx), 4,
		uintptr(unsafe.Pointer(diskNamePtr)),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)), 0, 0)
	if e != 0 {
		return DiskStatus{}, errors.New("failed to load disk status")
	}

	all := uint64(lpTotalNumberOfBytes)
	free := uint64(lpTotalNumberOfFreeBytes)

	ds := DiskStatus{
		All:  all,
		Free: free,
		Used: all - free,
	}

	/*log.Printf("Available  %dmb", lpFreeBytesAvailable/1024/1024.0)
	log.Printf("Total      %dmb", lpTotalNumberOfBytes/1024/1024.0)
	log.Printf("Free       %dmb", lpTotalNumberOfFreeBytes/1024/1024.0)*/
	return ds, nil
}

func winDriveList() (drives []string) {

	checkDriveWithinTimeOut := func(d int32) bool {
		driveChan := make(chan string, 1)
		go func() {
			_, err := os.Open(string(d) + ":\\")
			if err == nil {
				driveChan <- fmt.Sprintf("%c", d)
			}
		}()

		select {
		case <-driveChan:
			return true
		case <-time.After(time.Millisecond * 10):
			return false
		}
	}

	for _, d := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		if checkDriveWithinTimeOut(d) {
			drives = append(drives, fmt.Sprintf("%c", d))
		}
	}
	return
}
