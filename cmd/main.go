package main

import (
	"fmt"

	"github.com/bambutcha/cs2-bhop/internal/app/bhop"
	"github.com/bambutcha/cs2-bhop/internal/app/logger"
)

func main() {
	logger := logger.NewLogger()
	bhop := bhop.NewBhop(logger)
	bhop.Start()

	fmt.Println("\nНажмите Enter для выхода...")
	fmt.Scanln()
}
