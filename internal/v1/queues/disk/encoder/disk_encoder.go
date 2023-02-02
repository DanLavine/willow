package encoder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DanLavine/willow/internal/errors"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

/**
 ** File are encoded to disk with the follwing formats:
 **
 ** 0.idx - are the files with enqueud data from a producer.
 **		1. [ID]@[base64 encooded data].[ID]@[base64 encooded data]..
 **   2. ID is the identifier for the queue item. On a restart, the queue can be reconstructed so clients can still ACK a message that was previously processing
 **   3. data is base64 encoded so it will not contain special characters
 **   4. the '.' indicates a seperator entry for enqueued items
 **   5. the '..' indicator is the final entry in the queue

 ** 0_processing.idx - are the files with recorded states of 0.idx file.
 **		1. @[id].@[id].#[id].
 **   2. the '@' char indicates that an index has failed
 **   3. the '#' char indicates that an index was completed and should not be re-processed
 **   4. the '[id]' is the start location for the entry.
 **   5. the '.' char is the seperator

 ** update.idx - are the files with records of in process updates
 **		1. [base64 encoded data].[base64 encoded data]..
 **   2. the first base64 encoded data is the current value
 **   2. the second base64 encoded data is the next value
 **   During a crash this file can be used to reconstruct the last entry if
 **   it was corrupted, or never properly updated
 */

type DiskEncoder struct {
	updateFile *updateFile
	index      *index
	indexState *indexState
}

func NewDiskEncoder(baseDir string, queueTags []string) (*DiskEncoder, *v1.Error) {
	queueDir := filepath.Join(baseDir, EncodeStrings(queueTags))

	filePath, err := os.Stat(queueDir)
	if os.IsPermission(err) || os.IsNotExist(err) {
		// create the dir
		if err = os.MkdirAll(queueDir, 0755); err != nil {
			return nil, errors.FailedToCreateDir.With("", err.Error())
		}
	} else if err != nil {
		// some other error encountered
		return nil, errors.FailedToStatDir.With("", err.Error())
	} else {
		// path already exists and is not dir?
		if !filePath.IsDir() {
			return nil, errors.PathAlreadyExists.With(filePath.Name(), "to be a dir")
		}
	}

	index, indexErr := newIndex(filepath.Join(queueDir, "0.idx"))
	if err != nil {
		return nil, indexErr
	}

	indexState, stateErr := newIndexState(filepath.Join(queueDir, "0_processing.idx"))
	if err != nil {
		return nil, stateErr
	}

	updateFile, updateErr := newUpdateFile(filepath.Join(queueDir, "update.idx"))
	if err != nil {
		return nil, updateErr
	}

	return &DiskEncoder{
		updateFile: updateFile,
		index:      index,
		indexState: indexState,
	}, nil
}

// Write appends data to disk.
//
// RETURNS:
// * int - start location on disk where the write happend
// * int - size of encoded data written. Does not include things like path seperators or other info
// * error - an error encounterd during the write
func (de *DiskEncoder) Write(id uint64, data []byte) (int, int, *v1.Error) {
	startIndex, size, err := de.index.Write(id, data)
	if err != nil {
		return startIndex, size, err
	}

	return startIndex, size, nil
}

func (de *DiskEncoder) Processing(id uint64) *v1.Error {
	return de.indexState.Processing(id)
}

func (de *DiskEncoder) Read(startIndex, size int) ([]byte, *v1.Error) {
	encodedData, err := de.index.Read(startIndex, size)
	if err != nil {
		return nil, err
	}

	decodedData, err := DecodeByte(encodedData)
	if err != nil {
		return nil, errors.DecodeFailed.With(fmt.Sprintf("decode at start location: %d, for %d bytes", startIndex, size), err.Error())
	}

	return decodedData, nil
}

// // OverwriteLast can be used to update the last location on disk
// //
// // RETURNS:
// // * location - location state on disk
// // * error - an error encounterd during the write
//
//	func (de *DiskEncoder) Overwrite(data []byte, location *models.Location) error {
//		// get the current last entry
//		currentLastData, err := de.index.ReadLast()
//		if err != nil {
//			return err
//		}
//
//		// Overwrite has nothing to overwrite, so just write as normal
//		if currentLastData == nil {
//			return de.Write(data, location)
//		}
//
//		// save the current and next values to the update file
//		nextEncodedData := EncodeByteWithEnding(data)
//		if err = de.updateFile.Write(currentLastData, nextEncodedData); err != nil {
//			return err
//		}
//
//		// overwrite last index
//		location, err := de.index.OverwriteLast(nextEncodedData)
//		if err != nil {
//			return err
//		}
//
//		// clear the update file and return location
//		return de.updateFile.Clear()
//	}
//
//	func (de *DiskEncoder) Retry(l *Location) error {
//		if l == nil {
//			return fmt.Errorf("Received a nil location")
//		}
//
//		if err := de.indexState.Retry(l); err != nil {
//			return err
//		}
//
//		l.RetryCount++
//		return nil
//	}
func (de *DiskEncoder) Remove(id uint64) *v1.Error {
	return de.indexState.Delete(id)
}

func (de *DiskEncoder) Close() {
	de.updateFile.Close()
	de.index.Close()
	de.indexState.Close()
}
