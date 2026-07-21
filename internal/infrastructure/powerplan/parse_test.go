package powerplan

import "testing"

func TestParseListOutput_activeMarker(t *testing.T) {
	raw := `
Existing Power Schemes (* Active)
-----------------------------------
Power Scheme GUID: 381b4222-f694-41f0-9685-ff5bb260df2e  (Balanced)
Power Scheme GUID: 8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c  (High performance) *
Power Scheme GUID: a1841308-3541-4fab-bc81-f71556f20b4a  (Power saver)
`
	plans, err := parseListOutput(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(plans) != 3 {
		t.Fatalf("want 3, got %d", len(plans))
	}
	if plans[1].Name != "High performance" {
		t.Fatalf("active plan name = %q", plans[1].Name)
	}
	if plans[1].GUID != "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c" {
		t.Fatalf("guid = %q", plans[1].GUID)
	}
}

func TestParseListOutput_localeHeaderAndTabs(t *testing.T) {
	raw := "電源スキーム GUID:\t381b4222-f694-41f0-9685-ff5bb260df2e\t(バランス)\n" +
		"何か GUID: 8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c  (高パフォーマンス) *\n"
	plans, err := parseListOutput(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(plans) != 2 {
		t.Fatalf("want 2, got %d %#v", len(plans), plans)
	}
	if plans[0].GUID != "381b4222-f694-41f0-9685-ff5bb260df2e" || plans[0].Name != "バランス" {
		t.Fatalf("plan0=%#v", plans[0])
	}
	if plans[1].Name != "高パフォーマンス" {
		t.Fatalf("name=%q", plans[1].Name)
	}
}

func TestValidGUID(t *testing.T) {
	if !ValidGUID("381b4222-f694-41f0-9685-ff5bb260df2e") {
		t.Fatal("want valid")
	}
	if ValidGUID("381b4222-f694-41f0-9685-ff5bb260df2e; notepad") {
		t.Fatal("must reject junk")
	}
	if ValidGUID("") || ValidGUID("x") {
		t.Fatal("must reject empty/short")
	}
}

func TestResolvePresetFromPlans_knownGUID(t *testing.T) {
	plans := []Plan{
		{GUID: "381b4222-f694-41f0-9685-ff5bb260df2e", Name: "バランス"},
	}
	got, err := resolvePresetFromPlans("balanced", plans)
	if err != nil {
		t.Fatal(err)
	}
	if got != plans[0].GUID {
		t.Fatalf("got %q", got)
	}
}

func TestResolvePresetFromPlans_missing(t *testing.T) {
	_, err := resolvePresetFromPlans("balanced", []Plan{{GUID: "x", Name: "Custom"}})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestResolvePresetFromPlans_noSubstringFallback(t *testing.T) {
	plans := []Plan{{GUID: "x", Name: "Custom Balanced Mode"}}
	_, err := resolvePresetFromPlans("balanced", plans)
	if err == nil {
		t.Fatal("substring match must not succeed")
	}
}
