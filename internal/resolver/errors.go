package resolver

import "errors"

var (
	ErrTopicNotFound   = errors.New("topic not found in index")
	ErrInvalidLens     = errors.New("invalid lens flag")
	ErrNoLens          = errors.New("no lens flag provided")
	ErrMultipleLenses  = errors.New("only one lens flag may be used at a time")
	ErrLensFileMissing = errors.New("lens file not authored yet")
)
