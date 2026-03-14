package debug

import (
	"log"

	"github.com/spf13/viper"
)

// DebugPrint writes debug output only when --debug is enabled.
func DebugPrint(format string, args ...interface{}) {
	if viper.GetBool("debug") {
		log.Printf("[DEBUG] "+format+"\n", args...)
	}
}
