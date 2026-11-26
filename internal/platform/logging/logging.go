package logging

import (
	"io"
	"log/slog"
)

type Options struct {
	Service string
	Version string
	Level   slog.Level
	Format  string // text for humans and json for collectors
}

func New(w io.Writer, o Options) *slog.Logger {
	opts := &slog.HandlerOptions{Level: o.Level}
	var h slog.Handler
	if o.Format == "json" {
		h = slog.NewJSONHandler(w, opts)
	} else {
		// short time attribute changes
		opts.ReplaceAttr = shorTime
		h = slog.NewTextHandler(w, opts)
	}
	return slog.New(h).With(slog.String("service", o.Service), slog.String("version", o.Version))
}

// short time is the function which will return the time of the day in short hand format
func shorTime(groups []string, a slog.Attr) slog.Attr {
	//  it checks the if the attribute key is time and len of the group is zero it converts the value to the formatted time ...r
	if a.Key == slog.TimeKey && len(groups) == 0 {
		a.Value = slog.StringValue(a.Value.Time().Format("14:04:05.000"))
	}
	return a
}

// ParseLevel passes the debug , info , warrn or error in uppper and lower case
func ParseLevel(s string) (slog.Level, error) {
	var l slog.Level
	err := l.UnmarshalText([]byte(s))
	return l, err

}
