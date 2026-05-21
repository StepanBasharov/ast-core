package log

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

type zerologLogger struct {
	logger zerolog.Logger
}

// config holds the options used to construct a zerologLogger.
type config struct {
	writer io.Writer
	level  zerolog.Level
	pretty bool
}

// Option is a functional option for configuring NewZerologLogger.
type Option func(*config)

// NewZerologLogger creates a Logger backed by zerolog.
// Defaults: JSON output to os.Stderr at Info level.
// Use Option functions to override writer, level, or enable pretty console output.
func NewZerologLogger(opts ...Option) Logger {
	cfg := config{
		writer: os.Stderr,
		level:  zerolog.InfoLevel,
		pretty: false,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	var writer io.Writer = cfg.writer
	if cfg.pretty {
		writer = zerolog.ConsoleWriter{
			Out:        cfg.writer,
			TimeFormat: "15:04:05",
		}
	}

	zl := zerolog.New(writer).
		Level(cfg.level).
		With().
		Timestamp().
		Logger()

	return &zerologLogger{logger: zl}
}

func (l *zerologLogger) Info(msg string, fields ...FieldLogger) {
	applyFields(l.logger.Info(), fields).Msg(msg)
}

func (l *zerologLogger) Error(msg string, fields ...FieldLogger) {
	applyFields(l.logger.Error(), fields).Msg(msg)
}

func (l *zerologLogger) Debug(msg string, fields ...FieldLogger) {
	applyFields(l.logger.Debug(), fields).Msg(msg)
}

func (l *zerologLogger) Warn(msg string, fields ...FieldLogger) {
	applyFields(l.logger.Warn(), fields).Msg(msg)
}

// applyFields attaches structured fields to a zerolog event.
// error values are serialised via AnErr to produce a plain string in JSON output;
// all other values are serialised via Interface.
func applyFields(event *zerolog.Event, fields []FieldLogger) *zerolog.Event {
	for _, f := range fields {
		if err, ok := f.Value.(error); ok {
			event = event.AnErr(f.Key, err)
		} else {
			event = event.Interface(f.Key, f.Value)
		}
	}
	return event
}

type config struct {
	writer io.Writer
	level  zerolog.Level
	pretty bool
}

// Option configures the logger.
type Option func(*config)

// NewZerologLogger creates a Logger backed by zerolog.
// Default: JSON output to os.Stderr at Info level.
func NewZerologLogger(opts ...Option) Logger {
	cfg := config{
		writer: os.Stderr,
		level:  zerolog.InfoLevel,
		pretty: false,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	var writer io.Writer = cfg.writer
	if cfg.pretty {
		writer = zerolog.ConsoleWriter{
			Out:        cfg.writer,
			TimeFormat: "15:04:05",
		}
	}

	zl := zerolog.New(writer).
		Level(cfg.level).
		With().
		Timestamp().
		Logger()

	return &zerologLogger{logger: zl}
}

func (l *zerologLogger) Info(msg string, fields ...FieldLogger) {
	applyFields(l.logger.Info(), fields).Msg(msg)
}

func (l *zerologLogger) Error(msg string, fields ...FieldLogger) {
	applyFields(l.logger.Error(), fields).Msg(msg)
}

func (l *zerologLogger) Debug(msg string, fields ...FieldLogger) {
	applyFields(l.logger.Debug(), fields).Msg(msg)
}

func (l *zerologLogger) Warn(msg string, fields ...FieldLogger) {
	applyFields(l.logger.Warn(), fields).Msg(msg)
}

func applyFields(event *zerolog.Event, fields []FieldLogger) *zerolog.Event {
	for _, f := range fields {
		if err, ok := f.Value.(error); ok {
			event = event.AnErr(f.Key, err)
		} else {
			event = event.Interface(f.Key, f.Value)
		}
	}
	return event
}
