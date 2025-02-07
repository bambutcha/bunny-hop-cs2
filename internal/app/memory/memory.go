package memory

import (
    "fmt"
    "unsafe"
    "golang.org/x/sys/windows"
)

type MemoryReader struct {
    processHandle windows.Handle
}

func NewMemoryReader(processID uint32) (*MemoryReader, error) {
    handle, err := windows.OpenProcess(
        windows.PROCESS_QUERY_INFORMATION|
        windows.PROCESS_VM_READ|
        windows.PROCESS_VM_WRITE|
        windows.PROCESS_VM_OPERATION,
        false,
        processID,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to open process: %v", err)
    }
    return &MemoryReader{processHandle: handle}, nil
}

func (m *MemoryReader) WriteInt(address uintptr, value int32) error {
    var written uintptr
    data := (*byte)(unsafe.Pointer(&value))
    err := windows.WriteProcessMemory(
        m.processHandle,
        address,
        data,
        unsafe.Sizeof(value),
        &written,
    )
    if err != nil {
        return fmt.Errorf("failed to write memory: %v", err)
    }
    if written == 0 {
        return fmt.Errorf("no bytes written to memory")
    }
    return nil
}

func (m *MemoryReader) Close() {
	if m.processHandle != 0 {
		windows.CloseHandle(m.processHandle)
	}
}

func (m *MemoryReader) ReadMemory(address uint32, buffer []byte) error {
	var read uintptr
	err := windows.ReadProcessMemory(
		m.processHandle,
		uintptr(address),
		&buffer[0],
		uintptr(len(buffer)),
		&read,
	)
	if err != nil {
		return fmt.Errorf("Failed to read memory: %v", err)
	}
	return nil
}

func (m *MemoryReader) WriteMemory(address uint32, buffer []byte) error {
	var written uintptr
	err := windows.WriteProcessMemory(
		m.processHandle,
		uintptr(address),
		&buffer[0],
		uintptr(len(buffer)),
		&written,
	)
	if err != nil {
		return fmt.Errorf("Failed to write memory: %v", err)
	}
	return nil
}