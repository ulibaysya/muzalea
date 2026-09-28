package slog

import "log/slog"

var DiscardLogger = slog.New(slog.DiscardHandler)
