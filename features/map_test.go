package features

import (
	"errors"
	"math"
	"testing"

	"github.com/mscastanho/ebpf"
	"github.com/mscastanho/ebpf/internal"
	"github.com/mscastanho/ebpf/internal/testutils"
)

func TestHaveMapType(t *testing.T) {
	testutils.CheckFeatureMatrix(t, haveMapTypeMatrix)
}

func TestHaveMapFlag(t *testing.T) {
	testutils.CheckFeatureMatrix(t, haveMapFlagsMatrix)
}

func TestHaveMapTypeInvalid(t *testing.T) {
	if err := HaveMapType(ebpf.MapType(math.MaxUint32)); err == nil {
		t.Fatal("Expected an error")
	} else if errors.Is(err, internal.ErrNotSupported) {
		t.Fatal("Got ErrNotSupported:", err)
	}
}
