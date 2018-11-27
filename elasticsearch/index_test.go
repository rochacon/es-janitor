package elasticsearch

import "testing"

func TestIndexDateFromName(t *testing.T) {
	i := Index{Name: "filebeat-6.5.0-2018.11.27"}
	d, err := i.DateFromName()
	if err != nil {
		t.Errorf("%s", err)
	}
	if got := d.Year(); got != 2018 {
		t.Errorf("invalid year found, expected: 2018 got: %d", got)
	}
	if got := d.Month(); got != 11 {
		t.Errorf("invalid month found, expected: 11 got: %d", got)
	}
	if got := d.Day(); got != 27 {
		t.Errorf("invalid day found, expected: 27 got: %d", got)
	}
}

func TestIndexDateFromNameInvalid(t *testing.T) {
	i := Index{Name: ".kibana"}
	_, err := i.DateFromName()
	if err == nil {
		t.Errorf("expected error found nil")
	}
	if err.Error() != "unable to find date from index name: .kibana" {
		t.Errorf("invalid error message: %s", err)
	}
}
