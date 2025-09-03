package wzlib

import (
	"testing"

	"github.com/fastgox/utils/logger"
)

func TestLogger(t *testing.T) {
	logger.InitDefault()
	logger.Info("Hello, World!")
}
