package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/store"
)

func devs() []store.Device {
	t := time.Unix(1700000000, 0).UTC()
	return []store.Device{
		{MAC: "aa:00:00:00:00:01", IP: "192.168.1.10", Name: "=HYPERLINK(\"http://evil\")", Vendor: "Acme", Tags: []string{"iot", "sala"}, Known: true, Online: true, FirstSeen: t, LastSeen: t},
		{MAC: "aa:00:00:00:00:02", IP: "192.168.1.11", Name: "-2+3", Label: "@cmd", FirstSeen: t, LastSeen: t},
	}
}

func TestCSVNeutralisesFormulas(t *testing.T) {
	var b bytes.Buffer
	if err := CSV(&b, Rows(devs())); err != nil {
		t.Fatal(err)
	}
	recs, err := csv.NewReader(&b).ReadAll()
	if err != nil || len(recs) != 3 {
		t.Fatalf("%v %v", recs, err)
	}
	if recs[1][2] != `'=HYPERLINK("http://evil")` || recs[2][2] != "'-2+3" || recs[2][3] != "'@cmd" {
		t.Fatalf("formulas not neutralised: %q %q %q", recs[1][2], recs[2][2], recs[2][3])
	}
	if recs[1][7] != "iot sala" || recs[1][8] != "yes" || recs[0][0] != "mac" {
		t.Fatalf("%v", recs[1])
	}
}

func TestJSONKeepsValuesAsData(t *testing.T) {
	var b bytes.Buffer
	if err := JSON(&b, Rows(devs())); err != nil {
		t.Fatal(err)
	}
	var back []Row
	if err := json.Unmarshal(b.Bytes(), &back); err != nil || len(back) != 2 || back[0].Name != `=HYPERLINK("http://evil")` || len(back[1].Tags) != 0 {
		t.Fatalf("%+v %v", back, err)
	}
}
