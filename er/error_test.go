package er

import (
	"fmt"
	"regexp"
	"testing"
)

func TestMatch(t *testing.T) {
	match, _ := regexp.MatchString("release: .* not found", "release: \"mongodb\" not found")
	t.Log(match)
}

func TestError(t *testing.T) {
	err := fmt.Errorf("new error")
	fmt.Println(ErrorF(err, "test error!"))
}
