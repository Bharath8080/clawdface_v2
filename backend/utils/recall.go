package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	cron "github.com/robfig/cron/v3"
)

// Recall Events
type Event struct {
	Event string `json:"event"`
	Data  struct {
		Data struct {
			Words []struct {
				Text           string `json:"text"`
				StartTimestamp struct {
					Relative float64 `json:"relative"`
				} `json:"start_timestamp"`
				EndTimestamp *struct {
					Relative float64 `json:"relative"`
				} `json:"end_timestamp"` // nullable
			} `json:"words"`

			Participant struct {
				ID        int64       `json:"id"`
				Name      *string     `json:"name"`
				IsHost    bool        `json:"is_host"`
				Platform  *string     `json:"platform"`
				ExtraData interface{} `json:"extra_data"`
				Email     *string     `json:"email"`
			} `json:"participant"`
		} `json:"data"`

		RealtimeEndpoint struct {
			ID       string      `json:"id"`
			Metadata interface{} `json:"metadata"`
		} `json:"realtime_endpoint"`

		Transcript struct {
			ID       string      `json:"id"`
			Metadata interface{} `json:"metadata"`
		} `json:"transcript"`

		Recording struct {
			ID       string      `json:"id"`
			Metadata interface{} `json:"metadata"`
		} `json:"recording"`

		Bot struct {
			ID       string      `json:"id"`
			Metadata interface{} `json:"metadata"`
		} `json:"bot"`
	} `json:"data"`
}

// Bot Information
type MeetingResponse struct {
	ID               string            `json:"id"`
	MeetingURL       MeetingURL        `json:"meeting_url"`
	BotName          string            `json:"bot_name"`
	JoinAt           time.Time         `json:"join_at"`
	RecordingConfig  RecordingConfig   `json:"recording_config"`
	StatusChanges    []StatusChange    `json:"status_changes"`
	Recordings       []Recording       `json:"recordings"`
	OutputMedia      OutputMedia       `json:"output_media"`
	AutomaticLeave   AutomaticLeave    `json:"automatic_leave"`
	Variant          Variant           `json:"variant"`
	CalendarMeetings []json.RawMessage `json:"calendar_meetings"` // flexible, empty in sample
	Metadata         json.RawMessage   `json:"metadata"`
}

type MeetingURL struct {
	MeetingID string `json:"meeting_id"`
	Platform  string `json:"platform"`
}

type RecordingConfig struct {
	Transcript                                TranscriptConfig   `json:"transcript"`
	RealtimeEndpoints                         []RealtimeEndpoint `json:"realtime_endpoints"`
	Retention                                 RetentionConfig    `json:"retention"`
	VideoMixedLayout                          string             `json:"video_mixed_layout"`
	VideoMixedMP4                             json.RawMessage    `json:"video_mixed_mp4"`
	ParticipantEvents                         json.RawMessage    `json:"participant_events"`
	MeetingMetadata                           json.RawMessage    `json:"meeting_metadata"`
	VideoMixedParticipantVideoWhenScreenshare string             `json:"video_mixed_participant_video_when_screenshare"`
	StartRecordingOn                          string             `json:"start_recording_on"`
}

type TranscriptConfig struct {
	Provider ProviderWrapper `json:"provider"`
}

type ProviderWrapper struct {
	DeepgramStreaming *DeepgramStreaming `json:"deepgram_streaming,omitempty"`
	// Add other providers here as needed
}

type DeepgramStreaming struct {
	LanguageCode   string `json:"language_code"`
	Mode           string `json:"mode"`
	InterimResults *bool  `json:"interim_results,omitempty"` // present in some nested provider objects
}

type RealtimeEndpoint struct {
	Type   string   `json:"type"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

type RetentionConfig struct {
	Type string `json:"type"`
}

type StatusChange struct {
	Code      string    `json:"code"`
	Message   *string   `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	SubCode   *string   `json:"sub_code"`
}

type Recording struct {
	ID             string          `json:"id"`
	CreatedAt      time.Time       `json:"created_at"`
	StartedAt      time.Time       `json:"started_at"`
	CompletedAt    time.Time       `json:"completed_at"`
	ExpiresAt      *time.Time      `json:"expires_at"`
	Status         Status          `json:"status"`
	MediaShortcuts MediaShortcuts  `json:"media_shortcuts"`
	Metadata       json.RawMessage `json:"metadata"`
}

type Status struct {
	Code      string    `json:"code"`
	SubCode   *string   `json:"sub_code"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MediaShortcuts struct {
	VideoMixed        *MediaItem `json:"video_mixed,omitempty"`
	Transcript        *MediaItem `json:"transcript,omitempty"`
	ParticipantEvents *MediaItem `json:"participant_events,omitempty"`
	MeetingMetadata   *MediaItem `json:"meeting_metadata,omitempty"`
	AudioMixed        *MediaItem `json:"audio_mixed,omitempty"`
}

type MediaItem struct {
	ID        string          `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	Status    Status          `json:"status"`
	Metadata  json.RawMessage `json:"metadata"`
	Data      MediaData       `json:"data"`
	Format    string          `json:"format,omitempty"` // e.g. "mp4"
}

type MediaData struct {
	DownloadURL                  string `json:"download_url,omitempty"`
	ProviderDataDownloadURL      string `json:"provider_data_download_url,omitempty"`
	ParticipantEventsDownloadURL string `json:"participant_events_download_url,omitempty"`
	SpeakerTimelineDownloadURL   string `json:"speaker_timeline_download_url,omitempty"`
	ParticipantsDownloadURL      string `json:"participants_download_url,omitempty"`
	// add any other named URLs/fields you need from "data"
}

type OutputMedia struct {
	Camera CameraOutput `json:"camera"`
}

type CameraOutput struct {
	Kind   string       `json:"kind"`
	Config CameraConfig `json:"config"`
}

type CameraConfig struct {
	URL string `json:"url"`
}

type AutomaticLeave struct {
	WaitingRoomTimeout               int                    `json:"waiting_room_timeout"`
	NooneJoinedTimeout               int                    `json:"noone_joined_timeout"`
	EveryoneLeftTimeout              EveryoneLeftTimeout    `json:"everyone_left_timeout"`
	InCallNotRecordingTimeout        int                    `json:"in_call_not_recording_timeout"`
	RecordingPermissionDeniedTimeout int                    `json:"recording_permission_denied_timeout"`
	SilenceDetection                 SilenceDetectionConfig `json:"silence_detection"`
	BotDetection                     BotDetectionConfig     `json:"bot_detection"`
}

type EveryoneLeftTimeout struct {
	Timeout       int  `json:"timeout"`
	ActivateAfter *int `json:"activate_after"` // null allowed
}

type SilenceDetectionConfig struct {
	Timeout       int `json:"timeout"`
	ActivateAfter int `json:"activate_after"`
}

type BotDetectionConfig struct {
	UsingParticipantEvents BotDetectionUsingParticipantEvents `json:"using_participant_events"`
}

type BotDetectionUsingParticipantEvents struct {
	Timeout       int `json:"timeout"`
	ActivateAfter int `json:"activate_after"`
}

type Variant struct {
	Zoom           string `json:"zoom"`
	GoogleMeet     string `json:"google_meet"`
	MicrosoftTeams string `json:"microsoft_teams"`
	Webex          string `json:"webex"`
}

// RecallParticipant represents a participant returned by the Recall.AI API.
type RecallParticipant struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	IsHost bool   `json:"is_host"`
}

// Avatar WS Connection Events
type AvatarEvent struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

// AgentMail Events
type AgentMailEvent struct {
	BodyIncluded bool   `json:"body_included"`
	EventID      string `json:"event_id"`
	EventType    string `json:"event_type"`
	Message      struct {
		CreatedAt      string   `json:"created_at"`
		From           string   `json:"from"`
		From_          string   `json:"from_"`
		HTML           string   `json:"html"`
		InboxID        string   `json:"inbox_id"`
		Labels         []string `json:"labels"`
		MessageID      string   `json:"message_id"`
		OrganizationID string   `json:"organization_id"`
		PodID          string   `json:"pod_id"`
		Preview        string   `json:"preview"`
		Size           int      `json:"size"`
		SMTPID         string   `json:"smtp_id"`
		Subject        string   `json:"subject"`
		Text           string   `json:"text"`
		ThreadID       string   `json:"thread_id"`
		Timestamp      string   `json:"timestamp"`
		To             []string `json:"to"`
		Attachments    []struct {
			AttachmentID string `json:"attachment_id"`
			Size         int    `json:"size"`
			Inline       bool   `json:"inline"`
			Filename     string `json:"filename"`
			ContentType  string `json:"content_type"`
		} `json:"attachments"`
		UpdatedAt string `json:"updated_at"`
	} `json:"message"`
	Type string `json:"type"`
}

// Inbox Details
type Inbox struct {
	OrganizationID string `json:"organization_id"`
	PodID          string `json:"pod_id"`
	InboxId        string `json:"inbox_id"`
	DisplayName    string `json:"display_name"`
	UpdatedAt      string `json:"updated_at"`
	CreatedAt      string `json:"created_at"`
}

// OpenAI Compatible Objects
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type ResponseFormat struct {
	Type string `json:"type"`
}
type GenerationRequest struct {
	Messages            []LLMMessage   `json:"messages"`
	Model               string         `json:"model"`
	Temperature         uint8          `json:"temperature"`
	MaxCompletionTokens uint16         `json:"max_completion_tokens"`
	TopP                uint8          `json:"top_p"`
	Stream              bool           `json:"stream"`
	ResponseFormat      ResponseFormat `json:"response_format"`
}

// Meeting Info Type
type MeetingInfo struct {
	URL            string `json:"meeting_url"`
	Platform       string `json:"meeting_platform"`
	From           string `json:"from"`
	Cron           string `json:"cron"`
	ExpiryDateTime string `json:"expiry_date_time"`
}

type ctxKey string

const traceKey ctxKey = "traceID"

var JobTraceMap = make(map[string]string)

// Method to extract meeting info from email body
const systemPrompt = `
You are an expert calendar extraction engine.

Your task is to extract or infer meeting events from raw email content.

The input may be:
- a raw .ics calendar file
- a plain-text meeting email
- a forwarded email chain
- a reply
- mixed or noisy content

The input is the sole source of factual data.
You must NOT invent facts that are not supported by the input.

IMPORTANT DISTINCTION:
- You ARE allowed to apply rule-based inference.
- You are NOT allowed to freely guess or hallucinate.

Think in two phases:
1. Evidence extraction (times, people, links, wording)
2. Event construction using the inference rules below

Your job:
- Determine whether the content represents:
  - a meeting creation (scheduled for the future, or happening right now)
  - a meeting update
  - a meeting cancellation
  - or no meeting

- Priority order for event_type determination:
- 1. If STATUS:CANCELLED is present → always "cancel" regardless of SEQUENCE
- 2 .If ATTENDEE-ONLY CANCEL NORMALIZATION RULE applies → "cancel"
- 3. If SEQUENCE > 0 AND STATUS is not CANCELLED → "update"
- 4. If METHOD:CANCEL in the VCALENDAR header → "cancel"
- 5. Otherwise → "create"

- If ALL of the following conditions are true:
- METHOD:CANCEL is present
- STATUS is NOT CANCELLED (e.g., CONFIRMED or missing)
- SEQUENCE = 0 OR SEQUENCE has not increased meaningfully
- The agent’s email appears in ATTENDEE AND the agent has PARTSTAT=DECLINED

→ Then interpret this as:
“The agent has been removed or declined from the event, NOT a full event cancellation.”
FORWARDED MEETING RULE:
- If the email headers contain "x-ms-exchange-meetingforward-message: Forward"
  OR the SUMMARY starts with "FW:" or "FWD or the SUBJECT starts with "FW:" :"
  AND METHOD is REQUEST (not CANCEL)
  → treat as event_type "create" regardless of SEQUENCE value.
  This is a new invite being forwarded to a new recipient, not an update.

- If the email content indicates that the user or the agent has been removed from the guest list (e.g., "You have been removed from this event" or similar) → event_type MUST be "cancel".

- Strip "FW:", "FWD:", "RE:" prefixes from SUMMARY when setting the title field.
- Do NOT extract notes from VALARM or DESCRIPTION:REMINDER blocks.
  Notes should come from the main DESCRIPTION field of the VEVENT only.
CRITICAL OVERRIDE — takes priority over all other RRULE rules:
- If event_type is "create" due to the FORWARDED MEETING RULE
  AND RECURRENCE-ID is present in the ICS:
  → "rrule" MUST be null. No exceptions.
  → Do NOT copy any RRULE from anywhere in the ICS, including 
    from VTIMEZONE, STANDARD, or DAYLIGHT blocks.
  → VTIMEZONE rules are timezone DST definitions, NOT event recurrence.

MICROSOFT TIMEZONE NAMES:
- TZID=Eastern Standard Time  → UTC-5:00 (EST)
- TZID=Eastern Daylight Time  → UTC-4:00 (EDT)
- TZID=Pacific Standard Time  → UTC-8:00
- TZID=Pacific Daylight Time  → UTC-7:00
- TZID=India Standard Time    → UTC+5:30

- RECURRENCE-ID handling (critical for recurring meetings):
- If the ICS contains a RECURRENCE-ID field, it means only ONE specific occurrence of a recurring series is being modified or cancelled — NOT the entire series.
- When a RECURRENCE-ID is present, copy its value verbatim into the "recurrence_id" field of the output JSON.
- When recurrence_id is set and event_type is "cancel", the cancellation applies ONLY to that single occurrence; the rest of the recurring series must NOT be cancelled.
- When there is no RECURRENCE-ID field, set "recurrence_id" to null.
- IMPORTANT: If the email subject or body contains phrases like "happening now", "inviting you to join a video call", "join now", or similar — treat it as event_type "create" with start_time_utc set to null. The agent should join immediately.

- Resolve the FINAL intended state of the meeting.
- Prefer the most recent information if conflicts exist.
- Ignore signatures, disclaimers, and quoted history unless relevant.
- Leave the "cron" field as null always. The schedule is derived from start_time_utc. If the ICS contains an RRULE field, copy it verbatim into the "rrule" field (e.g. "FREQ=DAILY;BYHOUR=20;BYMINUTE=0"). If there is no recurrence, set "rrule" to null.
- Notes field should be a short summary of the contents of the mail and things present in "DESCRIPTION" and can contain important things such as passwords required to join .
- Give only the json expected and nothing else apart from it , no justification , no explaining how you got there.

- If multiple ICS attachments are present with the same UID, treat them as the same event and ignore duplicates.
- If multiple ICS attachments are present with different UID  , display them in different sets.

TIMEZONE CONVERSION RULES — CRITICAL, READ CAREFULLY:
You must convert all local times to UTC before writing start_time_utc and end_time_utc.
The output format must be RFC3339 UTC, ending in "Z" (e.g. "2026-02-27T10:48:00Z").

Common timezone offsets — memorise these exactly:
- TZID=Asia/Kolkata or IST  → UTC+5:30  (subtract 5 hours AND 30 minutes from local time)
- TZID=America/New_York EST → UTC-5:00  (add 5 hours)
- TZID=America/New_York EDT → UTC-4:00  (add 4 hours)
- TZID=America/Los_Angeles PST → UTC-8:00
- TZID=America/Los_Angeles PDT → UTC-7:00
- TZID=Europe/London GMT   → UTC+0:00
- TZID=Europe/London BST   → UTC+1:00
- TZID=Europe/Paris CET    → UTC+1:00
- TZID=Europe/Paris CEST   → UTC+2:00

WORKED EXAMPLE (IST, the most common error):
  Input:  DTSTART;TZID=Asia/Kolkata:20260227T161800
  Local time: 16:18 IST
  IST offset: +5:30 (five hours AND thirty minutes)
  UTC = 16:18 - 5:30 = 10:48
  Output: "start_time_utc": "2026-02-27T10:48:00Z"   ← CORRECT
  WRONG:  "start_time_utc": "2026-02-27T11:18:00Z"   ← this subtracts only 5h, missing the :30

If the DTSTART already ends in Z (e.g. DTSTART:20260227T104800Z), it is already UTC — use it as-is.

INFERENCE RULE FOR PLAIN TEXT EMAILS:

If ALL of the following are present:
- a clear start time
- a clear end time
- a sender email
- at least one recipient email

Then treat the content as a meeting creation,
EVEN IF:
- there is no meeting link yet
- there is no UID
- the word "meeting" is not explicitly used
- If the email is from a user and has a meeting link, then it is a meeting creation request.
- If the email has a clear start time along with a meeting link but not clear end time , then it is a meeting creation request and end time should be taken as 1 hour after the mentioned start time .

This represents an implicit meeting proposal.
`
const userPromptTemplate = `
Extract meeting information from the following content.

Return a SINGLE JSON object matching this schema exactly:

{
  "event_type": "create | update | cancel | none",
  "uid": "string | null",
  "recurrence_id": "string | null",
  "title": "string | null",
  "start_time_utc": "RFC3339 | null",
  "end_time_utc": "RFC3339 | null",
  "rrule": "string | null",
  "meeting_provider": "google_meet | zoom | teams | unknown | null",
  "meeting_link": "string | null",
  "organizer_email": "string | null",
  "notes": "string | null"
}

IMPORTANT: Do NOT include an "attendees" field — omit it completely to save space.

Content:
<<<
{{CONTENT}}
>>>
`

type WebhookPayload struct {
	EventType string  `json:"event_type"`
	EventID   string  `json:"event_id"`
	Message   Message `json:"message"`
}

type Message struct {
	From        []string     `json:"from_"`
	To          []string     `json:"to"`
	CC          []string     `json:"cc"`
	BCC         []string     `json:"bcc"`
	Subject     string       `json:"subject"`
	Preview     string       `json:"preview"`
	Text        string       `json:"text"`
	HTML        string       `json:"html"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Filename      string `json:"filename"`
	ContentBase64 string `json:"content_base64"`
	URL           string `json:"url"`
}

func normalize(input string) string {
	s := strings.ReplaceAll(input, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	var b strings.Builder
	for _, r := range s {
		if r == '\n' || r == '\t' || r >= 32 {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func htmlToText(html string) string {
	var b strings.Builder
	inTag := false
	for _, r := range html {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			b.WriteRune(r)
		}
	}
	out := b.String()
	out = strings.ReplaceAll(out, "&nbsp;", " ")
	out = strings.ReplaceAll(out, "&amp;", "&")
	return out
}

/* MIME NORMALIZER*/

type ExtractedParts struct {
	Calendar string
	Plain    string
	HTML     string
}

// parses the DTSTART line directly from the ICS/email content and converts it to UTC using known timezone offsets
func extractUTCFromICS(content string, field ...string) (string, bool) {
	targetField := "DTSTART"
	if len(field) > 0 && field[0] != "" {
		targetField = field[0]
	}
	// timezone offset map
	content = unfoldICS(content)
	tzOffsets := map[string]int{
		"Asia/Kolkata":                   330,
		"Asia/Calcutta":                  330,
		"IST":                            330,
		"India Standard Time":            330, //  Outlook/Windows timezone name
		"Eastern Standard Time":          -300,
		"Eastern Daylight Time":          -240,
		"Pacific Standard Time":          -480,
		"Pacific Daylight Time":          -420,
		"GMT Standard Time":              0,
		"Central European Standard Time": 60,
		"W. Europe Standard Time":        60,
		"Tokyo Standard Time":            540,
		"China Standard Time":            480,
		"Singapore Standard Time":        480,
		"AUS Eastern Standard Time":      600,
		"America/New_York":               -300,
		"America/Chicago":                -360,
		"America/Denver":                 -420,
		"America/Los_Angeles":            -480,
		"America/Phoenix":                -420,
		"Europe/London":                  0,
		"Europe/Paris":                   60,
		"Europe/Berlin":                  60,
		"Asia/Tokyo":                     540,
		"Asia/Shanghai":                  480,
		"Asia/Singapore":                 480,
		"Australia/Sydney":               600,
		"Pacific/Auckland":               720,
	}

	// Match: DTSTART;TZID=... or DTEND;TZID=...
	re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(targetField) + `;TZID=([^:\r\n]+):(\d{8}T\d{6})`)
	m := re.FindStringSubmatch(content)
	if m == nil {
		reUtc := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(targetField) + `:(\d{8}T\d{6}Z)`)
		mu := reUtc.FindStringSubmatch(content)
		if mu != nil {
			t, err := time.Parse("20060102T150405Z", mu[1])
			if err == nil {
				return t.UTC().Format(time.RFC3339), true
			}
		}
		return "", false
	}

	tzName := strings.TrimSpace(m[1])
	localStr := m[2]

	localTime, err := time.Parse("20060102T150405", localStr)
	if err != nil {
		log.Printf("[extractUTCFromICS] Could not parse local time %q for field %q: %v", localStr, targetField, err)
		return "", false
	}
	// Windows timezone name → IANA equivalent
	// IANA names get DST-correct resolution via time.LoadLocation
	var windowsToIANA = map[string]string{
		"Eastern Standard Time":          "America/New_York",
		"Eastern Daylight Time":          "America/New_York",
		"Pacific Standard Time":          "America/Los_Angeles",
		"Pacific Daylight Time":          "America/Los_Angeles",
		"Central Standard Time":          "America/Chicago",
		"Mountain Standard Time":         "America/Denver",
		"India Standard Time":            "Asia/Kolkata",
		"GMT Standard Time":              "Europe/London",
		"Central European Standard Time": "Europe/Paris",
		"W. Europe Standard Time":        "Europe/Berlin",
		"Tokyo Standard Time":            "Asia/Tokyo",
		"China Standard Time":            "Asia/Shanghai",
		"Singapore Standard Time":        "Asia/Singapore",
		"AUS Eastern Standard Time":      "Australia/Sydney",
	}

	// First try direct IANA load
	if loc, err := time.LoadLocation(tzName); err == nil {
		t := time.Date(localTime.Year(), localTime.Month(), localTime.Day(),
			localTime.Hour(), localTime.Minute(), localTime.Second(), 0, loc)
		return t.UTC().Format(time.RFC3339), true
	}

	// Try Windows→IANA mapping before static offset
	if ianaName, ok := windowsToIANA[tzName]; ok {
		if loc, err := time.LoadLocation(ianaName); err == nil {
			t := time.Date(localTime.Year(), localTime.Month(), localTime.Day(),
				localTime.Hour(), localTime.Minute(), localTime.Second(), 0, loc)
			log.Printf("[extractUTCFromICS] Resolved Windows TZ %q → IANA %q", tzName, ianaName)
			return t.UTC().Format(time.RFC3339), true
		}
	}
	// Fall back to static offset map for Windows/Outlook timezone names
	// that don't match IANA identifiers (e.g. "Eastern Standard Time").
	offsetMins, ok := tzOffsets[tzName]
	if !ok {
		log.Printf("[extractUTCFromICS] Unknown TZID=%q — cannot correct", tzName)
		return "", false
	}
	utc := localTime.Add(-time.Duration(offsetMins) * time.Minute)
	return utc.UTC().Format(time.RFC3339), true
}

func reduceICSForLLM(ics string) string {
	ics = unfoldICS(ics)
	lines := strings.Split(ics, "\n")

	allowedPrefixes := []string{
		"UID:",
		"METHOD:",
		"SEQUENCE:",
		"SUMMARY",
		"DTSTART",
		"DTEND",
		"RRULE:",
		"RECURRENCE-ID",
		"STATUS:",
		"ORGANIZER",
		"ATTENDEE",
		"LOCATION",
		"DESCRIPTION",
		"X-GOOGLE-CONFERENCE",
		"X-MICROSOFT-SKYPETEAMSMEETINGURL",
	}

	var kept []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		for _, p := range allowedPrefixes {
			if strings.HasPrefix(line, p) {
				// Truncate noisy fields to prevent LLM truncation
				if strings.HasPrefix(line, "DESCRIPTION") || strings.HasPrefix(line, "SUMMARY") {
					if len(line) > 500 {
						line = line[:500] + " (truncated...)"
					}
				}
				kept = append(kept, line)
				break
			}
		}
	}

	return strings.Join([]string{
		"BEGIN:VEVENT",
		strings.Join(kept, "\n"),
		"END:VEVENT",
	}, "\n")
}

func BuildLLMPayload(raw []byte) (string, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return simpleClean(string(raw)), nil
	}

	subject := msg.Header.Get("Subject")
	from := msg.Header.Get("From")
	to := msg.Header.Get("To")
	date := msg.Header.Get("Date")

	var headerBlock string
	if subject != "" || from != "" {
		headerBlock = fmt.Sprintf("Subject: %s\nFrom: %s\nTo: %s\nDate: %s\n\n", subject, from, to, date)
	}

	parts := &ExtractedParts{}
	if err := walkEntity(msg.Header, msg.Body, parts); err != nil {
		return "", err
	}
	if parts.Calendar == "" {
		log.Printf("[BuildLLMPayload] WARNING: No ICS attachment found — using plain text fallback.")
	} else {
		log.Printf("[BuildLLMPayload] ICS found — using calendar content (%d chars)", len(parts.Calendar))
	}
	if parts.Calendar == "" && parts.Plain == "" && parts.HTML == "" {
		log.Printf("[BuildLLMPayload] WARNING: All parts empty — possible nested email not unwrapped")
		return "", fmt.Errorf("no usable content")
	}
	if parts.Calendar != "" {
		log.Printf("[BuildLLMPayload] Path: ICS → LLM")
		reduced := reduceICSForLLM(parts.Calendar)
		return simpleClean(headerBlock + reduced), nil
	}
	if parts.Plain != "" {
		log.Printf("[BuildLLMPayload] Path: plain text → LLM (%d chars)", len(parts.Plain))
		return simpleClean(headerBlock + normalizeMeetingLinks(parts.Plain)), nil
	}
	if parts.HTML != "" {
		log.Printf("[BuildLLMPayload] Path: HTML stripped → LLM (%d chars)", len(parts.HTML))
		return simpleClean(headerBlock + normalizeMeetingLinks(parts.HTML)), nil
	}

	return "", fmt.Errorf("no usable content")
}

// normalizeMeetingLinks for meeting links without http
func normalizeMeetingLinks(s string) string {
	re := regexp.MustCompile(`(?i)\b(meet\.google\.com|zoom\.us/j|teams\.microsoft\.com/(l/meetup-join|meet)|webex\.com/meet)/([A-Za-z0-9/_\-?=&.%]+)`)
	return re.ReplaceAllStringFunc(s, func(m string) string {
		if !strings.HasPrefix(strings.ToLower(m), "http") {
			return "https://" + m
		}
		return m
	})
}

// Func for unfolding long lines for Outlook , exchange
func unfoldICS(s string) string {
	s = strings.ReplaceAll(s, "\r\n ", "")
	s = strings.ReplaceAll(s, "\r\n\t", "")
	s = strings.ReplaceAll(s, "\n ", "")
	s = strings.ReplaceAll(s, "\n\t", "")
	return s
}

func walkEntity(header mail.Header, body io.Reader, parts *ExtractedParts) error {
	mediaType, params, _ := mime.ParseMediaType(header.Get("Content-Type"))

	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if err := walkEntity(mail.Header(part.Header), part, parts); err != nil {
				return err
			}
		}
		return nil
	}

	data, _ := io.ReadAll(body)
	decoded := decodeIfNeeded(data, header.Get("Content-Transfer-Encoding"))

	// Handle nested email attachments
	contentDisp := header.Get("Content-Disposition")
	isAttachment := strings.Contains(contentDisp, "attachment")

	if (mediaType == "application/octet-stream" || mediaType == "message/rfc822") && isAttachment && parts.Calendar == "" {
		// Try to parse as a nested email
		nestedMsg, err := mail.ReadMessage(strings.NewReader(decoded))
		if err == nil {
			nestedParts := &ExtractedParts{}
			if walkErr := walkEntity(nestedMsg.Header, nestedMsg.Body, nestedParts); walkErr == nil {
				if nestedParts.Calendar != "" {
					log.Printf("[walkEntity] Found ICS inside nested email attachment — extracting")
					parts.Calendar = nestedParts.Calendar
					return nil
				}
				// Also grab plain/HTML if we have nothing yet
				if nestedParts.Plain != "" && parts.Plain == "" {
					parts.Plain = nestedParts.Plain
				}
				if nestedParts.HTML != "" && parts.HTML == "" {
					parts.HTML = nestedParts.HTML
				}
			}
		}
	}

	if strings.Contains(mediaType, "text/calendar") || strings.Contains(mediaType, "application/ics") || strings.Contains(header.Get("Content-Disposition"), ".ics") {
		if parts.Calendar == "" {
			parts.Calendar = unfoldICS(decoded)
		}
	}
	if strings.Contains(mediaType, "text/plain") && parts.Plain == "" {
		parts.Plain = decoded
	}
	if strings.Contains(mediaType, "text/html") && parts.HTML == "" {
		parts.HTML = stripHTML(decoded)
	}
	return nil
}

func decodeIfNeeded(data []byte, encoding string) string {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "base64":
		cleaned := strings.ReplaceAll(string(data), "\n", "")
		cleaned = strings.ReplaceAll(cleaned, "\r", "")
		decoded, err := base64.StdEncoding.DecodeString(cleaned)
		if err == nil {
			return string(decoded)
		}
	case "quoted-printable":
		return decodeQuotedPrintable(string(data))
	}
	return string(data)
}

func stripHTML(s string) string {
	s = decodeQuotedPrintable(s)
	hrefRe := regexp.MustCompile(`(?i)href=["']([^"']+)["']`)
	var links []string
	for _, m := range hrefRe.FindAllStringSubmatch(s, -1) {
		href := m[1]
		if strings.Contains(href, "meet.google.com") ||
			strings.Contains(href, "zoom.us") ||
			strings.Contains(href, "teams.microsoft.com") ||
			strings.Contains(href, "webex.com") || strings.Contains(href, "teams.microsoft.com/meet") {
			if !strings.Contains(href, "zoom.us") {
				if idx := strings.Index(href, "?"); idx != -1 {
					href = href[:idx]
				}

			}
			links = append(links, href)
		}
	}
	// strip all tags
	tagRe := regexp.MustCompile(`(?is)<[^>]+>`)
	text := tagRe.ReplaceAllString(s, "")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	for _, link := range links {
		if !strings.Contains(text, link) {
			text += "\n" + link
		}
	}
	return text
}
func decodeQuotedPrintable(s string) string {
	s = strings.ReplaceAll(s, "=\n", "")
	s = strings.ReplaceAll(s, "=\n", "")
	// Decode common =XX sequences
	qpRe := regexp.MustCompile(`=[0-9A-Fa-f]{2}`)
	s = qpRe.ReplaceAllStringFunc(s, func(match string) string {
		var b byte
		fmt.Sscanf(match[1:], "%02X", &b)
		return string([]byte{b})
	})
	return s
}

func unwrapGoogleRedirect(href string) string {
	if strings.Contains(href, "google.com/url") {
		u, err := url.Parse(href)
		if err == nil {
			q := u.Query().Get("q")
			if q != "" {
				return q
			}
		}
	}
	return href
}
func simpleClean(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.TrimSpace(s)
}

// extractRecipientEmail
func extractRecipientEmail(raw []byte) string {
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return ""
	}
	for _, headerName := range []string{"To", "Delivered-To", "X-Original-To", "X-Forwarded-To"} {
		val := msg.Header.Get(headerName)
		if val == "" {
			continue
		}
		addr, err := mail.ParseAddress(val)
		if err == nil && addr.Address != "" {
			// Extract from original val to preserve case
			if idx := strings.Index(val, "<"); idx != -1 {
				if end := strings.Index(val, ">"); end > idx {
					return strings.TrimSpace(val[idx+1 : end])
				}
			}
			return strings.TrimSpace(val)
		}
		// Manual angle bracket fallback for Outlook quoted format
		if idx := strings.Index(val, "<"); idx != -1 {
			if end := strings.Index(val, ">"); end > idx {
				return strings.TrimSpace(val[idx+1 : end])
			}
		}
	}
	re := regexp.MustCompile(`(?i)for\s+<([^>]+)>`)
	m := re.FindSubmatch(raw[:min(len(raw), 4000)])
	if m != nil {
		return string(m[1])
	}
	return ""
}

var LLM_URL = "https://truegn-agents-resource.services.ai.azure.com/api/projects/truegn-agents/openai/v1/responses" //AzureLLM Endpoint

func callLLM(content string) (string, error) {
	log.Printf("[callLLM] Starting LLM call — content length=%d chars", len(content))

	apiKey := os.Getenv("LLM_API_TOKEN")
	if apiKey == "" {
		return "", fmt.Errorf("LLM_API_KEY env var not set")
	}

	userPrompt := strings.Replace(userPromptTemplate, "{{CONTENT}}", content, 1)

	model := "gpt-5.4-mini" //changed model from gpt oss to gpt 5.4 mini

	payload := map[string]interface{}{
		"model":             model,
		"instructions":      systemPrompt,
		"input":             userPrompt,
		"temperature":       0.0,
		"max_output_tokens": 1500,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(
		"POST",
		LLM_URL,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	log.Printf("[callLLM] Response status: %d", resp.StatusCode)

	if resp.StatusCode != 200 {
		log.Printf("[callLLM] ERROR: non-200 status — body: %s", string(respBody))
		return "", fmt.Errorf("Azure AI foundry error  %d: %s", resp.StatusCode, respBody)
	}

	var parsed struct {
		Output []struct {
			Type    string `json:"type"`
			Role    string `json:"role"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}

	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}

	for _, item := range parsed.Output {
		if item.Type == "message" && item.Role == "assistant" {
			if len(item.Content) == 0 {
				return "", fmt.Errorf("assistant message has no content")
			}
			result := strings.TrimSpace(item.Content[0].Text)
			if result == "" {
				return "", fmt.Errorf("content empty — model returned no JSON output")
			}
			log.Printf("[callAzureLLM] SUCCESS — response length=%d chars", len(result))
			return result, nil
		}
	}
	return "", fmt.Errorf("no assistant message found in Azure AI Foundry response")

}

// Get to get participants info
func GetAllRecallParticipants(botId string) ([]RecallParticipant, error) {
	// Get Bot Information
	url := "https://us-west-2.recall.ai/api/v1/bot/" + botId
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return []RecallParticipant{}, err
	}
	req.Header.Add("Authorization", os.Getenv("RECALL_API_TOKEN"))
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return []RecallParticipant{}, err
	}
	defer res.Body.Close()
	// Parse Bot Information and Get Participants details
	var meetingResponse MeetingResponse
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return []RecallParticipant{}, err
	}
	err = json.Unmarshal(body, &meetingResponse)
	if err != nil {
		fmt.Println(err)
		return []RecallParticipant{}, err
	}
	if len(meetingResponse.Recordings) > 0 {
		// Get Participant Info
		url := meetingResponse.Recordings[0].MediaShortcuts.ParticipantEvents.Data.ParticipantsDownloadURL
		method := "GET"
		client := &http.Client{}
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			fmt.Println(err)
			return []RecallParticipant{}, err
		}
		req.Header.Add("Authorization", os.Getenv("RECALL_API_TOKEN"))
		req.Header.Add("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return []RecallParticipant{}, err
		}
		defer res.Body.Close()
		// Parse Bot Information and Get Participants details
		var participants []RecallParticipant
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return []RecallParticipant{}, err
		}
		err = json.Unmarshal(body, &participants)
		if err != nil {
			fmt.Println(err)
			return []RecallParticipant{}, err
		}
		return participants, nil

	}
	return []RecallParticipant{}, nil
}

// PostAvatarRequest resolves the agent by inbox email and creates a conversation directly.
func PostAvatarRequest(inboxID string, lkRoomID string, meetingUrl string, from string) (string, string, bool, error) {
	log.Printf("[PostAvatarRequest] Starting — inboxID=%q lkRoomID=%q meetingUrl=%q from=%q", inboxID, lkRoomID, meetingUrl, from)

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://app.clawdface.ai"
		log.Printf("[PostAvatarRequest] FRONTEND_URL not set, defaulting to: %s", frontendURL)
	}
	frontendURL = strings.TrimRight(frontendURL, "/")

	apiURL := frontendURL + "/api/start-agent"

	payload := map[string]interface{}{
		"email":             inboxID,
		"meetingUrl":        meetingUrl,
		"roomId":            lkRoomID,
		"skipRecallTrigger": false, // let frontend handle bot creation before dispatching the agent
	}

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", "", false, fmt.Errorf("failed to create frontend request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", false, fmt.Errorf("frontend request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", "", false, fmt.Errorf("frontend API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		VideoUrl  string `json:"videoUrl"`
		AgentName string `json:"agentName"`
		RoomId    string `json:"roomId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", false, fmt.Errorf("failed to parse frontend response: %v", err)
	}

	log.Printf("[PostAvatarRequest] SUCCESS (via Frontend API) — botName=%q videoUrl=%q", result.AgentName, result.VideoUrl)
	return result.AgentName, result.VideoUrl, true, nil
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  65536,
	WriteBufferSize: 65536,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	agentWriteTimeout  = 30 * time.Second
	queueCap           = 256
	recallReadDeadline = 75 * time.Second
	recallPingInterval = 30 * time.Second
	roomIdleTTL        = 5 * time.Minute
	janitorEvery       = 1 * time.Minute
)

type queuedEvent struct {
	event string
	raw   []byte
}

func dropRank(event string) int {
	switch event {
	case "video_separate_png.data":
		return 0
	case "transcript.data", "transcript.partial_data":
		return 2
	default:
		return 1
	}
}

type roomState struct {
	roomID string

	mu            sync.Mutex
	cond          *sync.Cond
	queue         []queuedEvent
	agentConn     *websocket.Conn
	writerStarted bool
	closed        bool
	recallActive  int
	lastActivity  time.Time

	droppedVideo   uint64
	droppedControl uint64
	lastDropLog    time.Time

	agentMu sync.Mutex
}

var (
	roomsMu     sync.Mutex
	rooms       = make(map[string]*roomState)
	janitorOnce sync.Once
)

func roomFor(roomID string) *roomState {
	janitorOnce.Do(func() { go roomJanitor() })

	roomsMu.Lock()
	defer roomsMu.Unlock()
	rs, ok := rooms[roomID]
	if !ok {
		rs = &roomState{roomID: roomID, lastActivity: time.Now()}
		rs.cond = sync.NewCond(&rs.mu)
		rooms[roomID] = rs
	} else {
		rs.mu.Lock()
		rs.lastActivity = time.Now()
		rs.mu.Unlock()
	}
	return rs
}

func roomJanitor() {
	t := time.NewTicker(janitorEvery)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		roomsMu.Lock()
		for id, rs := range rooms {
			rs.mu.Lock()
			idle := rs.recallActive == 0 &&
				rs.agentConn == nil &&
				now.Sub(rs.lastActivity) > roomIdleTTL
			if idle {
				droppedByRank := [3]int{}
				for _, e := range rs.queue {
					droppedByRank[dropRank(e.event)]++
				}
				queuedLen := len(rs.queue)
				rs.queue = nil
				rs.closed = true
				rs.mu.Unlock()
				rs.cond.Broadcast()
				delete(rooms, id)
				log.Printf("[Relay] reaped idle room - roomID=%s queued_dropped=%d (video=%d other=%d transcript=%d)",
					id, queuedLen, droppedByRank[0], droppedByRank[1], droppedByRank[2])
			} else {
				rs.mu.Unlock()
			}
		}
		roomsMu.Unlock()
	}
}

func (rs *roomState) evictOneLocked() bool {
	for i, e := range rs.queue {
		if dropRank(e.event) == 0 {
			rs.queue = append(rs.queue[:i], rs.queue[i+1:]...)
			rs.droppedVideo++
			return true
		}
	}
	for i, e := range rs.queue {
		if dropRank(e.event) == 1 {
			rs.queue = append(rs.queue[:i], rs.queue[i+1:]...)
			rs.droppedControl++
			return true
		}
	}
	return false
}

func (rs *roomState) enqueue(raw []byte, event string) {
	buf := make([]byte, len(raw))
	copy(buf, raw)
	rank := dropRank(event)

	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.lastActivity = time.Now()

	if len(rs.queue) >= queueCap {
		if !rs.evictOneLocked() {
			if rank < 2 {
				if rank == 0 {
					rs.droppedVideo++
				} else {
					rs.droppedControl++
				}
				rs.maybeLogDropLocked()
				return
			}
			log.Printf("[Relay] WARN queue cap hit with only protected transcripts queued; accepting transcript above cap roomID=%s event=%s queue_len=%d cap=%d",
				rs.roomID, event, len(rs.queue), queueCap)
		} else {
			rs.maybeLogDropLocked()
		}
	}
	rs.queue = append(rs.queue, queuedEvent{event: event, raw: buf})
	if !rs.writerStarted && !rs.closed {
		rs.writerStarted = true
		go rs.writerLoop()
	}
	rs.cond.Signal()
}

func (rs *roomState) requeueFrontLocked(evt queuedEvent) {
	if len(rs.queue) >= queueCap {
		if !rs.evictOneLocked() {
			if dropRank(evt.event) < 2 {
				if dropRank(evt.event) == 0 {
					rs.droppedVideo++
				} else {
					rs.droppedControl++
				}
				rs.maybeLogDropLocked()
				return
			}
			log.Printf("[Relay] WARN cap hit on requeue with only protected transcripts queued; requeueing transcript above cap roomID=%s event=%s queue_len=%d cap=%d",
				rs.roomID, evt.event, len(rs.queue), queueCap)
		}
	}
	rs.queue = append([]queuedEvent{evt}, rs.queue...)
}

func (rs *roomState) attachAgent(conn *websocket.Conn) {
	rs.mu.Lock()
	rs.agentConn = conn
	rs.lastActivity = time.Now()
	if !rs.writerStarted && !rs.closed {
		rs.writerStarted = true
		go rs.writerLoop()
	}
	rs.mu.Unlock()
	rs.cond.Broadcast()
}

func (rs *roomState) detachAgent(conn *websocket.Conn) {
	rs.mu.Lock()
	if rs.agentConn == conn {
		rs.agentConn = nil
	}
	rs.lastActivity = time.Now()
	rs.mu.Unlock()
	rs.cond.Broadcast()
}

func (rs *roomState) maybeLogDropLocked() {
	if time.Since(rs.lastDropLog) > 2*time.Second {
		log.Printf("[Relay] WARN dropping events - roomID=%s video=%d other=%d queue_len=%d",
			rs.roomID, rs.droppedVideo, rs.droppedControl, len(rs.queue))
		rs.lastDropLog = time.Now()
	}
}

func (rs *roomState) writerLoop() {
	for {
		rs.mu.Lock()
		for len(rs.queue) == 0 || rs.agentConn == nil {
			if rs.closed {
				rs.mu.Unlock()
				return
			}
			rs.cond.Wait()
		}
		evt := rs.queue[0]
		rs.queue = rs.queue[1:]
		conn := rs.agentConn
		rs.mu.Unlock()

		rs.agentMu.Lock()
		_ = conn.SetWriteDeadline(time.Now().Add(agentWriteTimeout))
		err := conn.WriteMessage(websocket.TextMessage, evt.raw)
		rs.agentMu.Unlock()

		if err != nil {
			log.Printf("[Relay] write to agent failed, requeueing - roomID=%s event=%s err=%v",
				rs.roomID, evt.event, err)
			rs.mu.Lock()
			rs.requeueFrontLocked(evt)
			if rs.agentConn == conn {
				rs.agentConn = nil
			}
			rs.mu.Unlock()
			_ = conn.Close()
		}
	}
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Println("[WS] HIT ENDPOINT")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("[WebSocket] Close frame from peer - RemoteAddr=%s code=%d reason=%q", r.RemoteAddr, code, text)
		msg := websocket.FormatCloseMessage(code, "")
		_ = conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(time.Second))
		return nil
	})

	var attached *roomState

	roomID := r.URL.Query().Get("roomID")
	if roomID != "" {
		attached = roomFor(roomID)
		attached.attachAgent(conn)
		log.Printf("[WS] Registered via query param - roomID=%s", roomID)
	}

	log.Printf("[WebSocket] New connection established - RemoteAddr=%s", r.RemoteAddr)
	defer func() {
		if attached != nil {
			attached.detachAgent(conn)
		}
		conn.Close()
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			closeCode := -1
			closeText := ""
			if ce, ok := err.(*websocket.CloseError); ok {
				closeCode = ce.Code
				closeText = ce.Text
			}
			log.Printf("[WebSocket] Connection closed - RemoteAddr=%s err=%v closeCode=%d closeText=%q unexpected=%v",
				r.RemoteAddr, err, closeCode, closeText,
				websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway))
			return
		}

		log.Printf("[WebSocket] Raw message received - length=%d bytes data=%s", len(data), string(data))
		var avatarEvent AvatarEvent
		if err := json.Unmarshal(data, &avatarEvent); err != nil {
			log.Printf("[WebSocket] Failed to unmarshal message - error=%v raw=%s", err, string(data))
			continue
		}
		log.Printf("[WebSocket] Parsed event - type=%q data=%q", avatarEvent.Type, avatarEvent.Data)
		if avatarEvent.Type == "set_lk_room_id" {
			if attached != nil && attached.roomID != avatarEvent.Data {
				attached.detachAgent(conn)
			}
			attached = roomFor(avatarEvent.Data)
			attached.attachAgent(conn)
			log.Printf("[WebSocket] Registered lkRoomID=%s", avatarEvent.Data)
		}
	}
}

func HandleRecallWS(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	log.Printf("[RecallWS] Recall connected - roomID=%s RemoteAddr=%s", roomID, r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[RecallWS] Upgrade error - roomID=%s err=%v", roomID, err)
		return
	}
	defer conn.Close()

	rs := roomFor(roomID)
	rs.mu.Lock()
	rs.recallActive++
	rs.lastActivity = time.Now()
	rs.mu.Unlock()
	defer func() {
		rs.mu.Lock()
		rs.recallActive--
		rs.lastActivity = time.Now()
		rs.mu.Unlock()
	}()

	_ = conn.SetReadDeadline(time.Now().Add(recallReadDeadline))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(recallReadDeadline))
		return nil
	})

	pingDone := make(chan struct{})
	go func() {
		t := time.NewTicker(recallPingInterval)
		defer t.Stop()
		for {
			select {
			case <-pingDone:
				return
			case <-t.C:
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
					log.Printf("[RecallWS] ping write failed - roomID=%s err=%v", roomID, err)
					return
				}
			}
		}
	}()
	defer close(pingDone)

	frameCount := 0
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[RecallWS] Connection closed - roomID=%s frames_received=%d err=%v", roomID, frameCount, err)
			return
		}
		_ = conn.SetReadDeadline(time.Now().Add(recallReadDeadline))
		frameCount++

		var peek struct {
			Event string `json:"event"`
		}
		_ = json.Unmarshal(data, &peek)

		switch peek.Event {
		case "transcript.data", "transcript.partial_data":
			var body Event
			if err := json.Unmarshal(data, &body); err == nil {
				parts := make([]string, len(body.Data.Data.Words))
				for i, w := range body.Data.Data.Words {
					parts[i] = w.Text
				}
				speaker := "unknown"
				if body.Data.Data.Participant.Name != nil {
					speaker = *body.Data.Data.Participant.Name
				}
				log.Printf("[RecallWS][Transcript] roomID=%s speaker=%s text=%s", roomID, speaker, strings.Join(parts, " "))
			}
		case "video_separate_png.data":
			frameType := "unknown"
			bufLen := 0
			var parsed map[string]interface{}
			if json.Unmarshal(data, &parsed) == nil {
				if eventData, ok := parsed["data"].(map[string]interface{}); ok {
					if inner, ok := eventData["data"].(map[string]interface{}); ok {
						if t, ok := inner["type"].(string); ok {
							frameType = t
						}
						if buf, ok := inner["buffer"].(string); ok {
							bufLen = len(buf)
						}
					}
				}
			}
			log.Printf("[RecallWS][Video] Frame #%d roomID=%s type=%s buf_len=%d", frameCount, roomID, frameType, bufLen)
		default:
			log.Printf("[RecallWS] Event roomID=%s type=%s bytes=%d", roomID, peek.Event, len(data))
		}

		rs.enqueue(data, peek.Event)
	}
}

// AWS Webhook

// types used
type LLMEvent struct {
	EventType    string   `json:"event_type"`
	UID          string   `json:"uid"`
	RecurrenceID string   `json:"recurrence_id"`
	Title        string   `json:"title"`
	StartTimeUTC string   `json:"start_time_utc"`
	EndTimeUTC   string   `json:"end_time_utc"`
	RRule        string   `json:"rrule"`
	MeetingProv  string   `json:"meeting_provider"`
	MeetingLink  string   `json:"meeting_link"`
	Organizer    string   `json:"organizer_email"`
	Attendees    []string `json:"attendees"`
	Notes        string   `json:"notes"`
	Cron         string   `json:"cron"`
}

// convertRRuleToCron converts a basic RRULE string into a 5-field cron expression.
func convertRRuleToCron(rrule string, startTime time.Time) string {
	hour := startTime.UTC().Hour()
	minute := startTime.UTC().Minute()

	parts := strings.Split(rrule, ";")
	params := make(map[string]string)
	for _, p := range parts {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) == 2 {
			params[strings.ToUpper(kv[0])] = kv[1]
		}
	}

	freq := params["FREQ"]
	switch freq {
	case "MINUTELY":
		interval := 1
		if v, ok := params["INTERVAL"]; ok {
			fmt.Sscanf(v, "%d", &interval)
		}
		if interval <= 1 {
			return "* * * * *"
		}
		return fmt.Sprintf("*/%d * * * *", interval)

	case "HOURLY":
		interval := 1
		if v, ok := params["INTERVAL"]; ok {
			fmt.Sscanf(v, "%d", &interval)
		}
		if interval <= 1 {
			return fmt.Sprintf("%d * * * *", minute)
		}
		return fmt.Sprintf("%d */%d * * *", minute, interval)

	case "DAILY":
		return fmt.Sprintf("%d %d * * *", minute, hour)

	case "WEEKLY":
		dayMap := map[string]string{
			"SU": "0", "MO": "1", "TU": "2", "WE": "3",
			"TH": "4", "FR": "5", "SA": "6",
		}
		byDay := params["BYDAY"]
		if byDay == "" {
			// Fall back to the weekday of the start time
			weekdayAbbr := strings.ToUpper(startTime.UTC().Weekday().String()[:2])
			byDay = weekdayAbbr
		}
		var cronDays []string
		for _, d := range strings.Split(byDay, ",") {
			d = strings.TrimSpace(strings.ToUpper(d))
			// Strip ordinal prefix: "1MO" → "MO", "2TH" → "TH", "-1FR" → "FR"
			if len(d) > 2 {
				d = d[len(d)-2:]
			}
			if num, ok := dayMap[d]; ok {
				cronDays = append(cronDays, num)
			}
		}
		if len(cronDays) == 0 {
			return ""
		}
		return fmt.Sprintf("%d %d * * %s", minute, hour, strings.Join(cronDays, ","))

	case "MONTHLY":
		day := startTime.UTC().Day()
		if d, ok := params["BYMONTHDAY"]; ok {
			fmt.Sscanf(d, "%d", &day)
		}
		return fmt.Sprintf("%d %d %d * *", minute, hour, day)

	default:
		return ""
	}
}

// extractRRuleUntil parses the UNTIL value from an RRULE string if present.
// Returns the parsed time and true if found and valid.
func extractRRuleUntil(rrule string) (time.Time, bool) {
	for _, part := range strings.Split(rrule, ";") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 && strings.ToUpper(kv[0]) == "UNTIL" {
			v := strings.TrimSpace(kv[1])
			// Full datetime format: 20260301T000000Z
			if t, err := time.Parse("20060102T150405Z", v); err == nil {
				return t.UTC(), true
			}
			// Date-only format: 20260301 — treat as end of that day (23:59:59 UTC)
			// so the final occurrence on that day is not skipped by the expiry check.
			if t, err := time.Parse("20060102", v); err == nil {
				endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
				return endOfDay, true
			}
		}
	}
	return time.Time{}, false
}

// helper func for duplicate job arriving in last 3 minutes
func recentlyFired(createdAt string) bool {
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return false
	}
	return time.Since(t) < 3*time.Minute
}

// Retry wrapper for LLM (GROQ)
func callLLMWithRetry(ctx context.Context, content string) (*LLMEvent, error) {
	maxAttempts := 3
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		rawResponse, err := callLLM(content)
		if err != nil {
			lastErr = err
			logWithTrace(ctx, "[LLM] Attempt %d/%d HTTP/API FAILED: %v", attempt+1, maxAttempts, err)
			continue
		}

		if rawResponse == "" {
			lastErr = fmt.Errorf("empty response from LLM")
			logWithTrace(ctx, "[LLM] Attempt %d/%d returned EMPTY", attempt+1, maxAttempts)
			continue
		}

		// Try to extract and validate JSON within the retry loop
		jsonStr := extractFirstJSONObject(rawResponse)
		if jsonStr == "" {
			lastErr = fmt.Errorf("no JSON object found in response")
			logWithTrace(ctx, "[LLM] Attempt %d/%d FAILED — No valid JSON structure found. Raw len: %d", attempt+1, maxAttempts, len(rawResponse))
			logWithTrace(ctx, "[LLM] DEBUG RAW RESPONSE ON FAILURE: %s", rawResponse)
			continue
		}

		var event LLMEvent
		if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
			lastErr = fmt.Errorf("failed to unmarshal JSON: %v", err)
			logWithTrace(ctx, "[LLM] Attempt %d/%d FAILED — JSON Corrupt/Incomplete: %v", attempt+1, maxAttempts, err)
			logWithTrace(ctx, "[LLM] DEBUG RAW RESPONSE ON FAILURE: %s", rawResponse)
			continue
		}

		// SUCCESS
		logWithTrace(ctx, "[LLM] Attempt %d/%d SUCCESS", attempt+1, maxAttempts)
		logWithTrace(ctx, "[LLM] RAW RESPONSE (SUCCESS): %s", rawResponse) //Temp
		return &event, nil
	}

	return nil, fmt.Errorf("LLM failed after %d attempts. Last error: %v", maxAttempts, lastErr)
}

func isValidForImplicit(ev *LLMEvent) bool {
	// If it's a creation request but has no link, we need at least a start time and some note/context
	return ev.StartTimeUTC != "" && (ev.EndTimeUTC != "")
}

func logWithTrace(ctx context.Context, format string, args ...interface{}) {
	traceID, _ := ctx.Value(traceKey).(string)
	prefix := ""
	if traceID != "" {
		prefix = "[trace=" + traceID + "] "
	}
	log.Printf(prefix+format, args...)
}

// saveExpiredJob writes a minimal "expired" record to the DB so that meetings
// rejected due to already-ended or stale conditions are still visible in the dashboard.
func saveExpiredJob(ctx context.Context, ev *LLMEvent, recipientEmail string, startTime time.Time, endTimeUTC string, traceID string, reason string) {
	agentEmail := recipientEmail
	if agentEmail == "" {
		agentEmail = ev.Organizer
	}
	jobName := ev.Title
	if strings.TrimSpace(jobName) == "" {
		if ev.MeetingLink != "" {
			jobName = ev.MeetingLink
		} else {
			jobName = "meeting-" + uuid.New().String()[:8]
		}
	}
	expiry := endTimeUTC
	if expiry == "" {
		expiry = startTime.Add(1 * time.Hour).Format(time.RFC3339)
	}
	job := ScheduledJob{
		ID:           uuid.New().String(),
		Name:         jobName,
		MeetingURL:   ev.MeetingLink,
		AgentEmailID: agentEmail,
		Status:       "expired",
		StartTime:    startTime.Format(time.RFC3339),
		Expiry:       expiry,
		UID:          ev.UID,
		TraceID:      traceID,
	}
	if err := AddJobToDB(ctx, job); err != nil {
		logWithTrace(ctx, "[saveExpiredJob] WARNING: could not persist expired job to DB (%s): %v", reason, err)
	} else {
		logWithTrace(ctx, "[saveExpiredJob] Expired job saved to DB id=%s reason=%s", job.ID, reason)
	}
}

func HandleAWSLLM(w http.ResponseWriter, r *http.Request) {
	traceID := "Email-" + uuid.New().String()[:5] //Trace
	ctx := context.WithValue(context.Background(), traceKey, traceID)
	log.Printf("[trace=%s] START HandleAWSLLM", traceID)
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	logWithTrace(ctx, "[HandleAWSLLM] Received webhook — body size=%d bytes, Content-Type=%s", len(raw), r.Header.Get("Content-Type"))

	recipientEmail := extractRecipientEmail(raw)
	if recipientEmail == "" {
		logWithTrace(ctx, "[HandleAWSLLM] WARNING: could not extract recipient email from headers")
	} else {
		logWithTrace(ctx, "[HandleAWSLLM] Recipient (agent inbox): %s", recipientEmail)
	}

	content, err := BuildLLMPayload(raw)
	if err != nil {
		logWithTrace(ctx, "BuildLLMPayload error: %v", err)
		logWithTrace(ctx, "failed to extract content")
		return
	}
	if content == "" {
		logWithTrace(ctx, "BuildLLMPayload returned empty content")
		http.Error(w, "no usable content", http.StatusBadRequest)
		return
	}
	logWithTrace(ctx, "[HandleAWSLLM] LLM input ready (%d chars)", len(content))

	//  LLM call
	ev, err := callLLMWithRetry(ctx, content)
	if err != nil {
		logWithTrace(ctx, "[HandleAWSLLM] Extraction completely failed: %v", err)
		http.Error(w, "Extraction failed", http.StatusInternalServerError)
		return
	}

	if ev.EventType == "none" || (ev.MeetingLink == "" && ev.EventType == "create" && !isValidForImplicit(ev)) {
		logWithTrace(ctx, "[HandleAWSLLM] No meeting detected (refusal or insufficient data)")
		return
	}

	if corrected, ok := extractUTCFromICS(content); ok {
		if ev.StartTimeUTC != corrected {
			logWithTrace(ctx, "[HandleAWSLLM] TIMEZONE CORRECTION: LLM said %q, ICS parser says %q — using ICS value", ev.StartTimeUTC, corrected)
			ev.StartTimeUTC = corrected
		} else {
			logWithTrace(ctx, "[HandleAWSLLM] Timezone check passed: LLM time %q matches ICS parse %q", ev.StartTimeUTC, corrected)
		}
	}
	if endUTC, ok := extractUTCFromICS(content, "DTEND"); ok {
		if endUTC != ev.EndTimeUTC {
			logWithTrace(ctx, "[HandleAWSLLM] TIMEZONE CORRECTION (end): LLM said %q, ICS parser says %q — using ICS value", ev.EndTimeUTC, endUTC)
			ev.EndTimeUTC = endUTC
		}
	}

	logWithTrace(ctx, "[HandleAWSLLM] Parsed event — type=%q uid=%q title=%q meetingLink=%q startTime=%q cron=%q organizer=%q",
		ev.EventType, ev.UID, ev.Title, ev.MeetingLink, ev.StartTimeUTC, ev.Cron, ev.Organizer)
	// Safeguard for case: If the agent is removed from the guest list,
	// or if the ICS method is explicitly CANCEL, force a cancellation.
	isCancelMethod := strings.Contains(strings.ToUpper(content), "METHOD:CANCEL")
	agentMissing := recipientEmail != "" && content != "" && !strings.Contains(strings.ToLower(content), strings.ToLower(recipientEmail))

	if isCancelMethod || agentMissing {
		if ev.EventType != "cancel" {
			logWithTrace(ctx, "[HandleAWSLLM] Forcing CANCEL (isCancelMethod=%v, agentMissing=%v) — LLM originally said %q", isCancelMethod, agentMissing, ev.EventType)
			ev.EventType = "cancel"
		}
	}

	if ev.EventType == "cancel" {
		logWithTrace(ctx, "[HandleAWSLLM] CANCEL event received — meetingLink=%q uid=%q recurrenceID=%q", ev.MeetingLink, ev.UID, ev.RecurrenceID)

		// Single-occurrence cancel: RECURRENCE-ID is present, meaning only one slot
		// of a recurring series is being cancelled — leave the series job intact.
		if ev.RecurrenceID != "" {
			logWithTrace(ctx, "[HandleAWSLLM] Single-occurrence cancel (recurrence_id=%q) — advancing series start_time past cancelled occurrence", ev.RecurrenceID)

			// Parse the cancelled occurrence time from RECURRENCE-ID (ICS timestamp, e.g. 20260501T043000Z).
			rid := strings.TrimSpace(ev.RecurrenceID)
			var cancelledTime time.Time
			for _, layout := range []string{"20060102T150405Z", "20060102T150405"} {
				if t, parseErr := time.Parse(layout, rid); parseErr == nil {
					cancelledTime = t.UTC()
					break
				}
			}

			if cancelledTime.IsZero() {
				logWithTrace(ctx, "[HandleAWSLLM] Single-occurrence cancel: could not parse RECURRENCE-ID %q — start_time not advanced", ev.RecurrenceID)
			} else {
				seriesJobs, dbErr := GetAllJobsFromDB([]string{})
				if dbErr != nil {
					logWithTrace(ctx, "[HandleAWSLLM] Single-occurrence cancel: DB error looking up series: %v", dbErr)
				} else {
					for _, j := range seriesJobs {
						if j.Cron == "" {
							continue // skip one-time jobs
						}
						matchByUID := ev.UID != "" && j.UID == ev.UID
						matchByURL := ev.MeetingLink != "" && j.MeetingURL == ev.MeetingLink
						if !matchByUID && !matchByURL {
							continue
						}
						cronSched, cronErr := cron.ParseStandard(j.Cron)
						if cronErr != nil {
							logWithTrace(ctx, "[HandleAWSLLM] Single-occurrence cancel: bad cron %q for job %s: %v", j.Cron, j.ID, cronErr)
							continue
						}
						nextRun := cronSched.Next(cancelledTime)
						if nextRun.IsZero() {
							logWithTrace(ctx, "[HandleAWSLLM] Single-occurrence cancel: cron %q has no future run after %v (job %s)", j.Cron, cancelledTime, j.ID)
							continue
						}
						// Only advance if nextRun is actually after the currently scheduled StartTime
						// This prevents moving a future series (e.g. starting in June) backward
						// just because an earlier occurrence was cancelled.
						if currentST, err := time.Parse(time.RFC3339, j.StartTime); err == nil {
							if nextRun.Before(currentST) || nextRun.Equal(currentST) {
								logWithTrace(ctx, "[HandleAWSLLM] Single-occurrence cancel: next run %v is not after current start %v — skipping update for %s", nextRun, currentST, j.ID)
								continue
							}
						}

						nextStr := nextRun.UTC().Format(time.RFC3339)
						if updErr := UpdateJobStartTimeInDB(j.ID, nextStr); updErr != nil {
							logWithTrace(ctx, "[HandleAWSLLM] Single-occurrence cancel: DB update failed for job %s: %v", j.ID, updErr)
						} else {
							logWithTrace(ctx, "[HandleAWSLLM] Single-occurrence cancel: job %s start_time successfully advanced to %s in DB", j.ID, nextStr)
							// Mirror the change in the in-memory slice.
							scheduleMu.Lock()
							for i, sj := range ScheduledJobs {
								if sj.ID == j.ID {
									ScheduledJobs[i].StartTime = nextStr
									break
								}
							}
							scheduleMu.Unlock()

							// If we skipped the current/upcoming occurrence, also remove any pending bootstrap OneTime job.
							Scheduler.RemoveByTags(j.ID + ":bootstrap")
							logWithTrace(ctx, "[HandleAWSLLM] Single-occurrence cancel: processed bootstrap job removal for %s", j.ID)
						}
					}
				}
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "occurrence_skipped", "recurrence_id": ev.RecurrenceID})
			return
		}

		if ev.MeetingLink == "" {
			logWithTrace(ctx, "[HandleAWSLLM] Cancel event has no meeting link — cannot find job to cancel")
			return
		}
		existing, err := GetAllJobsFromDB([]string{})
		if err != nil {
			logWithTrace(ctx, "[HandleAWSLLM] Cancel: DB error: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		cancelled := 0
		for _, j := range existing {
			matchByUID := ev.UID != "" && j.UID == ev.UID
			matchByURL := ev.MeetingLink != "" && j.MeetingURL == ev.MeetingLink
			if (matchByUID || matchByURL) && (j.Status == "scheduled" || j.Status == "" || j.Status == "processing" || j.Status == "retrying" || recentlyFired(j.CreatedAt)) {
				if matchByUID {
					logWithTrace(ctx, "[HandleAWSLLM] Cancel: matched job %s by UID=%q", j.ID, ev.UID)
				} else {
					logWithTrace(ctx, "[HandleAWSLLM] Cancel: matched job %s by URL=%q", j.ID, ev.MeetingLink)
				}
				// Remove from gocron scheduler
				if jobUUID, err := uuid.Parse(j.ID); err == nil {
					if err := Scheduler.RemoveJob(jobUUID); err != nil {
						logWithTrace(ctx, "[HandleAWSLLM] Cancel: could not remove job %s from scheduler: %v", j.ID, err)
					} else {
						logWithTrace(ctx, "[HandleAWSLLM] Cancel: removed job %s from scheduler", j.ID)
					}
				}
				cancelCronJobIfExists(j.ID)
				// Remove from in-memory list
				ScheduledJobs = RemoveScheduleJob(ScheduledJobs, j.ID)
				// Update status in DB
				if err := UpdateJobStatusInDB(j.ID, "cancelled"); err != nil {
					logWithTrace(ctx, "[HandleAWSLLM] Cancel: failed to update DB for job %s: %v", j.ID, err)
				} else {
					logWithTrace(ctx, "[HandleAWSLLM] Cancel: job %s marked cancelled in DB", j.ID)
				}
				cancelled++
			}
		}
		logWithTrace(ctx, "[HandleAWSLLM] Cancel complete — %d job(s) cancelled for meetingURL=%q", cancelled, ev.MeetingLink)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "cancelled", "jobs_cancelled": cancelled})
		return
	}

	if ev.EventType == "update" {
		logWithTrace(ctx, "[HandleAWSLLM] UPDATE event received — meetingLink=%q uid=%q newStartTime=%q", ev.MeetingLink, ev.UID, ev.StartTimeUTC)
		if ev.MeetingLink == "" {
			logWithTrace(ctx, "[HandleAWSLLM] Update event has no meeting link — skipping")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Handle single-occurrence update (RecurrenceID present)
		if ev.RecurrenceID != "" {
			// Parse the new start time
			newStart, err := time.Parse(time.RFC3339, ev.StartTimeUTC)
			if err != nil {
				logWithTrace(ctx, "[HandleAWSLLM] UPDATE: invalid start_time %q", ev.StartTimeUTC)
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "reason": "invalid_start_time"})
				return
			}
			if newStart.After(time.Now().UTC()) {
				// Schedule a one-time job for this specific occurrence
				logWithTrace(ctx, "[HandleAWSLLM] UPDATE: scheduling future occurrence for RecurrenceID %q", ev.RecurrenceID)
				// Proceed to cancel previous one-time jobs but KEEP the series job
			} else {
				// Occurrence is in the past; no need to schedule a new job.
				logWithTrace(ctx, "[HandleAWSLLM] UPDATE: occurrence %q is in the past – no reschedule", ev.RecurrenceID)
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_action", "reason": "occurrence_in_past"})
				return
			}
		} else {
			// Full series or one-time job update (no RecurrenceID)
			// Parse the new start time to decide if we should reschedule
			newStart, parseErr := time.Parse(time.RFC3339, ev.StartTimeUTC)
			if parseErr == nil && newStart.Before(time.Now().UTC()) && ev.RRule == "" && ev.Cron == "" {
				logWithTrace(ctx, "[HandleAWSLLM] UPDATE: start_time %q is in the past for a non-recurring event — no reschedule", ev.StartTimeUTC)
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_action", "reason": "start_time_in_past"})
				return
			}
		}

		// Cancel existing job for this meeting, then fall through to re-schedule below
		existing, err := GetAllJobsFromDB([]string{})
		if err != nil {
			logWithTrace(ctx, "[HandleAWSLLM] Update: DB error: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		for _, j := range existing {
			matchByUID := ev.UID != "" && j.UID == ev.UID
			matchByURL := ev.MeetingLink != "" && j.MeetingURL == ev.MeetingLink
			if (matchByUID || matchByURL) && (j.Status == "scheduled" || j.Status == "" || j.Status == "processing" || j.Status == "retrying") {
				// CRITICAL: If this is a single occurrence update, do NOT cancel the series job
				if ev.RecurrenceID != "" && j.Cron != "" {
					logWithTrace(ctx, "[HandleAWSLLM] Update: keeping series job %s while updating occurrence %q", j.ID, ev.RecurrenceID)
					continue
				}

				if matchByUID {
					logWithTrace(ctx, "[HandleAWSLLM] Update: matched old job %s by UID=%q", j.ID, ev.UID)
				} else {
					logWithTrace(ctx, "[HandleAWSLLM] Update: matched old job %s by URL=%q", j.ID, j.MeetingURL)
				}
				if jobUUID, err := uuid.Parse(j.ID); err == nil {
					Scheduler.RemoveJob(jobUUID)
					cancelCronJobIfExists(j.ID)
				}
				ScheduledJobs = RemoveScheduleJob(ScheduledJobs, j.ID)
				UpdateJobStatusInDB(j.ID, "superseded")
				logWithTrace(ctx, "[HandleAWSLLM] Update: cancelled old job %s (will reschedule at new time)", j.ID)
			}
		}
		logWithTrace(ctx, "[HandleAWSLLM] Update: old job removal complete, rescheduling at new start time %q", ev.StartTimeUTC)
	}

	if ev.EventType == "none" {
		logWithTrace(ctx, "[HandleAWSLLM] event_type=none — no action")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if ev.EventType != "create" && ev.EventType != "update" {
		logWithTrace(ctx, "[HandleAWSLLM] Unknown event_type=%q — skipping", ev.EventType)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if ev.MeetingLink == "" {
		logWithTrace(ctx, "[HandleAWSLLM] Skipping — no meeting link, cron, or start_time in LLM response")
		http.Error(w, "no meeting link, cron, or start_time provided", http.StatusBadRequest)
		return
	}

	// Dedup: prevent double-scheduling
	// Normalize meeting link
	ev.MeetingLink = strings.TrimSpace(ev.MeetingLink)
	// Strip any duplicate schema
	for strings.HasPrefix(ev.MeetingLink, "https://https://") {
		ev.MeetingLink = strings.TrimPrefix(ev.MeetingLink, "https://")
	}
	for strings.HasPrefix(ev.MeetingLink, "http://https://") {
		ev.MeetingLink = strings.TrimPrefix(ev.MeetingLink, "http://")
	}
	// Add schema if missing
	if ev.MeetingLink != "" && !strings.HasPrefix(ev.MeetingLink, "http://") && !strings.HasPrefix(ev.MeetingLink, "https://") {
		ev.MeetingLink = "https://" + ev.MeetingLink
	}
	logWithTrace(ctx, "[HandleAWSLLM] Normalized meetingLink=%q", ev.MeetingLink)
	scheduleMu.Lock()
	existing, dbErr := GetAllJobsFromDB([]string{})
	if dbErr == nil {
		for _, j := range existing {
			// Skip jobs that are explicitly dead/replaced — these do NOT count as duplicates.
			// "completed" and "failed" are also skipped so a re-sent invite after a
			// finished or failed meeting is allowed to schedule a new job.
			if j.Status == "superseded" || j.Status == "cancelled" ||
				j.Status == "expired" || j.Status == "completed" || j.Status == "failed" {
				continue
			}
			msg, _ := mail.ReadMessage(bytes.NewReader(raw))
			subject := ""
			if msg != nil {
				subject = msg.Header.Get("Subject")
			}
			uidMatch := ev.UID != "" && j.UID == ev.UID

			isForwarded := false
			for _, h := range []string{"X-MS-Exchange-MeetingForward-Message", "Auto-Submitted"} {
				if val := r.Header.Get(h); val != "" && strings.Contains(strings.ToLower(val), "forward") {
					isForwarded = true
					break
				}
			}
			if !isForwarded {
				if strings.HasPrefix(strings.ToLower(subject), "fw:") ||
					strings.HasPrefix(strings.ToLower(subject), "fwd:") {
					isForwarded = true
				}
			}

			inboxEmail := recipientEmail
			if inboxEmail == "" {
				inboxEmail = ev.Organizer
			}
			urlMatch := j.MeetingURL == ev.MeetingLink && ev.MeetingLink != "" &&
				(inboxEmail == "" || j.AgentEmailID == inboxEmail)

			if uidMatch {
				// Forwarded to a different agent → not a duplicate, let it create a new job.
				if isForwarded && inboxEmail != "" && j.AgentEmailID != inboxEmail {
					// fall through
				} else {
					// Same agent (or non-forwarded): check one-time vs series edge case.
					isNewOneTime := ev.RRule == "" && ev.Cron == ""
					if isNewOneTime && j.Cron != "" {
						logWithTrace(ctx, "[HandleAWSLLM] Allowing one-time update to coexist with series job %s", j.ID)
						continue
					}
					scheduleMu.Unlock()
					logWithTrace(ctx, "[HandleAWSLLM] DUPLICATE -- job already exists (jobID=%q status=%q uid_match=%v), skipping", j.ID, j.Status, uidMatch)
					return
				}
			} else if urlMatch {
				isNewOneTime := ev.RRule == "" && ev.Cron == ""
				if isNewOneTime && j.Cron != "" {
					logWithTrace(ctx, "[HandleAWSLLM] Allowing one-time update to coexist with series job %s", j.ID)
					continue
				}
				scheduleMu.Unlock()
				logWithTrace(ctx, "[HandleAWSLLM] DUPLICATE -- job already exists (jobID=%q status=%q url_match=%v), skipping", j.ID, j.Status, urlMatch)
				return
			}
		}
	}

	// Determine fire time for the one-time job.

	var startTime time.Time
	happeningNow := false

	if ev.StartTimeUTC == "" {
		time.Sleep(2 * time.Second)

		logWithTrace(ctx, "[HandleAWSLLM] No start_time_utc — treating as HAPPENING NOW, firing immediately")
		startTime = time.Now().UTC()
		happeningNow = true
	} else if t, err := time.Parse(time.RFC3339, ev.StartTimeUTC); err != nil {
		logWithTrace(ctx, "[HandleAWSLLM] Could not parse start_time %q (%v) — firing immediately", ev.StartTimeUTC, err)
		startTime = time.Now().UTC()
		happeningNow = true
	} else {
		startTime = t
	}

	now := time.Now().UTC()
	delay := startTime.Sub(now)

	cronExpr := ""
	isRecurring := ev.RRule != "" || ev.Cron != ""
	if ev.RRule != "" {
		cronExpr = convertRRuleToCron(ev.RRule, startTime)
		if cronExpr != "" {
			logWithTrace(ctx, "[HandleAWSLLM] Recurring meeting — RRULE=%q → cron=%q", ev.RRule, cronExpr)
		} else {
			logWithTrace(ctx, "[HandleAWSLLM] Could not convert RRULE=%q to cron — scheduling as one-time", ev.RRule)
		}
	}

	// STALE MEETING , If EndTime is already in the past, the meeting is over.
	if ev.EndTimeUTC != "" {
		if end, err := time.Parse(time.RFC3339, ev.EndTimeUTC); err == nil {
			if now.After(end) && !isRecurring {
				scheduleMu.Unlock()
				logWithTrace(ctx, "[HandleAWSLLM] Meeting already ended at %s (now %s) — marking expired", ev.EndTimeUTC, now.Format(time.RFC3339))
				saveExpiredJob(ctx, ev, recipientEmail, startTime, ev.EndTimeUTC, traceID, "meeting_already_ended")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{"status": "expired", "reason": "meeting_already_ended"})
				return
			}
		}
	}

	// If the meeting started too long ago don't auto-join.
	staleThreshold := 15 * time.Minute
	isStale := delay < -staleThreshold

	// For recurring meetings, we might be stale relative to the series START (e.g. 2025),
	// but an occurrence might be happening RIGHT NOW.
	if isRecurring && isStale && cronExpr != "" {
		if sched, err := cron.ParseStandard(cronExpr); err == nil {
			// Check if an occurrence started within the last 15 minutes.
			// Next(now - 15m) will give the first occurrence >= now - 15m.
			lastPossibleStart := now.Add(-staleThreshold)
			nextOccur := sched.Next(lastPossibleStart)

			// If the next occurrence after (now - 15m) is already in the past (or now),
			// it means a meeting started recently!
			if !nextOccur.After(now) {
				logWithTrace(ctx, "[HandleAWSLLM] Recurring occurrence started recently (%s) — overriding stale=false to allow late join", nextOccur.Format(time.RFC3339))
				isStale = false
				startTime = nextOccur // use the actual occurrence time for the bootstrap
				delay = startTime.Sub(now)
			}
		}
	}

	if isStale {
		logWithTrace(ctx, "[HandleAWSLLM] Meeting start %s was %s ago (>%s) — skipping auto-join (stale)",
			startTime.Format(time.RFC3339), (-delay).Round(time.Second), staleThreshold)

		// For one-time stale jobs, recorded as expired.
		if !isRecurring {
			scheduleMu.Unlock()
			saveExpiredJob(ctx, ev, recipientEmail, startTime, ev.EndTimeUTC, traceID, "stale_meeting_late_join")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "expired", "reason": "stale_meeting_late_join"})
			return
		}
		// For recurring jobs, cron continues but no fire immediately.
		happeningNow = false
	} else if happeningNow || delay <= 0 {
		if !happeningNow {
			logWithTrace(ctx, "[HandleAWSLLM] Meeting start %s was %s ago — firing immediately (late join)", startTime.Format(time.RFC3339), (-delay).Round(time.Second))
		}
		startTime = time.Now().UTC()
		happeningNow = true
	} else {
		logWithTrace(ctx, "[HandleAWSLLM] Meeting starts in %s (at %s UTC) — job will fire then", delay.Round(time.Second), startTime.Format(time.RFC3339))
	}

	agentEmail := recipientEmail
	if agentEmail == "" {
		logWithTrace(ctx, "[HandleAWSLLM] Falling back to organizer email as agent email: %s", ev.Organizer)
		agentEmail = ev.Organizer
	}

	jobName := ev.Title
	if strings.TrimSpace(jobName) == "" {
		if ev.MeetingLink != "" {
			jobName = ev.MeetingLink
		} else {
			jobName = "meeting-" + uuid.New().String()[:8]
		}
		logWithTrace(ctx, "[HandleAWSLLM] Empty title from LLM — using fallback name: %q", jobName)
	}

	status := "scheduled"
	if happeningNow {
		status = "processing"
	}
	job := ScheduledJob{
		ID:           uuid.New().String(),
		Name:         jobName,
		MeetingURL:   ev.MeetingLink,
		AgentEmailID: agentEmail,
		Status:       status,
		Cron:         cronExpr,                       // empty = one-time, set = recurring cron
		StartTime:    startTime.Format(time.RFC3339), // exact fire time
		UID:          ev.UID,                         // calendar UID
		TraceID:      traceID,
	}
	if cronExpr != "" {
		// Use UNTIL from RRULE if present, otherwise default to 6 months.
		job.Expiry = startTime.AddDate(0, 6, 0).Format(time.RFC3339)
		if until, ok := extractRRuleUntil(ev.RRule); ok {
			job.Expiry = until.Format(time.RFC3339)
			logWithTrace(ctx, "[HandleAWSLLM] Recurring job UNTIL=%s extracted from RRULE", job.Expiry)
		} else {
			logWithTrace(ctx, "[HandleAWSLLM] Recurring job — no UNTIL in RRULE, defaulting expiry to 1 year: %s", job.Expiry)
		}
	} else if ev.EndTimeUTC != "" {
		// One-time job: expiry = meeting end time
		job.Expiry = ev.EndTimeUTC
	} else {
		job.Expiry = startTime.Add(1 * time.Hour).Format(time.RFC3339)
	}

	logWithTrace(ctx, "[HandleAWSLLM] Scheduling job — id=%s name=%q meetingURL=%q agentEmail=%q fireAt=%s expiry=%q",
		job.ID, job.Name, job.MeetingURL, job.AgentEmailID, startTime.Format(time.RFC3339), job.Expiry)

	_, err = ScheduleJob(ctx, job)
	if err != nil {
		scheduleMu.Unlock()
		logWithTrace(ctx, "[HandleAWSLLM] ERROR scheduling job: %v", err)
		http.Error(w, "failed to schedule job", http.StatusInternalServerError)
		return
	}
	// CRITICAL: Stop overwriting job.ID with gocron's ID.
	// The internal UUID generated above must be used for consistency in DB and Registry.
	logWithTrace(ctx, "[HandleAWSLLM] Job scheduled successfully id=%s", job.ID)

	// For recurring jobs whose series started in the past, update start_time to the
	// next actual cron occurrence BEFORE persisting to DB. Otherwise AddJobToDB would
	// write the original (stale) start_time and overwrite the UpdateJobStartTimeInDB
	// call that happened inside scheduleCronJob (which ran before the row existed).
	if job.Cron != "" {
		if sched, err := cron.ParseStandard(job.Cron); err == nil {
			if nextRun := sched.Next(time.Now().UTC()); !nextRun.IsZero() {
				job.StartTime = nextRun.Format(time.RFC3339)
				logWithTrace(ctx, "[HandleAWSLLM] Recurring job — persisting with next cron run start_time=%s", job.StartTime)
			}
		}
	}

	ScheduledJobs = append(ScheduledJobs, job)
	if err := AddJobToDB(ctx, job); err != nil {
		logWithTrace(ctx, "[HandleAWSLLM] WARNING: job scheduled but failed to persist to DB: %v", err)
	} else {
		logWithTrace(ctx, "[HandleAWSLLM] Job persisted to DB id=%s", job.ID)
	}
	scheduleMu.Unlock()

	w.WriteHeader(http.StatusOK) // Ack SES now
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "scheduled", "job_id": job.ID})
}

var jsonObjRe = regexp.MustCompile(`(?s)\{.*\}`)

func extractFirstJSONObject(s string) string {
	s = strings.TrimSpace(s)

	// Remove markdown code blocks if present
	if idx := strings.Index(s, "{"); idx != -1 {
		if lastIdx := strings.LastIndex(s, "}"); lastIdx != -1 && lastIdx > idx {
			return s[idx : lastIdx+1]
		}
	}
	m := jsonObjRe.FindString(s)
	return strings.TrimSpace(m)
}

func normalizeLLMJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		if unq, err := strconv.Unquote(s); err == nil {
			s = unq
		}
	}
	s = strings.TrimSpace(s)
	return s
}

// enrichRecurringJob refreshes a recurring job's start_time and status at read-time
// so the API always returns the next upcoming occurrence instead of a stale past time.
func enrichRecurringJob(j ScheduledJob) ScheduledJob {
	if j.Cron == "" {
		return j
	}
	// Parse expiry — if the series has expired, leave it alone.
	if j.Expiry != "" {
		if exp, err := time.Parse(time.RFC3339, j.Expiry); err == nil && time.Now().UTC().After(exp) {
			return j
		}
	}
	baseline := time.Now().UTC()
	if j.StartTime != "" {
		if st, err := time.Parse(time.RFC3339, j.StartTime); err == nil {
			// If current start time is in the future, it might be a skipped/cancelled occurrence.
			// Use it as the baseline for the next run calculation to prevent regressing the UI time.
			if st.After(baseline) {
				baseline = st.Add(-1 * time.Second)
			}
		}
	}

	sched, err := cron.ParseStandard(j.Cron)
	if err != nil {
		return j
	}

	nextRun := sched.Next(baseline)
	if nextRun.IsZero() {
		return j
	}
	j.StartTime = nextRun.Format(time.RFC3339)
	// Keep the parent series as "scheduled" while the cron is active.
	if j.Status == "failed" || j.Status == "processing" || j.Status == "retrying" {
		j.Status = "scheduled"
	}
	return j
}

// Method to get all scheduled job
func HandleGetAllScheduledJobs(w http.ResponseWriter, r *http.Request) {
	scheduledJobsFromDB, err := GetAllJobsFromDB([]string{})
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for i, j := range scheduledJobsFromDB {
		scheduledJobsFromDB[i] = enrichRecurringJob(j)
	}
	scheduledJobs, err := json.Marshal(scheduledJobsFromDB)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(scheduledJobs)
}

// Method to get scheduled job by ID
func HandleGetScheduledJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")
	// Get job from DB
	scheduledJobsFromDB, err := GetAllJobsFromDB([]string{}) // TODO: Make this faster
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, item := range scheduledJobsFromDB {
		if item.ID == jobID {
			item = enrichRecurringJob(item)
			scheduledJob, err := json.Marshal(item)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(scheduledJob)
			return
		}
	}
	// Return 404
	w.WriteHeader(http.StatusNotFound)
}

// Method to delete a scheduled job by ID
func HandleDeleteScheduledJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")
	// Get job from DB
	scheduledJobsFromDB, err := GetAllJobsFromDB([]string{}) // TODO: Make this faster
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, item := range scheduledJobsFromDB {
		if item.ID == jobID {
			if item.Status == "cancelled" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			scheduleMu.Lock()
			UpdateJobStatusInDB(item.ID, "cancelled")
			Scheduler.RemoveByTags(item.ID)
			cancelCronJobIfExists(item.ID)
			log.Printf("%s Job removed from the scheduler\n", item.ID)
			ScheduledJobs = RemoveScheduleJob(ScheduledJobs, item.ID)
			scheduleMu.Unlock()

			item.Status = "cancelled" // reflect actual DB state in response
			scheduledJob, err := json.Marshal(item)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(scheduledJob)
			return
		}
	}
	// Return 404
	w.WriteHeader(http.StatusNotFound)
}

// Method to add new Scheduled Job
func HandleAddScheduledJob(w http.ResponseWriter, r *http.Request) {
	traceID := "Api-" + uuid.New().String()[:5]
	ctx := context.WithValue(context.Background(), traceKey, traceID)
	// Parse the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	var newJob ScheduledJob
	if err = json.Unmarshal(body, &newJob); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Println("Starting new Job from Rest API Trigger")
	if newJob.ID == "" {
		newJob.ID = uuid.New().String()
	}
	// Job Schedule Logic
	if newJob.Expiry == "" || len(newJob.Expiry) <= 3 {
		// If Expiry is not found; then set the default as 30 days
		now := time.Now().UTC().AddDate(0, 0, 30)
		layout := "2006-01-02T15:04:05Z07:00"
		newJob.Expiry = now.Format(layout)
	}
	scheduleMu.Lock()
	existing, dbErr := GetAllJobsFromDB([]string{})
	if dbErr == nil {
		for _, j := range existing {
			// Skip dead/finished jobs so re-scheduling the same meeting is allowed.
			if j.Status == "superseded" || j.Status == "cancelled" ||
				j.Status == "expired" || j.Status == "completed" || j.Status == "failed" {
				continue
			}

			uidMatch := newJob.UID != "" && j.UID == newJob.UID
			urlMatch := j.MeetingURL == newJob.MeetingURL && newJob.MeetingURL != "" &&
				(newJob.AgentEmailID == "" || j.AgentEmailID == newJob.AgentEmailID)

			if uidMatch || urlMatch {
				scheduleMu.Unlock()
				log.Printf("[HandleAddScheduledJob] DUPLICATE -- job already exists (jobID=%q status=%q uid_match=%v url_match=%v)", j.ID, j.Status, uidMatch, urlMatch)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"status": "already_scheduled",
					"job_id": j.ID,
					"reason": "duplicate_event",
				})
				return
			}
		}
	}
	status := "scheduled"
	if newJob.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, newJob.StartTime); err == nil {
			if t.Before(time.Now().UTC().Add(1 * time.Minute)) {
				status = "processing"
			}
		}
	} else {
		status = "processing"
	}
	newJob.Status = status
	// Schedule a new job

	newJob.TraceID = traceID
	_, err = ScheduleJob(ctx, newJob)
	if err != nil {
		scheduleMu.Unlock()
		log.Printf("[HandleAddScheduledJob] Failed to schedule: %v", err)
		http.Error(w, "failed to schedule job", http.StatusInternalServerError)
		return
	}
	// Set Job ID and return
	// id is the gocron internal ID; we keep using newJob.ID (our internal UUID) for consistency.
	log.Printf("[HandleAddScheduledJob] Job scheduled id=%s", newJob.ID)
	ScheduledJobs = append(ScheduledJobs, newJob)
	// Add new item to database

	err = AddJobToDB(ctx, newJob)
	if err != nil {
		fmt.Println(err)
	}
	scheduleMu.Unlock()
	job, err := json.Marshal(newJob)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(job)
}

func HandleUpdateScheduledJob(w http.ResponseWriter, r *http.Request) {
	traceID := "API-" + uuid.New().String()[:5]
	ctx := context.WithValue(context.Background(), traceKey, traceID)
	jobID := chi.URLParam(r, "jobID")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	var patch struct {
		MeetingURL   string `json:"meetingUrl"`
		StartTime    string `json:"start_time"`
		Expiry       string `json:"expiry"`
		AgentEmailID string `json:"agentEmailID"`
		Name         string `json:"name"`
		Cron         string `json:"cron"`
	}
	if err := json.Unmarshal(body, &patch); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Find job in DB
	existing, err := GetAllJobsFromDB([]string{})
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	var found *ScheduledJob
	for _, j := range existing {
		if j.ID == jobID {
			copy := j
			found = &copy
			break
		}
	}
	if found == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	scheduleMu.Lock()
	// Remove old job from scheduler
	if oldUUID, err := uuid.Parse(found.ID); err == nil {
		if err := Scheduler.RemoveJob(oldUUID); err != nil {
			log.Printf("[HandleUpdateScheduledJob] Job not in scheduler (may have already fired): %v", err)
		}
	}
	cancelCronJobIfExists(found.ID)
	ScheduledJobs = RemoveScheduleJob(ScheduledJobs, found.ID)
	UpdateJobStatusInDB(found.ID, "superseded")

	// Apply only provided fields
	if patch.MeetingURL != "" {
		found.MeetingURL = patch.MeetingURL
	}
	if patch.StartTime != "" {
		found.StartTime = patch.StartTime
	}
	if patch.Expiry != "" {
		found.Expiry = patch.Expiry
	}
	if patch.AgentEmailID != "" {
		found.AgentEmailID = patch.AgentEmailID
	}
	if patch.Name != "" {
		found.Name = patch.Name
	}
	if patch.Cron != "" {
		found.Cron = patch.Cron
	}
	found.Status = "scheduled"
	found.ID = uuid.New().String() // new ID for new job

	// Reschedule
	newID, err := ScheduleJob(ctx, *found)
	if err != nil {
		scheduleMu.Unlock()
		log.Printf("[HandleUpdateScheduledJob] Failed to reschedule: %v", err)
		http.Error(w, "failed to reschedule job", http.StatusInternalServerError)
		return
	}
	found.ID = newID.String()
	ScheduledJobs = append(ScheduledJobs, *found)

	if err := AddJobToDB(ctx, *found); err != nil {
		log.Printf("[HandleUpdateScheduledJob] Failed to persist updated job: %v", err)
	}
	scheduleMu.Unlock()

	log.Printf("[HandleUpdateScheduledJob] Job updated — oldID=%s newID=%s", jobID, found.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(found)
}

// Method to remove job entry from schedule jobs list
func RemoveScheduleJob(scheduledJobs []ScheduledJob, jobID string) []ScheduledJob {
	for i, item := range scheduledJobs {
		if item.ID == jobID {
			return append(scheduledJobs[:i], scheduledJobs[i+1:]...)
		}
	}
	return scheduledJobs
}

func joinMeeting(ctx context.Context, nj ScheduledJob, label string, failed *bool) {
	logWithTrace(ctx, "[ScheduleJob:%s] Job fired — Name=%q MeetingURL=%q AgentEmailID=%q", label, nj.Name, nj.MeetingURL, nj.AgentEmailID)
	lkRoomID := uuid.New().String()
	traceID := uuid.New().String()
	JobTraceMap[lkRoomID] = traceID
	logWithTrace(ctx, "[ScheduleJob:%s] Generated lkRoomID=%s", label, lkRoomID)
	// Temporary state Processing
	UpdateJobStatusInDB(nj.ID, "processing")
	logWithTrace(ctx, "[%s] Job status set to processing — lkRoomID=%s", label, lkRoomID)

	//Avatar Retry attempt
	avatarBackoff := []time.Duration{0, 5 * time.Second, 15 * time.Second}
	maxAvatarAttempts := 3
	avatarSuccess := false
	dn := "Lisa" // fallback
	generatedVideoUrl := ""
	var avatarErr error

	for attempt := 0; attempt < maxAvatarAttempts; attempt++ {
		if avatarBackoff[attempt] > 0 {
			logWithTrace(ctx, "[%s][Avatar] Retrying — attempt %d/%d — waiting %s", label, attempt+1, maxAvatarAttempts, avatarBackoff[attempt])
			UpdateJobStatusInDB(nj.ID, "retrying")
			time.Sleep(avatarBackoff[attempt])
		}
		logWithTrace(ctx, "[%s][Avatar] Attempt %d/%d — posting avatar request", label, attempt+1, maxAvatarAttempts)

		var name string
		var vUrl string
		name, vUrl, avatarSuccess, avatarErr = PostAvatarRequest(nj.AgentEmailID, lkRoomID, nj.MeetingURL, nj.AgentEmailID)
		if avatarSuccess {
			if name != "" {
				dn = name
			}
			generatedVideoUrl = vUrl
			logWithTrace(ctx, "[%s][Avatar] Attempt %d/%d SUCCESS — botName=%q videoUrl=%q", label, attempt+1, maxAvatarAttempts, dn, generatedVideoUrl)
			break
		}
		logWithTrace(ctx, "[%s][Avatar] Attempt %d/%d FAILED — error=%v", label, attempt+1, maxAvatarAttempts, avatarErr)
	}
	if !avatarSuccess {
		logWithTrace(ctx, "[%s][Avatar] All %d attempts failed — marking job failed", label, maxAvatarAttempts)
		UpdateJobStatusInDB(nj.ID, "failed")
		*failed = true
		return
	}

	// Prioritize VideoURL from the job request if provided
	finalVideoUrl := nj.VideoURL
	if finalVideoUrl == "" {
		finalVideoUrl = generatedVideoUrl
	}

	//Recall Retry attempt
	recallBackoff := []time.Duration{0, 5 * time.Second, 15 * time.Second}
	maxRecallAttempts := 3
	var recallBotId string
	var err error

	for attempt := 0; attempt < maxRecallAttempts; attempt++ {
		if recallBackoff[attempt] > 0 {
			logWithTrace(ctx, "[%s][Recall] Retrying — attempt %d/%d — waiting %s", label, attempt+1, maxRecallAttempts, recallBackoff[attempt])
			UpdateJobStatusInDB(nj.ID, "retrying")
			time.Sleep(recallBackoff[attempt])
		}
		logWithTrace(ctx, "[%s][Recall] Attempt %d/%d — sending bot to meeting", label, attempt+1, maxRecallAttempts)
		recallBotId, err = PostRecallRequestV2(nj.MeetingURL, dn, lkRoomID, finalVideoUrl)
		if recallBotId != "" {
			logWithTrace(ctx, "[%s][Recall] Attempt %d/%d SUCCESS — bot created id=%s", label, attempt+1, maxRecallAttempts, recallBotId)
			UpdateJobStatusInDB(nj.ID, "scheduled")
			break
		}
		logWithTrace(ctx, "[%s][Recall] Attempt %d/%d FAILED — error=%v", label, attempt+1, maxRecallAttempts, err)
	}

	if recallBotId == "" {
		logWithTrace(ctx, "[%s][Recall] All %d attempts failed — marking job failed", label, maxRecallAttempts)
		UpdateJobStatusInDB(nj.ID, "failed")
		*failed = true
		return
	}

	// Push recallBotId into LiveKit room metadata so the agent can receive it
	// via the room metadata_changed event and use it to kick the bot on max_call_duration.
	go PatchRecallBotIdToRoom(ctx, lkRoomID, recallBotId)
}

// scheduleCronJob registers a pure cron job (used for the 2nd+ occurrences of a recurring meeting).
func scheduleCronJob(ctx context.Context, nj ScheduledJob) error {
	cronRegistryMu.Lock()
	if _, exists := CronJobRegistry[nj.ID]; exists {
		cronRegistryMu.Unlock()
		log.Printf("[ScheduleJob:Cron] Cron already exists for job %s — skipping", nj.ID)
		return nil
	}
	cronRegistryMu.Unlock()

	cronJob, err := Scheduler.NewJob(
		gocron.CronJob(nj.Cron, false),
		gocron.NewTask(func(job ScheduledJob) error {
			// Create a fresh DB record for this occurrence so every
			// recurrence is visible in the scheduled-jobs list.
			occurrenceJob := job
			occurrenceJob.ID = uuid.New().String()
			occurrenceJob.StartTime = time.Now().UTC().Format(time.RFC3339)
			occurrenceJob.Status = "processing"
			occurrenceJob.CreatedAt = time.Now().UTC().Format(time.RFC3339)
			occurrenceJob.Cron = "" // occurrence records are one-time; omit the series cron
			if err := AddJobToDB(ctx, occurrenceJob); err != nil {
				log.Printf("[ScheduleJob:Cron] WARNING: could not persist occurrence to DB: %v", err)
			}
			scheduleMu.Lock()
			ScheduledJobs = append(ScheduledJobs, occurrenceJob)
			scheduleMu.Unlock()

			// Guard: Only join if this occurrence is the intended next run according to the DB.
			// If the series has been advanced (skipped) in the DB, skip this execution.
			if updatedJob, err := GetJobFromDBByID(nj.ID); err == nil && updatedJob.ID != "" {
				if st, err := time.Parse(time.RFC3339, updatedJob.StartTime); err == nil {
					if time.Now().UTC().Before(st.Add(-5 * time.Minute)) {
						log.Printf("[ScheduleJob:Cron] Skipping occurrence for job %s — series advanced to %s in DB", nj.ID, updatedJob.StartTime)
						return nil
					}
				}
			}

			// Use a per-invocation flag so a failure in one occurrence
			// does not permanently poison subsequent occurrences.
			var localFailed bool
			joinMeeting(ctx, occurrenceJob, "Cron", &localFailed)

			// Mark the occurrence record completed (or leave as failed).
			if !localFailed {
				UpdateJobStatusInDB(occurrenceJob.ID, "completed")
			}
			// Remove occurrence from in-memory list — it has fired.
			scheduleMu.Lock()
			ScheduledJobs = RemoveScheduleJob(ScheduledJobs, occurrenceJob.ID)
			scheduleMu.Unlock()
			return nil
		}, nj),
		gocron.WithName(nj.Name+"-recurring"),
		gocron.WithTags(nj.ID),
		gocron.WithEventListeners(
			gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
				layout := "2006-01-02T15:04:05Z07:00"
				exp, err := time.Parse(layout, nj.Expiry)
				if err != nil {
					log.Printf("[ScheduleJob:Cron] Could not parse expiry %q: %v", nj.Expiry, err)
					return
				}

				// Update the parent series record's start_time to the next cron run.
				if sched, err := cron.ParseStandard(nj.Cron); err == nil {
					if nextRun := sched.Next(time.Now().UTC()); !nextRun.IsZero() {
						UpdateJobStartTimeInDB(nj.ID, nextRun.Format(time.RFC3339))
						log.Printf("[ScheduleJob:Cron] Parent job %s next_run updated to %s", nj.ID, nextRun.Format(time.RFC3339))
					}
				}

				if time.Now().UTC().After(exp) {
					log.Printf("[ScheduleJob:Cron] Job expired — removing CronjobID=%s", jobID)
					Scheduler.RemoveJob(jobID)
					ScheduledJobs = RemoveScheduleJob(ScheduledJobs, jobID.String())
					cronRegistryMu.Lock()
					delete(CronJobRegistry, nj.ID)
					cronRegistryMu.Unlock()
					UpdateJobStatusInDB(nj.ID, "expired")
				}
			}),
		),
	)
	if err != nil {
		return err
	}
	// Store cron job ID in registry.
	cronRegistryMu.Lock()
	CronJobRegistry[nj.ID] = cronJob.ID()
	cronRegistryMu.Unlock()
	log.Printf("[ScheduleJob:Cron] Registered — cronJobID=%s parentJobID=%s cron=%q", cronJob.ID(), nj.ID, nj.Cron)

	// Immediately update the parent series record's start_time to the next cron run.
	// Use robfig/cron directly — gocron's NextRun() may not be populated yet on the
	// first tick, causing a silent no-op.
	if sched, err := cron.ParseStandard(nj.Cron); err == nil {
		nextRun := sched.Next(time.Now().UTC())
		if !nextRun.IsZero() {
			UpdateJobStartTimeInDB(nj.ID, nextRun.Format(time.RFC3339))
			log.Printf("[ScheduleJob:Cron] Parent job %s start_time set to next cron run %s", nj.ID, nextRun.Format(time.RFC3339))
		}
	} else {
		log.Printf("[ScheduleJob:Cron] WARNING: could not parse cron %q for next-run update: %v", nj.Cron, err)
	}

	return nil
}

// checks the in-memory registry and removes the cron job from gocron
func cancelCronJobIfExists(parentJobID string) {
	cronRegistryMu.Lock()
	defer cronRegistryMu.Unlock()
	cronUUID, exists := CronJobRegistry[parentJobID]
	if !exists {
		log.Printf("[CancelCron] No cron job registered for parentJobID=%s", parentJobID)
		return
	}
	if err := Scheduler.RemoveJob(cronUUID); err != nil {
		log.Printf("[CancelCron] gocron RemoveJob failed for parentJobID=%s: %v", parentJobID, err)
		// Fallback: try removing by tag
		Scheduler.RemoveByTags(parentJobID)
		log.Printf("[CancelCron] Cron job removal attempted by tag for parentJobID=%s", parentJobID)
	} else {
		log.Printf("[CancelCron] Cron job %s removed for parentJobID=%s", cronUUID, parentJobID)
	}
	delete(CronJobRegistry, parentJobID)
}

// Method to schedule a new job
func ScheduleJob(ctx context.Context, newJob ScheduledJob) (uuid.UUID, error) {
	// gocron empty string name set
	if strings.TrimSpace(newJob.Name) == "" {
		idSuffix := newJob.ID
		if len(idSuffix) > 8 {
			idSuffix = idSuffix[:8]
		}
		newJob.Name = "meeting-" + idSuffix
		logWithTrace(ctx, "[ScheduleJob] Name was empty, using fallback: %q", newJob.Name)
	}

	isRecurring := len(strings.TrimSpace(newJob.Cron)) > 3

	if isRecurring && strings.TrimSpace(newJob.Cron) == "" {
		return uuid.New(), fmt.Errorf("cron is empty for recurring job")
	}

	// Determine the first-fire time (applies to both one-time and recurring).
	var oneTimeStart gocron.OneTimeJobStartAtOption
	if newJob.StartTime != "" {
		t, err := time.Parse(time.RFC3339, newJob.StartTime)
		if err == nil && t.After(time.Now().UTC().Add(10*time.Second)) {
			logWithTrace(ctx, "[ScheduleJob] First fire at %s UTC (in %s)", t.Format(time.RFC3339), t.Sub(time.Now().UTC()).Round(time.Second))
			oneTimeStart = gocron.OneTimeJobStartDateTime(t)
		} else {
			if newJob.Status == "processing" || !isRecurring {
				logWithTrace(ctx, "[ScheduleJob] Start time %q is in the past (or happening now) — firing immediately", newJob.StartTime)
				oneTimeStart = gocron.OneTimeJobStartImmediately()
			} else {
				logWithTrace(ctx, "[ScheduleJob] Recurring job start time %q is in the past — skipping immediate bootstrap", newJob.StartTime)
			}
		}
	} else {
		logWithTrace(ctx, "[ScheduleJob] No start time — firing immediately")
		oneTimeStart = gocron.OneTimeJobStartImmediately()
	}
	jobFailed := false
	var gocronJobID uuid.UUID
	if oneTimeStart != nil {
		// Create a new One-Time Job
		j, err := Scheduler.NewJob(
			gocron.OneTimeJob(oneTimeStart),
			gocron.NewTask(func(nj ScheduledJob) error {
				joinMeeting(ctx, nj, "OneTime", &jobFailed) // Fire the bot for the first (or only) occurrence
				return nil
			}, newJob),
			gocron.WithName(newJob.Name),
			gocron.WithEventListeners(
				gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
					if !isRecurring {
						// One-time job: clean up after it fires.
						ScheduledJobs = RemoveScheduleJob(ScheduledJobs, jobID.String())
						if jobFailed {
							logWithTrace(ctx, "[AfterJobRuns] Job %s already failed ", jobID)
						} else {
							UpdateJobStatusInDB(jobID.String(), "completed")
							logWithTrace(ctx, "[AfterJobRuns] Job %s marked completed", jobID)
						}
					} else {
						// Recurring meeting: the bootstrap one-time job just handled the
						// first occurrence. Write a completed occurrence record so users can
						// track the full completion history (same pattern as scheduleCronJob).
						if !jobFailed {
							occurrenceJob := newJob
							occurrenceJob.ID = uuid.New().String()
							occurrenceJob.StartTime = time.Now().UTC().Format(time.RFC3339)
							occurrenceJob.Status = "processing"
							occurrenceJob.CreatedAt = time.Now().UTC().Format(time.RFC3339)
							occurrenceJob.Cron = "" // occurrence records are one-time; omit the series cron
							if err := AddJobToDB(ctx, occurrenceJob); err != nil {
								log.Printf("[ScheduleJob:OneTime] WARNING: could not persist occurrence to DB: %v", err)
							}
							scheduleMu.Lock()
							ScheduledJobs = append(ScheduledJobs, occurrenceJob)
							scheduleMu.Unlock()

							// Note: joinMeeting already happened in Task, here we just mark it completed.
							// Wait, the AfterJobRuns for one-time recurring doesn't actually call joinMeeting again.
							UpdateJobStatusInDB(occurrenceJob.ID, "completed")
						}
					}
				}),
			),
		)
		if err != nil {
			return uuid.New(), err
		}
		gocronJobID = j.ID()
		logWithTrace(ctx, "[ScheduleJob] Registered one-time bootstrap job id=%s recurring=%v", j.ID(), isRecurring)
	}

	if isRecurring {
		logWithTrace(ctx, "[ScheduleJob] Registering cron immediately (no bootstrap dependency)")

		// IMPORTANT: use newJob (not nj from task)
		if err := scheduleCronJob(ctx, newJob); err != nil {
			logWithTrace(ctx, "[ScheduleJob] ERROR registering cron job: %v", err)
		} else {
			logWithTrace(ctx, "[ScheduleJob] Cron job registered successfully")
		}
	}

	if oneTimeStart != nil {
		return gocronJobID, nil
	}
	return uuid.Parse(newJob.ID)
}

var ScheduledJobs []ScheduledJob
var Scheduler gocron.Scheduler
var scheduleMu sync.Mutex
var CronJobRegistry = make(map[string]uuid.UUID) //for storing recurring jobs gocron id
var cronRegistryMu sync.Mutex

func InitRecallScheduler() {
	ScheduledJobs = []ScheduledJob{}

	if err := EnsureSchema(); err != nil {
		log.Fatalf("[InitRecallScheduler] schema init failed: %v", err)
	}

	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("[InitRecallScheduler] failed to create scheduler: %v", err)
	}
	Scheduler = s
	Scheduler.Start()
	log.Println("[InitRecallScheduler] Scheduler started")

	pendingJobs, err := GetAllJobsFromDB([]string{})
	if err != nil {
		log.Printf("[InitRecallScheduler] could not read jobs from DB: %v", err)
		return
	}

	rehydrated, skipped := 0, 0
	for _, j := range pendingJobs {
		if j.Status == "processing" || j.Status == "retrying" {
			log.Printf("[Rehydration] Job %s was %s when server died — resetting to scheduled", j.ID, j.Status)
			UpdateJobStatusInDB(j.ID, "scheduled")
			j.Status = "scheduled"
		}
		if j.Status != "scheduled" {
			skipped++
			continue
		}

		// Recurring jobs (with a cron expression) are never expired by start_time staleness —
		// the cron itself controls when the next fire happens.
		isRecurringJob := strings.TrimSpace(j.Cron) != ""

		if !isRecurringJob && j.StartTime != "" {
			t, err := time.Parse(time.RFC3339, j.StartTime)
			if err == nil && t.Before(time.Now().UTC().Add(-2*time.Hour)) {
				log.Printf("[Rehydration] Job %s start time %s is >2h in the past — marking expired", j.ID, j.StartTime)
				UpdateJobStatusInDB(j.ID, "expired")
				skipped++
				continue
			}
		}

		// For recurring jobs, refresh start_time to the next cron occurrence before scheduling.
		if isRecurringJob {
			j = enrichRecurringJob(j)
		}

		ctx := context.WithValue(context.Background(), traceKey, "R-boot")
		if _, err := ScheduleJob(ctx, j); err != nil {
			log.Printf("[Rehydration] Failed to rehydrate job %s: %v", j.ID, err)
			skipped++
		} else {
			ScheduledJobs = append(ScheduledJobs, j)
			rehydrated++
			log.Printf("[Rehydration] Re-registered job %s name=%q fireAt=%s", j.ID, j.Name, j.StartTime)
		}
	}
	log.Printf("[Rehydration] Complete — rehydrated=%d skipped=%d total=%d", rehydrated, skipped, len(pendingJobs))
}
