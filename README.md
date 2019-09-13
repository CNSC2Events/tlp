tlp
---
[![Build Status](https://cloud.drone.io/api/badges/CNSC2Events/tlp/status.svg)](https://cloud.drone.io/CNSC2Events/tlp)
[![codecov](https://codecov.io/gh/CNSC2Events/tlp/graph/badge.svg)](https://codecov.io/gh/CNSC2Events/tlp)
[![Go Report Card](https://goreportcard.com/badge/github.com/CNSC2Events/tlp)](https://goreportcard.com/report/github.com/CNSC2Events/tlp)
[![godoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat)](https://godoc.org/github.com/CNSC2Events/tlp)

## Intro

tlp is the tl.net starcraft2 event parser extracted from [Astral](github.com/scbizu/Astral)


## Usage

```go
    //assume that you already had the tl response
    p := NewTimelineParser(resp)
    if err:= p.Parse();err != nil {
        // handle error
    }
    // get json info
    jsonOut,err:= p.FmtJSON()
    if err!=nil {
        // handle error
    }
```
