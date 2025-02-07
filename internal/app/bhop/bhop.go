package bhop

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/bambutcha/cs2-bhop/internal/app/config"
	"github.com/bambutcha/cs2-bhop/internal/app/logger"
	"github.com/bambutcha/cs2-bhop/internal/app/memory"
	"golang.org/x/sys/windows"
)

var (
	user32           = windows.NewLazySystemDLL("user32.dll")
	getAsyncKeyState = user32.NewProc("GetAsyncKeyState")
)

type Bhop struct {
	config           *config.Config
	logger           *logger.Logger
	memoryReader     *memory.MemoryReader
	processID        uint32
	clientBase       uintptr
	forceJumpAddress uintptr
	isEnabled        atomic.Bool
	isRunning        atomic.Bool
}

func NewBhop(logger *logger.Logger) *Bhop {
	return &Bhop{
		config: config.NewDefaultConfig(),
		logger: logger,
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
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPMODULE|windows.TH32CS_SNAPMODULE32, b.processID)
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
	response, err := http.Get(b.config.OffsetsURL)
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
	b.logger.Info("Searching for cs2.exe process...")
	pid, err := b.FindProcessID(b.config.ProcessName)
	if err != nil {
		return fmt.Errorf("failed to find cs2.exe process: %v", err)
	}
	b.processID = pid

	memReader, err := memory.NewMemoryReader(pid)
	if err != nil {
		return fmt.Errorf("failed to create memory reader: %v", err)
	}
	b.memoryReader = memReader

	b.logger.Info("Getting client.dll base address...")
	clientBase, err := b.GetModuleBaseAddress(b.config.ModuleName)
	if err != nil {
		return fmt.Errorf("failed to get client.dll base address: %v", err)
	}
	b.clientBase = clientBase

	b.logger.Info("Fetching offsets...")
	offset, err := b.FetchOffsets()
	if err != nil {
		return fmt.Errorf("failed to fetch offsets: %v", err)
	}

	b.forceJumpAddress = b.clientBase + offset
	b.logger.Info(fmt.Sprintf("Client base: 0x%X, Offset: 0x%X, Final address: 0x%X",
		b.clientBase, offset, b.forceJumpAddress))

	return nil
}

func (b *Bhop) handleToggle() {
	ret, _, _ := getAsyncKeyState.Call(uintptr(b.config.ToggleKey))
	if ret&0x1 != 0 { // Клавиша только что нажата
		if b.isEnabled.Load() {
			b.isEnabled.Store(false)
			b.logger.Info("Bhop disabled")
		} else {
			b.isEnabled.Store(true)
			b.logger.Info("Bhop enabled")
		}
	}
}

func (b *Bhop) Start() {
	if err := b.Initialize(); err != nil {
		b.logger.Error(err.Error())
		return
	}

	b.logger.Info(fmt.Sprintf("Bhop %s started. Press INSERT to toggle, SPACE to hop.", b.config.Version))
	b.isRunning.Store(true)
	b.isEnabled.Store(true)

	jump := false
	lastJumpTime := time.Now()

	for b.isRunning.Load() {
		b.handleToggle()

		if !b.isEnabled.Load() {
			time.Sleep(time.Duration(b.config.JumpDelay) * time.Millisecond)
			continue
		}

		ret, _, _ := getAsyncKeyState.Call(uintptr(b.config.JumpKey))
		if ret&0x8000 != 0 { // Space is pressed
			if !jump && time.Since(lastJumpTime) > 10*time.Millisecond {
				if err := b.memoryReader.WriteInt(b.forceJumpAddress, b.config.PressValue); err != nil {
					b.logger.Error(fmt.Sprintf("Failed to write memory: %v", err))
					continue
				}
				jump = true
				lastJumpTime = time.Now()
			} else if jump && time.Since(lastJumpTime) > 10*time.Millisecond {
				if err := b.memoryReader.WriteInt(b.forceJumpAddress, b.config.ReleaseValue); err != nil {
					b.logger.Error(fmt.Sprintf("Failed to write memory: %v", err))
					continue
				}
				jump = false
				lastJumpTime = time.Now()
			}
		}
		time.Sleep(time.Duration(b.config.JumpDelay) * time.Millisecond)
	}
}

func (b *Bhop) Stop() {
	b.isRunning.Store(false)
	if b.memoryReader != nil {
		b.memoryReader.Close()
	}
}
