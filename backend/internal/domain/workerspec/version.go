package workerspec

import "errors"

type Version uint16

const VersionV1 Version = 1

var ErrUnsupportedVersion = errors.New("workerspec version is unsupported")
