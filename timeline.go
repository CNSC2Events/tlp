package tlp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	timeFmt = `January 2, 2006 - 15:04 UTC`
)

var (
	MaxCountDuration = 20 * time.Minute
)

func EnableDebug() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

// TLMatchPage is a raw struct defined by
// https://liquipedia.net/starcraft2/api.php?action=parse&format=json&page=Liquipedia:Upcoming_and_ongoing_matches
// And removed useless contents
type TLMatchPage struct {
	Parse struct {
		Title  string `json:"title"`
		Pageid int    `json:"pageid"`
		Revid  int    `json:"revid"`
		Text   struct {
			RawHTML string `json:"*"`
		} `json:"text"`
	} `json:"parse"`
}

type Versus struct {
	P1      string
	P2      string
	P1Score string
	P2Score string
}

type Event struct {
	StartAt          time.Time
	IsOnGoing        bool
	VS               Versus
	TimeCountingDown string
	Series           string
	DetailURL        *url.URL
}

type TimelineParser struct {
	body     []byte
	Timezone *time.Location
	RevID    string
	Events   []*Event

	now time.Time
}

type Option func(*TimelineParser) *TimelineParser

func NewTimelineParser(respBody []byte, opts ...Option) *TimelineParser {

	tp := &TimelineParser{
		body: respBody,
		now:  time.Now(),
	}

	for _, opt := range opts {
		tp = opt(tp)
	}

	return tp
}

func NewTimelineParserFromReader(r io.Reader, opts ...Option) (*TimelineParser, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	tlMatches := new(TLMatchPage)

	if err := json.Unmarshal(body, tlMatches); err != nil {
		return nil, err
	}
	return NewTimelineParser(
		[]byte(tlMatches.Parse.Text.RawHTML),
		opts...), nil
}

func (tp *TimelineParser) SetTimezone(name string) error {
	tz, err := time.LoadLocation(name)
	if err != nil {
		return fmt.Errorf("parser: SetTimezone: %w", err)
	}
	tp.Timezone = tz
	return nil
}

func (tp *TimelineParser) getCountDownDuration(s *goquery.Selection) (time.Duration, error) {
	t, err := tp.getCurrentTiming(s)
	if err != nil {
		return 0, err
	}
	if tp.Timezone == nil {
		if err := tp.SetTimezone("Asia/Shanghai"); err != nil {
			return 0, fmt.Errorf("parser: isOnGoing: %w", err)
		}
	}

	countDown := tp.now.Sub(t.In(tp.Timezone))
	return countDown, nil
}

func (tp *TimelineParser) getCurrentTiming(s *goquery.Selection) (time.Time, error) {

	t, err := time.Parse(timeFmt, s.Find(`.timer-object-countdown-only`).Text())
	if err != nil {
		return time.Time{}, fmt.Errorf("parser: isOnGoing: %w", err)
	}

	return t, nil
}

func (tp *TimelineParser) Parse() error {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(tp.body))
	if err != nil {
		return err
	}
	doc.Find(`.infobox_matches_content`).
		Each(func(idx int, s *goquery.Selection) {
			e := new(Event)
			ct, err := tp.getCurrentTiming(s)
			if err != nil {
				log.Debug().Err(err)
				return
			}
			e.StartAt = ct
			e.buildVS(s)
			countdown, err := tp.getCountDownDuration(s)
			if err != nil {
				log.Debug().Err(err)
				return
			}
			if 0 < int64(countdown) && int64(countdown) < int64(MaxCountDuration) {
				e.buildSeriesByTag(s, ".matchticker-tournament-wrapper")
				e.TimeCountingDown = countdown.String()
				tp.Events = append(tp.Events, e)
			}

			if int64(countdown) <= 0 {
				e.buildSeriesByTag(s, ".match-filler > div")
				if err := e.buildDetailURL(s); err != nil {
					log.Debug().Err(err)
					tp.Events = append(tp.Events, e)
					return
				}
				tp.Events = append(tp.Events, e)
			}

			return
		})
	return nil
}

func (tp *TimelineParser) FmtJSON(opts ...Option) ([]byte, error) {

	for _, opt := range opts {
		tp = opt(tp)
	}

	var es []*Event
	for _, e := range tp.Events {
		es = append(es, e)
	}

	return json.Marshal(es)

}

func (e *Event) GetVersus() string {
	vs := e.VS
	if vs.P1Score != "" && vs.P2Score != "" {
		return fmt.Sprintf("%s vs %s (%s:%s)",
			vs.P1, vs.P2, vs.P1Score, vs.P2Score)
	}
	return fmt.Sprintf("%s vs %s ", vs.P1, vs.P2)
}

func (e *Event) buildDetailURL(s *goquery.Selection) error {
	detail := s.Find(`.match-filler > div > div > a`)
	if detail.Length() == 0 {
		return errors.New("buildDetailURL: no match detail URL detected")
	}
	detailURL, ok := detail.Attr("href")
	if !ok {
		return errors.New("buildDetailURL: match detail has no wiki page")
	}

	u, err := url.Parse("https://liquipedia.net" + detailURL)
	if err != nil {
		return fmt.Errorf("buildDetailURL : %w", err)
	}

	e.DetailURL = u

	return nil
}

func (e *Event) buildSeriesByTag(s *goquery.Selection, t string) {
	tournament := s.Find(t).Text()
	e.Series = strings.TrimSpace(tournament)
}

func (e *Event) buildVS(s *goquery.Selection) {
	lp := s.Find(`.team-left`).Text()
	rp := s.Find(`.team-right`).Text()
	versus := s.Find(`.versus`).Text()
	vs := strings.Replace(versus, "\n", "", -1)
	if strings.Contains(vs, "vs") {
		e.VS = Versus{
			P1: trimPlayer(lp),
			P2: trimPlayer(rp),
		}
		return
	}
	score := strings.Split(versus, ":")
	var s1, s2 string
	if len(score) < 2 {
		s1 = "0"
		s2 = "0"
	} else {
		s1 = score[0]
		s2 = score[1]
	}
	v := Versus{
		P1:      lp,
		P2:      rp,
		P1Score: s1,
		P2Score: s2,
	}
	e.VS = v
	return
}

func trimPlayer(p string) string {
	return strings.Replace(strings.TrimSpace(p), "\n", "", -1)
}
