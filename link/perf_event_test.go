package link

import (
	"testing"

	"github.com/mscastanho/ebpf/internal/testutils"
)

func TestHaveBPFLinkPerfEvent(t *testing.T) {
	testutils.CheckFeatureTest(t, haveBPFLinkPerfEvent)
}
