package processor

import (
	"context"
	"hash/crc32"
	"io"
	"time"
)

// Process creates a file from the provided path and uploads it to the store.
func (p *Processor) Process(ctx context.Context, path string) (*File, error) {
	ctime, err := p.getCTime(path)
	if err != nil {
		return nil, err
	}
	p.log.Infow(
		"Fetched ctime for path",
		"ctime", ctime,
		"path", path,
	)

	prevFile, err := p.registry.FetchLatest(ctx, path)
	if err != nil {
		return nil, err
	}

	file := &File{
		Key:       p.keyGen.GenerateKey(),
		LocalPath: path,
		CTime:     ctime,
	}

	if prevFile != nil {
		p.log.Infow(
			"Found previous file version",
			"key", prevFile.Key,
			"checksum", prevFile.Checksum,
			"ctime", prevFile.CTime,
		)
		if prevFile.CTime.Equal(file.CTime) {
			p.log.Infow(
				"Skipping previously uploaded file",
				"path", prevFile.LocalPath,
				"version", prevFile.Version,
			)
			return prevFile, nil
		}

		if err := p.attachChecksum(file); err != nil {
			return nil, err
		}

		if prevFile.Checksum == file.Checksum {
			p.log.Infow(
				"Skipping previously uploaded file",
				"path", prevFile.LocalPath,
				"version", prevFile.Version,
			)
			return prevFile, nil
		}
		file.Key = prevFile.Key
	} else if err := p.attachChecksum(file); err != nil {
		return nil, err
	}

	file, err = p.uploader.Upload(ctx, file)
	if err != nil {
		return nil, err
	}
	p.log.Infow(
		"Stored file in storage backend",
		"path", file.LocalPath,
		"etag", file.ETag,
		"version", file.Version,
	)

	file, err = p.registry.Create(ctx, file)
	if err != nil {
		return nil, err
	}
	p.log.Infow(
		"Stored file in file registry",
		"path", file.LocalPath,
	)

	return file, nil
}

func (p *Processor) getCTime(path string) (time.Time, error) {
	f, err := p.fs.Open(path)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()

	return p.ctg.GetCTime(f)
}

func (p *Processor) attachChecksum(file *File) error {
	checksum, err := p.computeChecksum(file.LocalPath)
	if err != nil {
		return err
	}
	p.log.Infow(
		"Computed checksum for path",
		"path", file.LocalPath,
		"checksum", checksum,
	)
	file.Checksum = checksum

	return nil
}

func (p *Processor) computeChecksum(path string) (Checksum, error) {
	f, err := p.fs.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	h := crc32.NewIEEE()
	io.Copy(h, f)

	return Checksum(h.Sum32()), nil
}
