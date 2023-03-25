package queues

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestTagReader_GetGlobalReader(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns a global reader even if no tags have been  defined yet", func(t *testing.T) {
		reader := NewReaderTree()
		g.Expect(reader).ToNot(BeNil())
		g.Expect(reader.GetGlobalReader()).ToNot(BeNil())
	})
}
