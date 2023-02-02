package encoder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DanLavine/willow/internal/v1/queues/disk/encoder"
	. "github.com/onsi/gomega"
)

func TestEncoderQueue_Write(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("writes data to disk encoded", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		queueEncoder, deErr := encoder.NewEncoderQueue(tmpDir, []string{"tag"})
		g.Expect(deErr).ToNot(HaveOccurred())

		startIndex, size, deErr := queueEncoder.Write(1, []byte("hello world")) //echo -n "hello world" | base64 -> aGVsbG8gd29ybGQ=
		g.Expect(deErr).ToNot(HaveOccurred())
		g.Expect(startIndex).To(Equal(2))
		g.Expect(size).To(Equal(16))

		fileData, err := os.ReadFile(filepath.Join(tmpDir, encoder.EncodeString("tag"), "0.idx"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(fileData).To(Equal([]byte(`1@aGVsbG8gd29ybGQ=..`))) // record the ID@[base64]..
	})

	t.Run("writes multiple calls to disk encoded and in the proper format", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		queueEncoder, deErr := encoder.NewEncoderQueue(tmpDir, []string{"tag"})
		g.Expect(deErr).ToNot(HaveOccurred())

		startIndex, size, deErr := queueEncoder.Write(1, []byte("hello world")) // echo -n "hello world" | base64 -> aGVsbG8gd29ybGQ=
		g.Expect(deErr).ToNot(HaveOccurred())
		g.Expect(startIndex).To(Equal(2))
		g.Expect(size).To(Equal(16))

		startIndex, size, deErr = queueEncoder.Write(2, []byte("second call")) // echo -n "hello world" | base64 -> c2Vjb25kIGNhbGw=
		g.Expect(deErr).ToNot(HaveOccurred())
		g.Expect(startIndex).To(Equal(21))
		g.Expect(size).To(Equal(16))

		fileData, err := os.ReadFile(filepath.Join(tmpDir, encoder.EncodeString("tag"), "0.idx"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(fileData).To(Equal([]byte(`1@aGVsbG8gd29ybGQ=.2@c2Vjb25kIGNhbGw=..`)))
	})

	t.Run("returns an error if the write fails", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		queueEncoder, deErr := encoder.NewEncoderQueue(tmpDir, []string{"tag"})
		g.Expect(deErr).ToNot(HaveOccurred())

		// close all files for the disk encoder
		queueEncoder.Close()

		_, _, deErr = queueEncoder.Write(1, []byte("hello world")) // echo -n "hello world" | base64 -> aGVsbG8gd29ybGQ=
		g.Expect(deErr).To(HaveOccurred())
		g.Expect(deErr.Error()).To(ContainSubstring("Failed to write data to disk"))
	})
}

func TestEncoderQueue_Processing(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("records the processing id", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		queueEncoder, deErr := encoder.NewEncoderQueue(tmpDir, []string{"tag"})
		g.Expect(deErr).ToNot(HaveOccurred())

		deErr = queueEncoder.Processing(1)
		g.Expect(deErr).ToNot(HaveOccurred())

		fileData, err := os.ReadFile(filepath.Join(tmpDir, encoder.EncodeString("tag"), "0_processing.idx"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(fileData).To(Equal([]byte(`P1.`))) // record the ID@[base64]..
	})
}

func TestEncoderQueue_Read(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns data decoded from disk", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		queueEncoder, deErr := encoder.NewEncoderQueue(tmpDir, []string{"tag"})
		g.Expect(err).ToNot(HaveOccurred())

		startIndex, size, deErr := queueEncoder.Write(1, []byte("hello world")) // echo -n "hello world" | base64 -> aGVsbG8gd29ybGQ=
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(startIndex).To(Equal(2))
		g.Expect(size).To(Equal(16))

		data, deErr := queueEncoder.Read(startIndex, size)
		g.Expect(deErr).ToNot(HaveOccurred())
		g.Expect(data).To(Equal([]byte(`hello world`)))
	})

	t.Run("returns an error we give a wring size to read from", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		queueEncoder, deErr := encoder.NewEncoderQueue(tmpDir, []string{"tag"})
		g.Expect(err).ToNot(HaveOccurred())

		err = os.WriteFile(
			filepath.Join(tmpDir, encoder.EncodeString("tag"), "0.idx"),
			[]byte(`asdjkbas@asd4a3aascca`),
			0755,
		)
		g.Expect(err).ToNot(HaveOccurred())

		data, deErr := queueEncoder.Read(0, 3176)
		g.Expect(deErr).To(HaveOccurred())
		g.Expect(deErr.Error()).To(ContainSubstring("Failed to read data from disk"))
		g.Expect(data).To(BeNil())
	})

	t.Run("returns an error if the data is malformed for the given location", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "")
		g.Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		queueEncoder, deErr := encoder.NewEncoderQueue(tmpDir, []string{"tag"})
		g.Expect(deErr).ToNot(HaveOccurred())

		err = os.WriteFile(
			filepath.Join(tmpDir, encoder.EncodeString("tag"), "0.idx"),
			[]byte(`asdjkbas@asd4a3aascca`),
			0755,
		)
		g.Expect(err).ToNot(HaveOccurred())

		data, deErr := queueEncoder.Read(0, 16)
		g.Expect(deErr).To(HaveOccurred())
		g.Expect(deErr.Error()).To(ContainSubstring("Failed to decode data from disk"))
		g.Expect(data).To(BeNil())
	})
}

//func TestEncoderQueue_OverwriteLast(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	t.Run("writes data to if nothing is currently there", func(t *testing.T) {
//		tmpDir, err := os.MkdirTemp("", "")
//		g.Expect(err).ToNot(HaveOccurred())
//		defer os.RemoveAll(tmpDir)
//
//		queueEncoder, err := encoder.encoder.NewEncoderQueue(tmpDir, "queue", "tag")
//		g.Expect(err).ToNot(HaveOccurred())
//
//		location, err := queueEncoder.OverwriteLast([]byte("hello world")) //echo -n "hello world" | base64 -> aGVsbG8gd29ybGQ=
//		g.Expect(err).ToNot(HaveOccurred())
//
//		g.Expect(location.StartIndex).To(Equal(0))
//		g.Expect(location.Size).To(Equal(16))
//		g.Expect(location.RetryCount).To(Equal(0))
//
//		fileData, err := os.ReadFile(filepath.Join(tmpDir, encoder.encoder.EncodeString("queue"), encoder.encoder.EncodeString("tag"), "0.idx"))
//		g.Expect(err).ToNot(HaveOccurred())
//		g.Expect(fileData).To(Equal([]byte(`aGVsbG8gd29ybGQ=..`)))
//	})
//
//	t.Run("overwrites the last index and then can continue writing", func(t *testing.T) {
//		tmpDir, err := os.MkdirTemp("", "")
//		g.Expect(err).ToNot(HaveOccurred())
//		defer os.RemoveAll(tmpDir)
//
//		queueEncoder, err := encoder.encoder.NewEncoderQueue(tmpDir, "queue", "tag")
//		g.Expect(err).ToNot(HaveOccurred())
//
//		// normal write
//		location, err := queueEncoder.Write([]byte("hello world")) //echo -n "hello world" | base64 -> aGVsbG8gd29ybGQ=
//		g.Expect(err).ToNot(HaveOccurred())
//
//		g.Expect(location.StartIndex).To(Equal(0))
//		g.Expect(location.Size).To(Equal(16))
//		g.Expect(location.RetryCount).To(Equal(0))
//
//		fileData, err := os.ReadFile(filepath.Join(tmpDir, encoder.encoder.EncodeString("queue"), encoder.encoder.EncodeString("tag"), "0.idx"))
//		g.Expect(err).ToNot(HaveOccurred())
//		g.Expect(fileData).To(Equal([]byte(`aGVsbG8gd29ybGQ=..`)))
//
//		// overwrite
//		location, err = queueEncoder.OverwriteLast([]byte("hello")) //echo -n "hello world" | base64 -> aGVsbG8=
//		g.Expect(err).ToNot(HaveOccurred())
//
//		g.Expect(location.StartIndex).To(Equal(0))
//		g.Expect(location.Size).To(Equal(8))
//		g.Expect(location.RetryCount).To(Equal(0))
//
//		fileData, err = os.ReadFile(filepath.Join(tmpDir, encoder.encoder.EncodeString("queue"), encoder.encoder.EncodeString("tag"), "0.idx"))
//		g.Expect(err).ToNot(HaveOccurred())
//		// NOTE the first set of '..' During a normal load this indicates the end of the data. The remaning bits are garbage
//		// and will be overwritten
//		g.Expect(fileData).To(Equal([]byte(`aGVsbG8=..9ybGQ=..`)))
//
//		// normal write should append to the end
//		location, err = queueEncoder.Write([]byte("world")) //echo -n "world" | base64 -> d29ybGQ=
//		g.Expect(err).ToNot(HaveOccurred())
//
//		g.Expect(location.StartIndex).To(Equal(9))
//		g.Expect(location.Size).To(Equal(8))
//		g.Expect(location.RetryCount).To(Equal(0))
//
//		fileData, err = os.ReadFile(filepath.Join(tmpDir, encoder.encoder.EncodeString("queue"), encoder.encoder.EncodeString("tag"), "0.idx"))
//		g.Expect(err).ToNot(HaveOccurred())
//		g.Expect(fileData).To(Equal([]byte(`aGVsbG8=.d29ybGQ=..`)))
//	})
//
//	t.Run("overwrite cleans up the update file", func(t *testing.T) {
//		tmpDir, err := os.MkdirTemp("", "")
//		g.Expect(err).ToNot(HaveOccurred())
//		defer os.RemoveAll(tmpDir)
//
//		queueEncoder, err := encoder.encoder.NewEncoderQueue(tmpDir, "queue", "tag")
//		g.Expect(err).ToNot(HaveOccurred())
//
//		// normal write
//		_, err = queueEncoder.Write([]byte("hello world")) //echo -n "hello world" | base64 -> aGVsbG8gd29ybGQ=
//		g.Expect(err).ToNot(HaveOccurred())
//
//		// overwrite
//		_, err = queueEncoder.OverwriteLast([]byte("hello")) //echo -n "hello world" | base64 -> aGVsbG8=
//		g.Expect(err).ToNot(HaveOccurred())
//
//		// clean update file
//		fileData, err := os.ReadFile(filepath.Join(tmpDir, encoder.encoder.EncodeString("queue"), encoder.encoder.EncodeString("tag"), "update.idx"))
//		g.Expect(err).ToNot(HaveOccurred())
//		g.Expect(fileData).To(Equal([]byte(``)))
//	})
//}
//
//func TestEncoderQueue_Retry(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	t.Run("records a remove index in the 0_retrty.idx", func(t *testing.T) {
//		tmpDir, err := os.MkdirTemp("", "")
//		g.Expect(err).ToNot(HaveOccurred())
//		defer os.RemoveAll(tmpDir)
//
//		queueEncoder, err := encoder.encoder.NewEncoderQueue(tmpDir, "queue", "tag")
//		g.Expect(err).ToNot(HaveOccurred())
//
//		location, err := queueEncoder.Write([]byte("hello world"))
//		g.Expect(err).ToNot(HaveOccurred())
//
//		err = queueEncoder.Remove(location)
//		g.Expect(err).ToNot(HaveOccurred())
//
//		fileData, err := os.ReadFile(filepath.Join(tmpDir, encoder.encoder.EncodeString("queue"), encoder.encoder.EncodeString("tag"), "0_retry.idx"))
//		g.Expect(err).ToNot(HaveOccurred())
//		g.Expect(fileData).To(Equal([]byte(`#0.`)))
//	})
//
//	t.Run("records multiple retry and remove indexes the 0_retrty.idx", func(t *testing.T) {
//		tmpDir, err := os.MkdirTemp("", "")
//		g.Expect(err).ToNot(HaveOccurred())
//		defer os.RemoveAll(tmpDir)
//
//		queueEncoder, err := encoder.encoder.NewEncoderQueue(tmpDir, "queue", "tag")
//		g.Expect(err).ToNot(HaveOccurred())
//
//		locationOne, err := queueEncoder.Write([]byte("one")) // b25l
//		g.Expect(err).ToNot(HaveOccurred())
//
//		locationTwo, err := queueEncoder.Write([]byte("two")) // dHdv
//		g.Expect(err).ToNot(HaveOccurred())
//
//		locationThree, err := queueEncoder.Write([]byte("three")) // dGhyZWU=
//		g.Expect(err).ToNot(HaveOccurred())
//
//		err = queueEncoder.Retry(locationOne)
//		g.Expect(err).ToNot(HaveOccurred())
//
//		err = queueEncoder.Remove(locationOne)
//		g.Expect(err).ToNot(HaveOccurred())
//
//		err = queueEncoder.Remove(locationThree)
//		g.Expect(err).ToNot(HaveOccurred())
//
//		err = queueEncoder.Retry(locationTwo)
//		g.Expect(err).ToNot(HaveOccurred())
//
//		fileData, err := os.ReadFile(filepath.Join(tmpDir, encoder.encoder.EncodeString("queue"), encoder.encoder.EncodeString("tag"), "0_retry.idx"))
//		g.Expect(err).ToNot(HaveOccurred())
//		g.Expect(fileData).To(Equal([]byte(`@0.#0.#11.@5.`)))
//	})
//
//	t.Run("returns an error if the location is nil", func(t *testing.T) {
//		tmpDir, err := os.MkdirTemp("", "")
//		g.Expect(err).ToNot(HaveOccurred())
//		defer os.RemoveAll(tmpDir)
//
//		queueEncoder, err := encoder.encoder.NewEncoderQueue(tmpDir, "queue", "tag")
//		g.Expect(err).ToNot(HaveOccurred())
//
//		err = queueEncoder.Remove(nil)
//		g.Expect(err).To(HaveOccurred())
//		g.Expect(err.Error()).To(Equal("Received a nil location"))
//	})
//}
//
//func TestEncoderQueue_Remove(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	t.Run("records a retry count in the 0_retrty.idx", func(t *testing.T) {
//		tmpDir, err := os.MkdirTemp("", "")
//		g.Expect(err).ToNot(HaveOccurred())
//		defer os.RemoveAll(tmpDir)
//
//		queueEncoder, err := encoder.encoder.NewEncoderQueue(tmpDir, "queue", "tag")
//		g.Expect(err).ToNot(HaveOccurred())
//
//		location, err := queueEncoder.Write([]byte("hello world"))
//		g.Expect(err).ToNot(HaveOccurred())
//
//		err = queueEncoder.Retry(location)
//		g.Expect(err).ToNot(HaveOccurred())
//
//		fileData, err := os.ReadFile(filepath.Join(tmpDir, encoder.encoder.EncodeString("queue"), encoder.encoder.EncodeString("tag"), "0_retry.idx"))
//		g.Expect(err).ToNot(HaveOccurred())
//		g.Expect(fileData).To(Equal([]byte(`@0.`)))
//	})
//
//	t.Run("records multiple retry counts in the 0_retrty.idx", func(t *testing.T) {
//		tmpDir, err := os.MkdirTemp("", "")
//		g.Expect(err).ToNot(HaveOccurred())
//		defer os.RemoveAll(tmpDir)
//
//		queueEncoder, err := encoder.encoder.NewEncoderQueue(tmpDir, "queue", "tag")
//		g.Expect(err).ToNot(HaveOccurred())
//
//		locationOne, err := queueEncoder.Write([]byte("one")) // b25l
//		g.Expect(err).ToNot(HaveOccurred())
//
//		locationTwo, err := queueEncoder.Write([]byte("two")) // dHdv
//		g.Expect(err).ToNot(HaveOccurred())
//
//		locationThree, err := queueEncoder.Write([]byte("three")) // dGhyZWU=
//		g.Expect(err).ToNot(HaveOccurred())
//
//		err = queueEncoder.Retry(locationOne)
//		g.Expect(err).ToNot(HaveOccurred())
//
//		err = queueEncoder.Retry(locationOne)
//		g.Expect(err).ToNot(HaveOccurred())
//
//		err = queueEncoder.Retry(locationThree)
//		g.Expect(err).ToNot(HaveOccurred())
//
//		err = queueEncoder.Retry(locationTwo)
//		g.Expect(err).ToNot(HaveOccurred())
//
//		fileData, err := os.ReadFile(filepath.Join(tmpDir, encoder.encoder.EncodeString("queue"), encoder.encoder.EncodeString("tag"), "0_retry.idx"))
//		g.Expect(err).ToNot(HaveOccurred())
//		g.Expect(fileData).To(Equal([]byte(`@0.@0.@11.@5.`)))
//	})
//
//	t.Run("updates the retry count on the location", func(t *testing.T) {
//		tmpDir, err := os.MkdirTemp("", "")
//		g.Expect(err).ToNot(HaveOccurred())
//		defer os.RemoveAll(tmpDir)
//
//		queueEncoder, err := encoder.encoder.NewEncoderQueue(tmpDir, "queue", "tag")
//		g.Expect(err).ToNot(HaveOccurred())
//
//		location, err := queueEncoder.Write([]byte("hello world"))
//		g.Expect(err).ToNot(HaveOccurred())
//
//		err = queueEncoder.Retry(location)
//		g.Expect(err).ToNot(HaveOccurred())
//
//		g.Expect(location.RetryCount).To(Equal(1))
//	})
//
//	t.Run("returns an error if the location is nil", func(t *testing.T) {
//		tmpDir, err := os.MkdirTemp("", "")
//		g.Expect(err).ToNot(HaveOccurred())
//		defer os.RemoveAll(tmpDir)
//
//		queueEncoder, err := encoder.encoder.NewEncoderQueue(tmpDir, "queue", "tag")
//		g.Expect(err).ToNot(HaveOccurred())
//
//		err = queueEncoder.Retry(nil)
//		g.Expect(err).To(HaveOccurred())
//		g.Expect(err.Error()).To(Equal("Received a nil location"))
//	})
//}
