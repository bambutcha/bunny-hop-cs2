package config

type Config struct {
    // Основные настройки
    ProcessName     string
    ModuleName     string
    OffsetsURL     string
    Version        string

    // Настройки бхопа
    JumpDelay      int // в миллисекундах
    PressValue     int32
    ReleaseValue   int32
    
    // Клавиши
    JumpKey        int
    ToggleKey      int // Клавиша для включения/выключения
}

func NewDefaultConfig() *Config {
    return &Config{
        ProcessName:   "cs2.exe",
        ModuleName:   "client.dll",
        OffsetsURL:   "https://raw.githubusercontent.com/a2x/cs2-dumper/main/output/buttons.hpp",
        Version:      "1.1.0",
        JumpDelay:    1,
        PressValue:   65537,
        ReleaseValue: 256,
        JumpKey:      0x20, // Space
        ToggleKey:    0x2D, // Insert
    }
}