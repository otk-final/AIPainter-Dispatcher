package lb

import (
	"fmt"
	"github.com/samber/lo"
	"testing"
)

func TestNew(t *testing.T) {
	m := New(20, nil)
	m.Add("http://localhost:8001", "http://localhost:8002", "http://localhost:8003", "http://localhost:8004")

	var targets []string
	for i := 0; i < 10000; i++ {
		targets = append(targets, m.Get(fmt.Sprintf("user_%d", i)))
	}

	t.Log(lo.CountValues(targets))
}
