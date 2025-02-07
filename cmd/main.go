package main

import (
	"github.com/bambutcha/cs2-bhop/internal/app/bhop"
	"github.com/bambutcha/cs2-bhop/internal/app/logger"
)

func main() {
	logger := logger.NewLogger()
	bhop := bhop.NewBhop(logger)
	bhop.Start()
}