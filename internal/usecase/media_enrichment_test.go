package usecase

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/media"
)

type memEnrichmentRepo struct {
	rows map[string]*media.ScreenshotEnrichment
}

func (m *memEnrichmentRepo) GetByScreenshotID(_ context.Context, id string) (*media.ScreenshotEnrichment, error) {
	return m.rows[id], nil
}

func (m *memEnrichmentRepo) Save(_ context.Context, e *media.ScreenshotEnrichment) error {
	if m.rows == nil {
		m.rows = make(map[string]*media.ScreenshotEnrichment)
	}
	cp := *e
	m.rows[e.ScreenshotID] = &cp
	return nil
}

func (m *memEnrichmentRepo) ListByStatus(_ context.Context, status string) ([]*media.ScreenshotEnrichment, error) {
	var out []*media.ScreenshotEnrichment
	for _, e := range m.rows {
		if e.Status == status {
			cp := *e
			out = append(out, &cp)
		}
	}
	return out, nil
}

func TestMediaUseCase_EnrichScreenshot_VRCCameraPhoto(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cam.jpg")
	xmp := `<?xpacket begin="" id="w"?><x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description xmlns:vrc="http://vrchat.com/ns/" vrc:WorldID="wrld_cam" xmlns:xmp="http://ns.adobe.com/xap/1.0/" xmp:CreateDate="2026:02:17 12:00:00+09:00"/></rdf:RDF></x:xmpmeta>`
	if err := os.WriteFile(path, buildTestJPEGWithXMP(xmp), 0o644); err != nil {
		t.Fatal(err)
	}
	takenAt := time.Date(2026, 2, 17, 12, 0, 0, 0, time.Local)
	screenshotRepo := newMockScreenshotRepo()
	screenshotRepo.Save(context.Background(), &media.Screenshot{
		ID: "s1", FilePath: path, WorldID: "wrld_cam", TakenAt: &takenAt,
	})
	end := takenAt.Add(time.Hour)
	playRepo := &fakePlaySessionRepo{sessions: []*activity.PlaySession{{
		ID: "ps1", InstanceID: "wrld_cam:1~hidden()~region(jp)", StartTime: takenAt.Add(-time.Minute), EndTime: &end,
	}}}
	encRepo := &memEncounterRepo{encounters: []*activity.UserEncounter{{
		VRCUserID: "usr_guest", DisplayName: "Guest", InstanceID: "wrld_cam:1~hidden()~region(jp)",
		WorldID: "wrld_cam", JoinedAt: takenAt.Add(-time.Minute),
	}}}
	enrichRepo := &memEnrichmentRepo{rows: map[string]*media.ScreenshotEnrichment{}}
	uc := NewMediaUseCase(screenshotRepo, nil, nil)
	uc.SetEnrichmentDeps(MediaEnrichmentDeps{
		PlaySessions: playRepo,
		Encounters:   encRepo,
		Enrichment:   enrichRepo,
	})

	res, err := uc.EnrichScreenshot(context.Background(), "s1", true)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != media.EnrichmentStatusSuccess {
		t.Fatalf("status = %q reason=%q", res.Status, res.SkipReason)
	}
	fields, err := media.ReadEnrichmentFieldsFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if fields.InstanceID == "" {
		t.Fatal("expected instance in file")
	}
}

func buildTestJPEGWithXMP(xmp string) []byte {
	data := []byte{0xFF, 0xD8}
	prefix := []byte("http://ns.adobe.com/xap/1.0/\x00")
	payload := append(append([]byte{}, prefix...), []byte(xmp)...)
	segLen := 2 + len(payload)
	data = append(data, 0xFF, 0xE1, byte(segLen>>8), byte(segLen&0xff))
	data = append(data, payload...)
	return append(data, 0xFF, 0xD9)
}
