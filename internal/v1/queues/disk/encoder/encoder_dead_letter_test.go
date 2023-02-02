package encoder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DanLavine/willow/internal/v1/queues/disk/encoder"
	. "github.com/onsi/gomega"
)

func TestEncoderDeadLetter_Write(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("writes data encoded to the dead letter file", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		encoderDeadLetter, deadLetterErr := encoder.NewEncoderDeadLetter(tmpDir, []string{"tag"})
		g.Expect(deadLetterErr).ToNot(HaveOccurred())

		// write first item
		diskLocation, deadLetterErr := encoderDeadLetter.Write([]byte(`hello world`)) //echo -n "hello world" | base64 -> aGVsbG8gd29ybGQ=
		g.Expect(deadLetterErr).ToNot(HaveOccurred())
		g.Expect(diskLocation.StartIndex).To(Equal(0))
		g.Expect(diskLocation.Size).To(Equal(16))

		// write second item
		diskLocation, deadLetterErr = encoderDeadLetter.Write([]byte(`this is item 2`)) // echo -n "this is item 2" | base64 -> dGhpcyBpcyBpdGVtIDI=
		g.Expect(deadLetterErr).ToNot(HaveOccurred())
		g.Expect(diskLocation.StartIndex).To(Equal(17))
		g.Expect(diskLocation.Size).To(Equal(20))

		// read the file to make sure its correct
		data, err := os.ReadFile(filepath.Join(tmpDir, encoder.EncodeString("tag"), "deadLetter.idx"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(data).To(Equal([]byte(`aGVsbG8gd29ybGQ=.dGhpcyBpcyBpdGVtIDI=.`)))
	})
}

func TestEncoderDeadLetter_Read(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("can read proper locations", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		encoderDeadLetter, deadLetterErr := encoder.NewEncoderDeadLetter(tmpDir, []string{"tag"})
		g.Expect(deadLetterErr).ToNot(HaveOccurred())

		// write first item
		diskLocation1, deadLetterErr := encoderDeadLetter.Write([]byte(`hello world`)) //echo -n "hello world" | base64 -> aGVsbG8gd29ybGQ=
		g.Expect(deadLetterErr).ToNot(HaveOccurred())
		// write second item
		diskLocation2, deadLetterErr := encoderDeadLetter.Write([]byte(`this is item 2`)) // echo -n "this is item 2" | base64 -> dGhpcyBpcyBpdGVtIDI=
		g.Expect(deadLetterErr).ToNot(HaveOccurred())

		// read first item
		data, deadLetterErr := encoderDeadLetter.Read(diskLocation1.StartIndex, diskLocation1.Size)
		g.Expect(deadLetterErr).ToNot(HaveOccurred())
		g.Expect(data).To(Equal([]byte(`hello world`)))
		// read second item
		data, deadLetterErr = encoderDeadLetter.Read(diskLocation2.StartIndex, diskLocation2.Size)
		g.Expect(deadLetterErr).ToNot(HaveOccurred())
		g.Expect(data).To(Equal([]byte(`this is item 2`)))
	})
}

func TestEncoderDeadLetter_Clean(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("removes the contents of the dead letter file", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		encoderDeadLetter, deadLetterErr := encoder.NewEncoderDeadLetter(tmpDir, []string{"tag"})
		g.Expect(deadLetterErr).ToNot(HaveOccurred())

		_, deadLetterErr = encoderDeadLetter.Write([]byte(`hello world`)) //echo -n "hello world" | base64 -> aGVsbG8gd29ybGQ=
		g.Expect(deadLetterErr).ToNot(HaveOccurred())
		g.Expect(encoderDeadLetter.Clear()).ToNot(HaveOccurred())

		// read the file to make sure its correct
		data, err := os.ReadFile(filepath.Join(tmpDir, encoder.EncodeString("tag"), "deadLetter.idx"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(data).To(Equal([]byte(``)))
	})
}
