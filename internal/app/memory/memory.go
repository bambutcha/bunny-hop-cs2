package memory

import (
    "fmt"
    "sync"
    "unsafe"
    "golang.org/x/sys/windows"
)

type MemoryReader struct {
    processHandle windows.Handle
    mutex         sync.Mutex
    writeCache    map[uintptr]int32
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
    return &MemoryReader{
        processHandle: handle,
        writeCache:   make(map[uintptr]int32),
    }, nil
}

func (m *MemoryReader) WriteInt(address uintptr, value int32) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    // Проверяем кэш, чтобы избежать лишних записей
    if cachedValue, exists := m.writeCache[address]; exists && cachedValue == value {
        return nil
    }

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

    m.writeCache[address] = value
    return nil
}

func (m *MemoryReader) ReadInt(address uintptr) (int32, error) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    var value int32
    var read uintptr
    err := windows.ReadProcessMemory(
        m.processHandle,
        address,
        (*byte)(unsafe.Pointer(&value)),
        unsafe.Sizeof(value),
        &read,
    )
    if err != nil {
        return 0, fmt.Errorf("failed to read memory: %v", err)
    }
    if read == 0 {
        return 0, fmt.Errorf("no bytes read from memory")
    }
    return value, nil
}

func (m *MemoryReader) Close() {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    if m.processHandle != 0 {
        windows.CloseHandle(m.processHandle)
        m.processHandle = 0
    }
}