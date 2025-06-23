package api

type ArchiveFormat string

const (
	ArchiveTar     ArchiveFormat = "tar"
	ArchiveTarGzip               = "tgz"
	ArchiveTarZstd               = "tzst"
)

var ArchiveFormatAll = []ArchiveFormat{
	ArchiveTar,
	ArchiveTarGzip,
	ArchiveTarZstd,
}

func (f ArchiveFormat) String() string {
	return f.Name()
}

func (f ArchiveFormat) Name() string {
	return string(f)
}

func (f ArchiveFormat) ContentType() string {
	switch f {
	case ArchiveTar:
		return "application/x-tar"

	case ArchiveTarGzip:
		// https://www.rfc-editor.org/rfc/rfc6713.html
		return "application/gzip"

	case ArchiveTarZstd:
		// https://www.rfc-editor.org/rfc/rfc8878
		return "application/zstd"
	}

	return "application/octet-stream"
}

func (f ArchiveFormat) FileExtension() string {
	switch f {
	case ArchiveTar:
		return ".tar"

	case ArchiveTarGzip:
		return ".tar.gz"

	case ArchiveTarZstd:
		return ".tar.zst"
	}

	return ".bin"
}
