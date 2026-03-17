package activity

import (
	"regexp"
	"time"
)

// EventKind identifies the type of parsed log event.
type EventKind int

const (
	EventKindUnknown EventKind = iota
	EventKindEncounter
	EventKindSession
)

// EncounterAction represents join or leave.
const (
	EncounterActionJoin  = "join"
	EncounterActionLeave = "leave"
)

// SessionEventType represents session start or end.
const (
	SessionEventStart = "start"
	SessionEventEnd   = "end"
)

// ParsedEvent is a discriminated union of log-derived events.
type ParsedEvent interface {
	Kind() EventKind
}

// EncounterEvent is emitted when a player joins or leaves the instance.
type EncounterEvent struct {
	VRCUserID     string
	DisplayName   string
	Action        string // EncounterActionJoin or EncounterActionLeave
	InstanceID    string
	EncounteredAt time.Time
}

// Kind implements ParsedEvent.
func (EncounterEvent) Kind() EventKind { return EventKindEncounter }

// SessionEvent is emitted when the local user starts or ends an instance session.
type SessionEvent struct {
	Type       string // SessionEventStart or SessionEventEnd
	InstanceID string
	OccurredAt time.Time
}

// Kind implements ParsedEvent.
func (SessionEvent) Kind() EventKind { return EventKindSession }

// LogParser parses VRChat output_log.txt lines into events.
type LogParser struct {
	// encounterPatterns are applied in order; first match wins.
	encounterPatterns []encounterPattern
	// sessionPatterns for session start/end.
	sessionPatterns []sessionPattern
}

type encounterPattern struct {
	re     *regexp.Regexp
	action string
}

type sessionPattern struct {
	re   *regexp.Regexp
	kind string
}

// Default log patterns. Extensible via table-driven config.
var (
	// OnPlayerJoined DisplayName (usr_xxxx) or OnPlayerJoined DisplayName
	encounterJoinRE = regexp.MustCompile(`(?i)OnPlayerJoined\s+(\S.*?)\s*\((usr_[a-zA-Z0-9_-]+)\)`)
	encounterJoinNoIDRE = regexp.MustCompile(`(?i)OnPlayerJoined\s+(\S.+?)(?:\s+\(usr_|$)`)

	// OnPlayerLeft DisplayName (usr_xxxx) or OnPlayerLeft DisplayName
	encounterLeaveRE = regexp.MustCompile(`(?i)OnPlayerLeft\s+(\S.*?)\s*\((usr_[a-zA-Z0-9_-]+)\)`)
	encounterLeaveNoIDRE = regexp.MustCompile(`(?i)OnPlayerLeft\s+(\S.+?)(?:\s+\(usr_|$)`)

	// Session: Joining room / Joining or Creating Room / Entering room
	sessionStartRE = regexp.MustCompile(`(?i)(?:Joining|Entering)\s+(?:or\s+Creating\s+)?(?:room|Room)[\s:]*([a-zA-Z0-9_-]+(?::[a-zA-Z0-9_-]+)?)?`)
	sessionStartWrldRE = regexp.MustCompile(`(?i)Joining\s+(wrld_[a-zA-Z0-9_-]+:[a-zA-Z0-9]+)`)

	// Session end
	sessionEndRE = regexp.MustCompile(`(?i)(?:OnLeftRoom|Left\s+room|Leaving\s+room)`)
)

// NewLogParser returns a LogParser with default patterns.
func NewLogParser() *LogParser {
	return &LogParser{
		encounterPatterns: []encounterPattern{
			{encounterJoinRE, EncounterActionJoin},
			{encounterLeaveRE, EncounterActionLeave},
			{encounterJoinNoIDRE, EncounterActionJoin},
			{encounterLeaveNoIDRE, EncounterActionLeave},
		},
		sessionPatterns: []sessionPattern{
			{sessionStartWrldRE, SessionEventStart},
			{sessionStartRE, SessionEventStart},
			{sessionEndRE, SessionEventEnd},
		},
	}
}

// ParseLine parses a single log line and returns zero or more events.
// Unparseable lines return nil, nil (no error). baseTime is used when the log line has no timestamp.
func (p *LogParser) ParseLine(line string, baseTime time.Time) ([]ParsedEvent, error) {
	line = trimLogPrefix(line)

	// Try encounter patterns first
	for _, pat := range p.encounterPatterns {
		if m := pat.re.FindStringSubmatch(line); len(m) >= 2 {
			e := p.buildEncounterEvent(m, pat.action, baseTime)
			if e != nil {
				return []ParsedEvent{e}, nil
			}
		}
	}

	// Try session patterns
	for _, pat := range p.sessionPatterns {
		if m := pat.re.FindStringSubmatch(line); len(m) >= 1 {
			e := p.buildSessionEvent(m, pat.kind, baseTime)
			if e != nil {
				return []ParsedEvent{e}, nil
			}
		}
	}

	return nil, nil
}

func trimLogPrefix(line string) string {
	// Common VRChat prefix: [Time: 123.45] or similar
	if re := regexp.MustCompile(`^\[.*?\]\s*`); re.MatchString(line) {
		return re.ReplaceAllString(line, "")
	}
	return line
}

func (p *LogParser) buildEncounterEvent(m []string, action string, baseTime time.Time) *EncounterEvent {
	displayName := ""
	vrcUserID := ""
	// m[0] = full match, m[1] = display name, m[2] = user id (optional)
	if len(m) >= 2 {
		displayName = trimDisplayName(m[1])
	}
	if len(m) >= 3 {
		vrcUserID = m[2]
	}
	if displayName == "" {
		return nil
	}
	return &EncounterEvent{
		VRCUserID:     vrcUserID,
		DisplayName:   displayName,
		Action:        action,
		InstanceID:    "",
		EncounteredAt: baseTime,
	}
}

func trimDisplayName(s string) string {
	return regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(s, "")
}

func (p *LogParser) buildSessionEvent(m []string, kind string, baseTime time.Time) *SessionEvent {
	instanceID := ""
	if len(m) >= 2 && m[1] != "" {
		instanceID = m[1]
	}
	return &SessionEvent{
		Type:       kind,
		InstanceID: instanceID,
		OccurredAt: baseTime,
	}
}
