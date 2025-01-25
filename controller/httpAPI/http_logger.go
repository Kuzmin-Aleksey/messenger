package httpAPI

import (
	"fmt"
	"math/bits"
	"strconv"
	"time"
)

const timeLayout = "2006/01/02 15:04:05.00"

var LogFormat = "%-" + strconv.Itoa(len(timeLayout)) + "s │%-7s│%-25s│%-5s│%-16s│%-12s│%-12s│%-12s│\n"

const l = len(timeLayout)
const (
	defaultColor = "\033[38;5;37m"
	colorHttpOk  = "\033[38;5;28m"
	colorHttp400 = "\033[38;5;184m"
	colorHttp500 = "\033[38;5;196m"
)

type HttpLogger struct {
}

func NewHttpLogger() *HttpLogger {
	fmt.Print(defaultColor)
	fmt.Printf(LogFormat, "DATE TIME", "METHOD", "URL", "CODE", "REMOTE ADDR", "READ", "WRITE", "TIME")
	return &HttpLogger{}
}

func (l *HttpLogger) Log(method string, url string, code int, remoteAddr string, read uint64, write uint64, t time.Duration) {
	fmt.Printf(LogFormat, time.Now().Format(timeLayout), method, url, l.formatCode(code), remoteAddr,
		l.formatBytes(read), l.formatBytes(write), l.formatDuration(t))
}

func (l *HttpLogger) formatCode(code int) string {
	c := colorHttpOk
	if code >= 400 && code < 500 {
		c = colorHttp400
	} else if code >= 500 {
		c = colorHttp500
	}
	return fmt.Sprintf("%s%-5d%s", c, code, defaultColor)
}

func (l *HttpLogger) formatBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}

	base := uint(bits.Len64(bytes) / 10)
	val := float64(bytes) / float64(uint64(1<<(base*10)))

	return fmt.Sprintf("%.1f %ciB", val, " KMGTPE"[base])
}

func (l *HttpLogger) formatDuration(d time.Duration) string {
	d = d.Round(time.Millisecond)
	ms := d.Milliseconds()
	s := ms / 1000
	if s == 0 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%ds%dms", s, ms-(s*1000))
}
