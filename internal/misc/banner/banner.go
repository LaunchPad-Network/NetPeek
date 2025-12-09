package banner

import (
	"fmt"

	"github.com/LaunchPad-Network/NetPeek/internal/version"
)

func PrintBanner(extraName string) {
	if extraName == "" {
		printOutBanner("NetPeek")
	} else {
		printOutBanner("NetPeek " + extraName)
	}
	fmt.Printf("\nBuild Time: %s; Commit Hash: %s\n\n", version.BuildTime(), version.CommitHash())
}
