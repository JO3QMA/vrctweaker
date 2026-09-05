package media

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const enrichmentXMLNS = "http://vrchat-tweaker/ns/"

var (
	reEnrichmentInstanceIDAttr = regexp.MustCompile(`vrctweaker:InstanceID\s*=\s*"([^"]*)"`)
	reEnrichmentInstanceIDElem = regexp.MustCompile(`<vrctweaker:InstanceID[^>]*>([^<]*)</vrctweaker:InstanceID>`)
	reEnrichmentParticipants   = regexp.MustCompile(`<vrctweaker:Participants[^>]*>([\s\S]*?)</vrctweaker:Participants>`)
)

// ParseEnrichmentXMP extracts vrctweaker enrichment fields from raw XMP XML.
func ParseEnrichmentXMP(xmp string) EnrichmentFields {
	if xmp == "" {
		return EnrichmentFields{}
	}
	var out EnrichmentFields
	if sub := reEnrichmentInstanceIDAttr.FindStringSubmatch(xmp); len(sub) > 1 {
		out.InstanceID = strings.TrimSpace(sub[1])
	}
	if out.InstanceID == "" {
		if sub := reEnrichmentInstanceIDElem.FindStringSubmatch(xmp); len(sub) > 1 {
			out.InstanceID = strings.TrimSpace(sub[1])
		}
	}
	if sub := reEnrichmentParticipants.FindStringSubmatch(xmp); len(sub) > 1 {
		_ = json.Unmarshal([]byte(strings.TrimSpace(sub[1])), &out.Participants)
	}
	return out
}

// MergeEnrichmentIntoXMP adds or updates vrctweaker fields without touching vrc:WorldID.
func MergeEnrichmentIntoXMP(xmp string, fields EnrichmentFields) (string, error) {
	if fields.InstanceID == "" && len(fields.Participants) == 0 {
		return xmp, nil
	}
	if xmp == "" {
		return minimalXMPPacket(fields), nil
	}
	existing := ParseEnrichmentXMP(xmp)
	if existing.InstanceID != "" && fields.InstanceID != "" && existing.InstanceID != fields.InstanceID {
		return xmp, fmt.Errorf("existing instance_id %q conflicts with new %q", existing.InstanceID, fields.InstanceID)
	}
	if strings.Contains(xmp, "<rdf:Description") {
		return injectIntoDescription(xmp, fields, existing), nil
	}
	return minimalXMPPacket(fields), nil
}

func minimalXMPPacket(fields EnrichmentFields) string {
	attrs := ` xmlns:vrctweaker="` + enrichmentXMLNS + `"`
	if fields.InstanceID != "" {
		attrs += ` vrctweaker:InstanceID="` + escapeXMLAttr(fields.InstanceID) + `"`
	}
	participantsXML := participantsElement(fields.Participants)
	body := `<rdf:Description` + attrs
	if participantsXML == "" {
		body += `/>`
	} else {
		body += `>` + participantsXML + `</rdf:Description>`
	}
	return `<?xpacket begin="" id="w"?>` +
		`<x:xmpmeta xmlns:x="adobe:ns:meta/">` +
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">` +
		body +
		`</rdf:RDF></x:xmpmeta><?xpacket end="w"?>`
}

func injectIntoDescription(xmp string, fields EnrichmentFields, existing EnrichmentFields) string {
	out := xmp
	if !strings.Contains(out, `xmlns:vrctweaker="`) {
		out = strings.Replace(out, "<rdf:Description", `<rdf:Description xmlns:vrctweaker="`+enrichmentXMLNS+`"`, 1)
	}
	if fields.InstanceID != "" && existing.InstanceID == "" && !strings.Contains(out, "vrctweaker:InstanceID") {
		out = strings.Replace(out, "<rdf:Description", `<rdf:Description vrctweaker:InstanceID="`+escapeXMLAttr(fields.InstanceID)+`"`, 1)
	}
	if len(fields.Participants) > 0 && !strings.Contains(out, "<vrctweaker:Participants") {
		block := participantsElement(fields.Participants)
		if selfClose := strings.Index(out, "/>"); selfClose > 0 && strings.Contains(out[:selfClose], "<rdf:Description") {
			out = out[:selfClose] + `>` + block + `</rdf:Description>` + out[selfClose+2:]
		} else {
			out = strings.Replace(out, "</rdf:Description>", block+`</rdf:Description>`, 1)
		}
	}
	out = mirrorDCSubjects(out, fields.Participants)
	return out
}

func mirrorDCSubjects(xmp string, participants []EnrichmentParticipant) string {
	subjects := MirrorSubjects(participants, "")
	if len(subjects) == 0 || strings.Contains(xmp, "dc:subject") {
		return xmp
	}
	if !strings.Contains(xmp, `xmlns:dc="`) {
		xmp = strings.Replace(xmp, "<rdf:Description", `<rdf:Description xmlns:dc="http://purl.org/dc/elements/1.1/"`, 1)
	}
	var b strings.Builder
	b.WriteString(`<dc:subject><rdf:Bag>`)
	for _, s := range subjects {
		b.WriteString(`<rdf:li>`)
		b.WriteString(escapeXMLText(s))
		b.WriteString(`</rdf:li>`)
	}
	b.WriteString(`</rdf:Bag></dc:subject>`)
	block := b.String()
	if selfClose := strings.Index(xmp, "/>"); selfClose > 0 && strings.Contains(xmp[:selfClose], "<rdf:Description") {
		return xmp[:selfClose] + `>` + block + `</rdf:Description>` + xmp[selfClose+2:]
	}
	return strings.Replace(xmp, "</rdf:Description>", block+`</rdf:Description>`, 1)
}

func participantsElement(participants []EnrichmentParticipant) string {
	if len(participants) == 0 {
		return ""
	}
	raw, err := json.Marshal(participants)
	if err != nil {
		return ""
	}
	return `<vrctweaker:Participants>` + escapeXMLText(string(raw)) + `</vrctweaker:Participants>`
}

func escapeXMLAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	return s
}

func escapeXMLText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// MirrorSubjects returns dc:subject values for external tools.
func MirrorSubjects(participants []EnrichmentParticipant, worldName string) []string {
	var out []string
	if worldName != "" {
		out = append(out, worldName)
	}
	for _, p := range participants {
		if p.DisplayName != "" {
			out = append(out, p.DisplayName)
		}
	}
	return out
}
