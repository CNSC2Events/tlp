tlp
---

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
