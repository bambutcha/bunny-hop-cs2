package memory

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	PROCESS_ALL_ACCESS = 0x1F0FFF
)

type MemoryReader struct {
	ProcessID uint32
}

func NewMemoryReader(processID uint32) *MemoryReader {
	return &MemoryReader{
		ProcessID: processID,
	}
}

func (M *MemoryReader) ReadMemory(address uint32, size uint32) ([]byte, error) {
	processHandle, err := windows.OpenProcess(PROCESS_ALL_ACCESS, false, M.ProcessID)
	if err != nil {
		return nil, fmt.Errorf("Failed to open process: %v", err)
	}
	defer windows.CloseHandle(processHandle)

	buffer := make([]byte, size)
	var bytesRead uintptr
	err = windows.ReadProcessMemory(processHandle, 
		uintptr(address),
		(*byte)(unsafe.Pointer(&buffer[0])),
		uintptr(size),
		&bytesRead)
	if err != nil {
		return nil, fmt.Errorf("Failed to read memory: %v", err)
	}

	return buffer, nil
}

func (M *MemoryReader) WriteMemory(address uint32, data []byte) error {
	processHandle, err := windows.OpenProcess(PROCESS_ALL_ACCESS, false, M.ProcessID)
	if err != nil {
		return fmt.Errorf("Failed to open process: %v", err)
	}
	defer windows.CloseHandle(processHandle)
	
	var bytesWritten uintptr
	err = windows.WriteProcessMemory(processHandle,
		uintptr(address),
		(*byte)(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		&bytesWritten)
	if err != nil {
		return fmt.Errorf("Failed to write memory: %v", err)
	}

	return nil
}