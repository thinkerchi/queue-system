package utils

import (
	"testing"
)

func TestGetGuid(t *testing.T) {
	uid := GetGuid()
	t.Log(uid)
	t.Log(len(uid))
}
