package tlp

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func init() {
	EnableDebug()
}

func TestFmtJSON(t *testing.T) {
	f, err := ioutil.ReadFile("testing/raw.json")
	if err != nil {
		t.Fatalf("readfile: %q", err)
	}
	setNow := func(t *TimelineParser) *TimelineParser {
		t.now = time.Date(2019, time.September, 13, 17, 0, 0, 0, time.Local)
		return t
	}
	r, err := NewTimelineParserFromReader(bytes.NewBuffer(f), Option(setNow))
	if err != nil {
		t.Fatalf("timeline: %q", err)
	}
	if err := r.Parse(); err != nil {
		t.Fatalf("parse: %q", err)
	}
	jsonStr, err := r.FmtJSON()
	if err != nil {
		t.Fatalf("json: %q", err)
	}

	if strings.TrimSpace(string(jsonStr)) != strings.TrimSpace(string(readResfile(t))) {
		t.Errorf("unexpected result: get %s , want %s",
			string(jsonStr), string(readResfile(t)))
	}

}

func readResfile(t *testing.T) []byte {
	f, err := ioutil.ReadFile("testing/res.txt")
	if err != nil {
		t.Fatalf("readfile: %q", err)
	}
	return f
}
