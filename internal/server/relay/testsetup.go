//go:build testsetup

package relay

import (
	"time"
)

func init() {
	refreshInterval = 1 * time.Second
}
