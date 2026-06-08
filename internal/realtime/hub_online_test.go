package realtime

import (
	"errors"
	"testing"
	"time"

	"zenmind-kanban-server/internal/kanban"
)

func TestBuildDesktopOnlineListGroupsSessionsByDeviceID(t *testing.T) {
	now := time.Date(2026, 6, 8, 9, 0, 0, 0, time.UTC)
	status := kanban.DesktopStatus{
		Online: true,
		Sessions: []kanban.DesktopSessionStatus{
			{
				SessionID:         "session-a",
				DeviceID:          "device-1",
				CurrentUserID:     "user-1",
				CurrentUserName:   "Alice",
				SelectedProjectID: "project-a",
				Capabilities:      []string{"desktop.assistant.listAgents", "kanban.issue.dispatch"},
				LastSeenAt:         now,
			},
			{
				SessionID:         "session-b",
				DeviceID:          "device-1",
				CurrentUserID:     "user-1",
				CurrentUserName:   "Alice",
				SelectedProjectID: "project-a",
				Capabilities:      []string{"desktop.assistant.listAgents", "desktop.automation.sync"},
				LastSeenAt:         now.Add(2 * time.Minute),
			},
			{
				SessionID:         "session-c",
				CurrentUserName:   "Fallback",
				SelectedProjectID: "default",
				Capabilities:      []string{"desktop.assistant.listAgents"},
				LastSeenAt:         now.Add(time.Minute),
			},
		},
	}
	agents := map[string]desktopAgentLoadResult{
		"session-a": {
			agents: []kanban.DesktopAgentOption{
				{AgentKey: "zenmi", DisplayName: "小宅", Role: "平台总管"},
				{AgentKey: "webOperator", DisplayName: "网驭", Role: "Desktop 内嵌网站操作助手"},
			},
		},
		"session-b": {
			agents: []kanban.DesktopAgentOption{
				{AgentKey: "zenmi", DisplayName: "小宅", Role: "平台总管"},
				{AgentKey: "dailyOfficeProAssistant", DisplayName: "文衡", Role: "办公助手Pro"},
			},
		},
		"session-c": {
			err: errors.New("agent-platform unavailable"),
		},
	}

	result := buildDesktopOnlineList(status, agents)

	if !result.OK || !result.Online {
		t.Fatalf("expected online ok result, got %#v", result)
	}
	if result.SessionCount != 3 {
		t.Fatalf("expected 3 sessions, got %d", result.SessionCount)
	}
	if result.AgentCount != 3 {
		t.Fatalf("expected 3 unique agents, got %d", result.AgentCount)
	}
	if len(result.Devices) != 2 {
		t.Fatalf("expected 2 devices, got %#v", result.Devices)
	}

	device := result.Devices[0]
	if device.DeviceID != "device-1" {
		t.Fatalf("expected first device to be device-1, got %q", device.DeviceID)
	}
	if len(device.Sessions) != 2 {
		t.Fatalf("expected device-1 to contain 2 sessions, got %#v", device.Sessions)
	}
	if device.LastSeenAt != now.Add(2*time.Minute) {
		t.Fatalf("expected latest lastSeenAt, got %s", device.LastSeenAt)
	}
	if len(device.Capabilities) != 3 {
		t.Fatalf("expected merged capabilities, got %#v", device.Capabilities)
	}
	if len(device.Agents) != 3 {
		t.Fatalf("expected deduped agents, got %#v", device.Agents)
	}

	fallback := result.Devices[1]
	if fallback.DeviceID != "session:session-c" {
		t.Fatalf("expected fallback device id, got %q", fallback.DeviceID)
	}
	if fallback.AgentError != "agent-platform unavailable" {
		t.Fatalf("expected per-device agent error, got %q", fallback.AgentError)
	}
}
