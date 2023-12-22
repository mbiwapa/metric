package sl

import "log/slog"

// Err returns an slog.Attr for the given error.
func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
