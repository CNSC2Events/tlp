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

func TestGetVersus(t *testing.T) {
	e := &Event{
		VS: Versus{
			P1:      "scnace",
			P2:      "Astral",
			P1Score: "2",
			P2Score: "1",
		},
	}
	if e.GetVersus() != "scnace vs Astral (2:1)" {
		t.Errorf("versus: test player has score failed")
	}
	e2 := &Event{
		VS: Versus{
			P1: "scnace",
			P2: "Astral",
		},
	}
	if e2.GetVersus() != "scnace vs Astral " {
		t.Errorf("versus: test player dont have score failed")
	}
}

func TestFmtJSON(t *testing.T) {
	f, err := ioutil.ReadFile("testing/raw.json")
	if err != nil {
		t.Fatalf("readfile: %q", err)
	}
	sh, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("set timezone: %q", err)
	}
	setNow := func(t *TimelineParser) *TimelineParser {
		t.now = time.Date(2019, time.September, 13, 17, 0, 0, 0, sh)
		return t
	}
	r, err := NewTimelineParserFromReader(
		bytes.NewBuffer(f),
		Option(setNow),
	)
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
