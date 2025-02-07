package bhop

import (
    "fmt"
    "time"
    "strings"
    "io/ioutil"
    "net/http"
    "strconv"
    "unsafe"

    "github.com/bambutcha/cs2-bhop/internal/app/logger"
    "github.com/bambutcha/cs2-bhop/internal/app/memory"
    "golang.org/x/sys/windows"
)

const (
    VK_SPACE = 0x20
)

var (
    user32 = windows.NewLazySystemDLL("user32.dll")
    getAsyncKeyState = user32.NewProc("GetAsyncKeyState")
)

type Bhop struct {
    Version          string
    ProcessID        uint32
    ClientBase       uintptr
    ForceJumpAddress uintptr
    Logger           *logger.Logger
    MemoryReader     *memory.MemoryReader
}

func NewBhop(logger *logger.Logger) *Bhop {
    return &Bhop{
        Version: "1.0.0",
        Logger:  logger,
    }
}

func (b *Bhop) FindProcessID(processName string) (uint32, error) {
    snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
    if err != nil {
        return 0, err
    }
    defer windows.CloseHandle(snapshot)

    var entry windows.ProcessEntry32
    entry.Size = uint32(unsafe.Sizeof(entry))

    err = windows.Process32First(snapshot, &entry)
    if err != nil {
        return 0, err
    }

    for {
        if strings.EqualFold(windows.UTF16ToString(entry.ExeFile[:]), processName) {
            return entry.ProcessID, nil
        }
        err = windows.Process32Next(snapshot, &entry)
        if err != nil {
            break
        }
    }
    return 0, fmt.Errorf("process not found")
}

func (b *Bhop) GetModuleBaseAddress(moduleName string) (uintptr, error) {
    snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPMODULE|windows.TH32CS_SNAPMODULE32, b.ProcessID)
    if err != nil {
        return 0, err
    }
    defer windows.CloseHandle(snapshot)

    var me windows.ModuleEntry32
    me.Size = uint32(unsafe.Sizeof(me))

    err = windows.Module32First(snapshot, &me)
    if err != nil {
        return 0, err
    }

    for {
        if strings.EqualFold(windows.UTF16ToString(me.Module[:]), moduleName) {
            return uintptr(me.ModBaseAddr), nil
        }
        err = windows.Module32Next(snapshot, &me)
        if err != nil {
            break
        }
    }
    return 0, fmt.Errorf("module not found")
}

func (b *Bhop) FetchOffsets() (uintptr, error) {
    response, err := http.Get("https://raw.githubusercontent.com/a2x/cs2-dumper/main/output/buttons.hpp")
    if err != nil {
        return 0, err
    }
    defer response.Body.Close()

    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return 0, err
    }

    lines := strings.Split(string(body), "\n")
    for _, line := range lines {
        if strings.Contains(line, "jump") {
            parts := strings.Split(line, "=")
            if len(parts) < 2 {
                continue
            }
            offsetStr := strings.TrimSpace(strings.TrimSuffix(parts[1], ";"))
            offset, err := strconv.ParseUint(offsetStr, 0, 64)
            if err != nil {
                continue
            }
            return uintptr(offset), nil
        }
    }
    return 0, fmt.Errorf("offset not found")
}

func (b *Bhop) Initialize() error {
    b.Logger.Info("Searching for cs2.exe process...")
    pid, err := b.FindProcessID("cs2.exe")
    if err != nil {
        return fmt.Errorf("failed to find cs2.exe process: %v", err)
    }
    b.ProcessID = pid

    memReader, err := memory.NewMemoryReader(pid)
    if err != nil {
        return fmt.Errorf("failed to create memory reader: %v", err)
    }
    b.MemoryReader = memReader

    b.Logger.Info("Getting client.dll base address...")
    clientBase, err := b.GetModuleBaseAddress("client.dll")
    if err != nil {
        return fmt.Errorf("failed to get client.dll base address: %v", err)
    }
    b.ClientBase = clientBase

    b.Logger.Info("Fetching offsets...")
    offset, err := b.FetchOffsets()
    if err != nil {
        return fmt.Errorf("failed to fetch offsets: %v", err)
    }

    b.ForceJumpAddress = b.ClientBase + offset
    b.Logger.Info(fmt.Sprintf("Client base: 0x%X, Offset: 0x%X, Final address: 0x%X",
        b.ClientBase, offset, b.ForceJumpAddress))

    return nil
}

func (b *Bhop) Start() {
    if err := b.Initialize(); err != nil {
        b.Logger.Error(err.Error())
        return
    }

    b.Logger.Info("Bunnyhop started. Hold SPACE to hopping.")
    
    jump := false
    for {
        ret, _, _ := getAsyncKeyState.Call(uintptr(VK_SPACE))
        if ret&0x8000 != 0 {
            if !jump {
                time.Sleep(10 * time.Millisecond)
                if err := b.MemoryReader.WriteInt(b.ForceJumpAddress, 6); err != nil {
                    b.Logger.Error(fmt.Sprintf("Failed to write memory: %v", err))
                    continue
                }
                jump = true
            }
        } else {
            if jump {
                time.Sleep(10 * time.Millisecond)
                if err := b.MemoryReader.WriteInt(b.ForceJumpAddress, 4); err != nil {
                    b.Logger.Error(fmt.Sprintf("Failed to write memory: %v", err))
                    continue
                }
                jump = false
            }
        }
        time.Sleep(1 * time.Millisecond)
    }
}