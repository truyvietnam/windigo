/**
 * Part of Wingows - Win32 API layer for Go
 * https://github.com/rodrigocfd/wingows
 * This library is released under the MIT license.
 */

package win

import (
	"fmt"
	"syscall"
	"unsafe"
	"wingows/co"
	"wingows/win/proc"
)

// https://docs.microsoft.com/en-us/windows/win32/winprog/windows-data-types#haccel
type HACCEL HANDLE

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-copyacceleratortablew
func (hAccel HACCEL) CopyAcceleratorTable() []ACCEL {
	szRet, _, _ := syscall.Syscall(proc.CopyAcceleratorTable.Addr(), 3,
		uintptr(hAccel), 0, 0)
	if szRet == 0 {
		return []ACCEL{}
	}
	accelList := make([]ACCEL, uint32(szRet))
	syscall.Syscall(proc.CopyAcceleratorTable.Addr(), 3,
		uintptr(hAccel), uintptr(unsafe.Pointer(&accelList[0])), szRet)
	return accelList
}

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-createacceleratortablew
func CreateAcceleratorTable(accelList []ACCEL) HACCEL {
	ret, _, lerr := syscall.Syscall(proc.CreateAcceleratorTable.Addr(), 2,
		uintptr(unsafe.Pointer(&accelList[0])), uintptr(len(accelList)),
		0)
	if ret == 0 {
		panic(fmt.Sprintf("CreateAcceleratorTable failed. %s",
			co.ERROR(lerr).Error()))
	}
	return HACCEL(ret)
}

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-destroyacceleratortable
func (hAccel HACCEL) DestroyAcceleratorTable() bool {
	ret, _, _ := syscall.Syscall(proc.DestroyAcceleratorTable.Addr(), 1,
		uintptr(hAccel), 0, 0)
	return ret != 0
}
