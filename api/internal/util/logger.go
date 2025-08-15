package utils

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func InitilizeLogger() {
	Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
