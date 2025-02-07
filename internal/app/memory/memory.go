package memory

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
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
	processHandle, err := windows.OpenProcess(windows.PROCESS_ALL_ACCESS, false, M.ProcessID)
	if err != nil {
		return nil, fmt.Errorf("Failed to open process: %v", err)
	}
	defer windows.CloseHandle(processHandle)

	buffer := make([]byte, size)
	var bytesRead uint32
	if err := windows.ReadProcessMemory(processHandle, address, uintptr(unsafe.Pointer(&buffer[0])), size, &bytesRead); err != nil {
		return nil, fmt.Errorf("Failed to read memory: %v", err)
	}

	return buffer, nil
}

func (M *MemoryReader) WriteMemory(address uint32, data []byte) error {
	processHandle, err := windows.OpenProcess(windows.PROCESS_ALL_ACCESS, false, M.ProcessID)
	if err != nil {
		return fmt.Errorf("Failed to open process: %v", err)
	}
	defer windows.CloseHandle(processHandle)
	
	var bytesWritten uint32
	if err := windows.WriteProcessMemory(processHandle, address, uintptr(unsafe.Pointer(&data[0])), uint32(len(data)), &bytesWritten); err != nil {
		return fmt.Errorf("Failed to write memory: %v", err)
	}

	return nil
}