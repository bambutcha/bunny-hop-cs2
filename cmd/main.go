package main

func main() {
	logger := logger.NewLogger()
	bhop := bhop.NewBhop(logger)
	bhop.Start()
}