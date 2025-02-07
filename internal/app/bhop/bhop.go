package bhop

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/bambutcha/cs2-bhop/internal/app/logger"
	"github.com/bambutcha/cs2-bhop/internal/app/memory"
	"golang.org/x/sys/windows"
)

type Bhop struct {
	Version 		 string
	ProcessID 		 uint32
	ClientBase 		 uintptr
	ForceJumpAddress uintptr
	Logger 			 *logger.Logger
	MemoryReader 	 *memory.MemoryReader
}

func NewBhop(logger *logger.Logger) *Bhop {
	return &Bhop {
		Version: "1.0.0",
		Logger: logger,
	}
}

func (b *Bhop) Initialize() error {
	b.Logger.Info("Searching for cs2.exe process...")
	pid, err := b.FindProcessID("cs2.exe")
	if err != nil {
		return fmt.Errorf("Failed to find cs2.exe process: %v", err)
	}
	b.ProcessID = pid

	b.MemoryReader = memory.NewMemoryReader(b.ProcessID)

	b.Logger.Info("Getting client.dll base address...")
	clientBase, err := b.GetModuleBaseAddress("client.dll")
	if err != nil {
		return fmt.Errorf("Failed to get client.dll base address: %v", err)
	}
	b.ClientBase = clientBase

	b.Logger.Info("Fetching offsets...")
	offset, err := b.FetchOffsets()
	if err != nil {
		return fmt.Errorf("Failed to fetch offsets: %v", err)
	}
	b.ForceJumpAddress = b.ClientBase + offset

	b.Logger.Info("Bhop initialized successfully!")
	
	return nil
}

func (b *Bhop) FindProcessID(processName string) (uint32, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if err := windows.Process32First(snapshot, &entry); err != nil {
		return 0, err
	}

	for {
		if strings.EqualFold(windows.UTF16ToString(entry.ExeFile[:]), processName) {
			return entry.ProcessID, nil
		}
		if err := windows.Process32Next(snapshot, &entry); err != nil {
			break
		}
	}

	return 0, fmt.Errorf("Process not found")
}

func (b *Bhop) GetModuleBaseAddress(moduleName string) (uintptr, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPMODULE, b.ProcessID)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.MODULEENTRY32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if err := windows.Module32First(snapshot, &entry); err != nil {
		return 0, err
	}

	for {
		if strings.EqualFold(windows.UTF16ToString(entry.SzModule[:]), moduleName) {
			return uintptr(entry.ModBaseAddr), nil
		}
		if err := windows.Module32Next(snapshot, &entry); err != nil {
			break
		}
	}

	return 0, fmt.Errorf("Module not found")	
}

func (b *Bhop) FetchOffsets() (uintptr, error) {
	response, err := http.Get("https://raw.githubusercontent.com/a2x/cs2-dumper/main/output/buttons.hpp")
	if err != nil {
		return 0, fmt.Errorf("Failed to fetch offsets: %v", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, fmt.Errorf("Failed to read response body: %v", err)
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
				return 0, fmt.Errorf("Failed to parse offset: %v", err)
			}

			return uintptr(offset), nil
		}
	}

	return 0, fmt.Errorf("Offsets not found")
}

func (b *Bhop) Start() {
	if err := b.Initialize(); err != nil {
		b.Logger.Error(err.Error())
		return
	}

	b.Logger.Info("Bunnyhop started. Hold SPACE to hopping.")
	for {
		if windows.GetAsyncKeyState(windows.VK_SPACE) & 0x8000 != 0 {
			b.Jump()
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (b *Bhop) Jump() {
	value := int32(65537)
	if err := b.MemoryReader.WriteMemory(uint32(b.ForceJumpAddress), (*[4]byte)(unsafe.Pointer(&value))[:]); err != nil {
		b.Logger.Error(fmt.Sprintf("Failed to write memory: %v", err))
		return
	}

	time.Sleep(10 * time.Millisecond)

	value = 256
	if err := b.MemoryReader.WriteMemory(uint32(b.ForceJumpAddress), (*[4]byte)(unsafe.Pointer(&value))[:]); err != nil {
		b.Logger.Error(fmt.Sprintf("Failed to write memory: %v", err))
		return
	}
}