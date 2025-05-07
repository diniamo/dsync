package walk

import (
	"fmt"
	"strings"
)

// Yes, this is hard-coded.
// No, I will not spend an unreasonable amount of time implementing a non-caveman way.
const maxLeftLength = len("local -> remote (mkdir)")
const extraPadding = 3
const effectivePadding = maxLeftLength + extraPadding

func paddedPrint(left, right string) {
	padding := effectivePadding - len(left)
	fmt.Printf("%s%s%s\n", left, strings.Repeat(" ", padding), right)
}
