package activity

import (
	"regexp"
	"strings"
	"time"
)

// EventKind identifies the type of parsed log event.
type EventKind int

const (
	EventKindUnknown EventKind = iota
	EventKindEncounter
	EventKindSession
	EventKindDestination
	EventKindRoomName
	EventKindAvatarSwitch
	EventKindVideoPlayback
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

// DestinationSetEvent is emitted for [Behaviour] Destination set: wrld_...:instance~...
type DestinationSetEvent struct {
	WorldID      string
	InstanceID   string // numeric / alphanumeric segment after first colon (before ~)
	InstanceType string // private, hidden, public, etc.
	OwnerUserID  string // often usr_...
	Region       string
	OccurredAt   time.Time
	FullInstance string // wrld_xxx:instance~... for session alignment
}

// Kind implements ParsedEvent.
func (DestinationSetEvent) Kind() EventKind { return EventKindDestination }

// RoomNameEvent is emitted for Entering Room: <name>.
type RoomNameEvent struct {
	RoomName   string
	OccurredAt time.Time
}

// Kind implements ParsedEvent.
func (RoomNameEvent) Kind() EventKind { return EventKindRoomName }

// AvatarSwitchEvent is emitted for Switching <user> to avatar <name>.
type AvatarSwitchEvent struct {
	DisplayName string
	AvatarName  string
	OccurredAt  time.Time
}

// Kind implements ParsedEvent.
func (AvatarSwitchEvent) Kind() EventKind { return EventKindAvatarSwitch }

// VideoPlaybackEvent is emitted when a video URL is resolved.
type VideoPlaybackEvent struct {
	URL        string
	OccurredAt time.Time
}

// Kind implements ParsedEvent.
func (VideoPlaybackEvent) Kind() EventKind { return EventKindVideoPlayback }

// LogParser parses VRChat output_log.txt lines into events.
type LogParser struct {
	encounterPatterns []encounterPattern
	sessionPatterns   []sessionPattern
	destinationRE     *regexp.Regexp
	roomNameRE        *regexp.Regexp
	avatarRE          *regexp.Regexp
	videoRE           *regexp.Regexp
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
//
// Activity-related lines use the [Behaviour] logger prefix only, so stack traces and other sources
// do not false-trigger session/destination/room parsing. The exception is [Video Playback], which
// remains its own tag.
var (
	// Encounter join/leave only from the Behaviour logger line (excludes VisitorsInformationBoard, etc.).
	encounterJoinRE = regexp.MustCompile(`(?i)\[Behaviour\]\s+OnPlayerJoined\s+(\S.*?)\s*\((usr_[a-zA-Z0-9_-]+)\)`)

	encounterLeaveRE = regexp.MustCompile(`(?i)\[Behaviour\]\s+OnPlayerLeft\s+(\S.*?)\s*\((usr_[a-zA-Z0-9_-]+)\)`)

	// Capture full instance token (may include ~private(usr_)~region(jp) etc.).
	sessionStartWrldRE = regexp.MustCompile(`(?i)\[Behaviour\]\s+Joining\s+(wrld_[^\s]+)`)

	// Local session end: OnLeftRoom / Left room / Leaving room (Behaviour line only).
	// Do not match OnPlayerLeftRoom — it appears before another user's OnPlayerLeft while still in the
	// instance; treating it as SessionEventEnd cleared world context and dropped world_id on encounters.
	// Do not match stack traces containing ".OnLeftRoom" without [Behaviour].
	sessionEndRE = regexp.MustCompile(`(?i)\[Behaviour\]\s+(?:OnLeftRoom|Left\s+room|Leaving\s+room)`)

	// Destination set: wrld_uuid:64190~private(usr_...)~region(jp)
	destinationSetRE = regexp.MustCompile(`(?i)\[Behaviour\]\s+Destination\s+set:\s*(wrld_[a-f0-9-]+):([a-zA-Z0-9]+)~([a-z]+)\(([^)]*)\)~region\(([^)]*)\)`)

	// Fallback: group instances use e.g. ~group(grp...)~groupAccessType(public)~region(use) (extra segments before ~region).
	destinationSetLooseRE = regexp.MustCompile(`(?i)\[Behaviour\]\s+Destination\s+set:\s*(wrld_[a-f0-9-]+:[^\s]+)`)

	destinationRegionFromKeyRE = regexp.MustCompile(`(?i)~region\(([^)]*)\)`)

	roomNameRE = regexp.MustCompile(`(?i)\[Behaviour\]\s+Entering\s+Room:\s*(.+)$`)

	avatarSwitchRE = regexp.MustCompile(`(?i)\[Behaviour\]\s+Switching\s+(.+?)\s+to\s+avatar\s+(.+)$`)

	videoPlaybackRE = regexp.MustCompile(`(?i)\[Video Playback\]\s+(?:Attempting to resolve URL|Resolving URL)\s+'([^']+)'`)
)

// vrchatLineTimeRE matches the leading local timestamp in output_log.txt lines.
var vrchatLineTimeRE = regexp.MustCompile(`^(\d{4}\.\d{2}\.\d{2}\s+\d{2}:\d{2}:\d{2})`)

// ParseVRChatTimestamp extracts YYYY.MM.DD HH:MM:SS at the start of a line; returns fallback if absent or invalid.
func ParseVRChatTimestamp(line string, fallback time.Time) time.Time {
	m := vrchatLineTimeRE.FindStringSubmatch(line)
	if len(m) < 2 {
		return fallback
	}
	t, err := time.ParseInLocation("2006.01.02 15:04:05", m[1], time.Local)
	if err != nil {
		return fallback
	}
	return t
}

// NewLogParser returns a LogParser with default patterns.
func NewLogParser() *LogParser {
	return &LogParser{
		encounterPatterns: []encounterPattern{
			{encounterJoinRE, EncounterActionJoin},
			{encounterLeaveRE, EncounterActionLeave},
		},
		sessionPatterns: []sessionPattern{
			{sessionStartWrldRE, SessionEventStart},
			{sessionEndRE, SessionEventEnd},
		},
		destinationRE: destinationSetRE,
		roomNameRE:    roomNameRE,
		avatarRE:      avatarSwitchRE,
		videoRE:       videoPlaybackRE,
	}
}

// ParseLine parses a single log line and returns zero or more events.
// Unparseable lines return nil, nil (no error). baseTime is used when the log line has no timestamp.
func (p *LogParser) ParseLine(line string, baseTime time.Time) ([]ParsedEvent, error) {
	line = trimLogPrefix(line)

	for _, pat := range p.encounterPatterns {
		if m := pat.re.FindStringSubmatch(line); len(m) >= 2 {
			e := p.buildEncounterEvent(m, pat.action, baseTime)
			if e != nil {
				return []ParsedEvent{e}, nil
			}
		}
	}

	for _, pat := range p.sessionPatterns {
		if m := pat.re.FindStringSubmatch(line); len(m) >= 1 {
			e := p.buildSessionEvent(m, pat.kind, baseTime)
			if e != nil {
				return []ParsedEvent{e}, nil
			}
		}
	}

	if m := p.destinationRE.FindStringSubmatch(line); len(m) >= 6 {
		owner := strings.TrimSpace(m[4])
		fullInst := m[1] + ":" + m[2] + "~" + m[3] + "(" + m[4] + ")~region(" + m[5] + ")"
		return []ParsedEvent{&DestinationSetEvent{
			WorldID:      m[1],
			InstanceID:   m[2],
			InstanceType: strings.ToLower(m[3]),
			OwnerUserID:  owner,
			Region:       m[5],
			OccurredAt:   baseTime,
			FullInstance: fullInst,
		}}, nil
	}

	if m := destinationSetLooseRE.FindStringSubmatch(line); len(m) >= 2 {
		full := strings.TrimSpace(m[1])
		if e := destinationSetEventFromFullInstanceKey(full, baseTime); e != nil {
			return []ParsedEvent{e}, nil
		}
	}

	if m := p.roomNameRE.FindStringSubmatch(line); len(m) >= 2 {
		name := trimDisplayName(m[1])
		if name != "" {
			return []ParsedEvent{&RoomNameEvent{RoomName: name, OccurredAt: baseTime}}, nil
		}
	}

	if m := p.avatarRE.FindStringSubmatch(line); len(m) >= 3 {
		if !strings.Contains(strings.ToLower(m[1]), "to network region") {
			dn := trimDisplayName(m[1])
			av := trimDisplayName(m[2])
			if dn != "" && av != "" {
				return []ParsedEvent{&AvatarSwitchEvent{DisplayName: dn, AvatarName: av, OccurredAt: baseTime}}, nil
			}
		}
	}

	if m := p.videoRE.FindStringSubmatch(line); len(m) >= 2 {
		u := strings.TrimSpace(m[1])
		if u != "" {
			return []ParsedEvent{&VideoPlaybackEvent{URL: u, OccurredAt: baseTime}}, nil
		}
	}

	return nil, nil
}

func destinationSetEventFromFullInstanceKey(full string, at time.Time) *DestinationSetEvent {
	wid := WorldIDFromInstanceKey(full)
	if wid == "" {
		return nil
	}
	region := ""
	if m := destinationRegionFromKeyRE.FindStringSubmatch(full); len(m) >= 2 {
		region = strings.TrimSpace(m[1])
	}
	return &DestinationSetEvent{
		WorldID:      wid,
		InstanceID:   instanceIDSegmentFromFullInstanceKey(full),
		InstanceType: "",
		OwnerUserID:  "",
		Region:       region,
		OccurredAt:   at,
		FullInstance: full,
	}
}

func instanceIDSegmentFromFullInstanceKey(full string) string {
	i := strings.Index(full, ":")
	if i < 0 {
		return ""
	}
	rest := full[i+1:]
	j := strings.Index(rest, "~")
	if j < 0 {
		return rest
	}
	return rest[:j]
}

var trimLeadingBracketPrefix = regexp.MustCompile(`^\[.*?\]\s*`)

func trimLogPrefix(line string) string {
	if trimLeadingBracketPrefix.MatchString(line) {
		return trimLeadingBracketPrefix.ReplaceAllString(line, "")
	}
	return line
}

func (p *LogParser) buildEncounterEvent(m []string, action string, baseTime time.Time) *EncounterEvent {
	displayName := ""
	vrcUserID := ""
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
