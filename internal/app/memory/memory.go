package memory

import (
	"fmt"
	"unsafe"

	"github.com/0xrawsec/golang-win32/win32"
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
	processHandle, err := win32.OpenProcess(win32.PROCESS_ALL_ACCESS, false, M.ProcessID)
	if err != nil {
		return nil, fmt.Errorf("Failed to open process: %v", err)
	}
	defer win32.CloseHandle(processHandle)

	buffer := make([]byte, size)
	var bytesRead uint32
	if err := win32.ReadProcessMemory(processHandle, address, uintptr(unsafe.Pointer(&buffer[0])), size, &bytesRead); err != nil {
		return nil, fmt.Errorf("Failed to read memory: %v", err)
	}

	return buffer, nil
}
