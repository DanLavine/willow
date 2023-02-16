package disk

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/DanLavine/willow/internal/errors"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

// record info about he queue
func recordQueueInfo(baseDir string, create *v1.Create) *v1.Error {
	data, err := create.ToBytes()
	if err != nil {
		return err
	}

	file, err := CreateOrOpenFile(baseDir, create.Name, "queue.info")
	if err != nil {
		return err
	}
	defer file.Close()

	if _, writeErr := file.Write(data); writeErr != nil {
		return errors.WriteFailed.With(fmt.Sprintf("write at file '%s' to succeed", file.Name()), err.Error())
	}

	return nil
}

// load info about the queue
func loadQueueInfo(baseDir, queueName string) (*v1.Create, *v1.Error) {
	filePath, filePathErr := FilePath(baseDir, queueName)
	if filePathErr != nil {
		return nil, filePathErr
	}

	file, err := os.Open(filepath.Join(filePath, "queue.info"))
	if err != nil {
		return nil, errors.FileNotFound.With("", err.Error())
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.ReadFailed.With("", err.Error())
	}

	create := &v1.Create{}
	if err = json.Unmarshal(data, create); err != nil {
		return nil, errors.DecodeFailed.With("", err.Error())
	}

	return create, nil
}
