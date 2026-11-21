package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLiveOverlaySSHGrammar(t *testing.T) {
	for _, c := range []struct{ args, want string }{{"live L1 overlay get", "LIVE L1 OVERLAY GET"}, {"live L1 overlay set builtin-translation 7", "LIVE L1 OVERLAY SET builtin-translation 7"}, {"live L1 overlay clear 7", "LIVE L1 OVERLAY CLEAR 7"}} {
		got, e := commandFor(strings.Fields(c.args))
		if e != nil || got.String() != c.want {
			t.Fatalf("%s: %s %v", c.args, got.String(), e)
		}
	}
	for _, v := range []string{"live L1 overlay clear -1", "live L1 overlay set builtin-translation", "live L1 overlay clear 9223372036854775808"} {
		if _, e := commandFor(strings.Fields(v)); e == nil {
			t.Fatal("accepted invalid overlay mutation", v)
		}
	}
}
func TestLiveOverlayHTTPUsesExplicitSequenceWithoutRefetch(t *testing.T) {
	for _, v := range []struct {
		args   string
		design any
	}{{"live L1 overlay set builtin-translation 7", "builtin-translation"}, {"live L1 overlay clear 7", nil}} {
		p, e := planHTTP(strings.Fields(v.args), false)
		if e != nil {
			t.Fatal(e)
		}
		if p.operation != "setLiveOverlay" || p.options.IfMatch != `"live-7"` || p.conditional {
			t.Fatalf("%+v", p)
		}
		data, e := json.Marshal(p.options.JSON)
		if e != nil {
			t.Fatal(e)
		}
		var body map[string]any
		if e = json.Unmarshal(data, &body); e != nil {
			t.Fatal(e)
		}
		if len(body) != 1 || body["design_id"] != v.design {
			t.Fatalf("%s", data)
		}
	}
	p, e := planHTTP([]string{"live", "L1", "overlay", "get"}, false)
	if e != nil || p.operation != "getLiveOverlay" || p.options.IfMatch != "" {
		t.Fatalf("%+v %v", p, e)
	}
}
