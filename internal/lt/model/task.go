package model

import "encoding/json"

type LtTask struct {
	TaskID string          `json:"taskid"`
	Task   json.RawMessage `json:"task"`
}

type TaskItem struct {
	LtGm        *LtGm        `json:"lt_gm,omitempty"`
	LtPm        *LtPm        `json:"lt_pm,omitempty"`
	LtAllGroups *LtAllGroups `json:"lt_all_groups,omitempty"`
	LtMembers   *LtMembers   `json:"lt_members,omitempty"`
	LtCron      *LtCron      `json:"lt_cron,omitempty"`
	LtAdmins    *LtAdmins    `json:"lt_admins,omitempty"`
	LtGroups    *LtGroups    `json:"lt_groups,omitempty"`
	LtUpdate    []string     `json:"lt_update,omitempty"`
	LtNotify    *LtNotify    `json:"lt_notify,omitempty"`
}

type LtGm struct {
	Ltid string `json:"ltid"`
}

type LtPm struct {
	Ltid string `json:"ltid"`
}

type LtAllGroups struct {
	Ltid string `json:"ltid"`
}

type LtMembers struct {
	Ltid   string `json:"ltid"`
	Groups string `json:"groups"` // group1,group2,...
}

type LtCron struct {
	Ltid string `json:"ltid"`
	Cron string `json:"cron"`
}

type LtAdmins struct {
	Ltid   string `json:"ltid"`
	Admins string `json:"admins"` // admin1,admin2,...
}

type LtGroups struct {
	Ltid   string      `json:"ltid"`
	Groups []LtGroupIn `json:"groups"`
}

type LtGroupIn struct {
	Name   string `json:"name"`
	Cursor int64  `json:"cursor"`
}

type LtNotify struct {
	Ltid  string `json:"ltid"`
	Error string `json:"error"`
}

// LtTaskResult 上传任务结果
type LtTaskResult struct {
}
