package output

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type OutputLevel int

const Debug OutputLevel = 0
const Info OutputLevel = 1
const Error OutputLevel = 2

func Println(targetLevel OutputLevel, line string, args ...interface{}) {
	currentLevel := outputLevelFromEnv()
	if targetLevel < OutputLevel(currentLevel) {
		return
	}
	if !strings.HasSuffix(line, "\n") {
		line = line + "\n"
	}
	t := time.Now()
	line = "%s: " + line
	fmt.Printf(line, t.Format(time.RFC3339), args)
}

func outputLevelFromEnv() (outputLevel OutputLevel) {
	outputLevel = Info
	envStr := os.Getenv("OUTPUT_LEVEL")
	if len(envStr) > 0 {
		newInt, err := strconv.Atoi(envStr)
		if err == nil {
			outputLevel = OutputLevel(newInt)
		} else {
			fmt.Println(err)
		}
	}

	return outputLevel
}
