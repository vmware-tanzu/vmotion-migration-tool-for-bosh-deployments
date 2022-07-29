package duration_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/duration"
)

func TestHumanReadable(t *testing.T) {
	durations := []time.Duration{
		52 * time.Second,
		28*time.Hour + 34*time.Minute + 9*time.Second + 34*time.Millisecond,
	}
	expected := []string{
		"52s",
		"28h 34m 9s",
	}
	for i, d := range durations {
		require.Equal(t, expected[i], duration.HumanReadable(d))
	}
}
