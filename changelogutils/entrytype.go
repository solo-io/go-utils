package changelogutils

import (
	"encoding/json"
	"fmt"
)

type ChangelogEntryType int

const (
	BREAKING_CHANGE ChangelogEntryType = iota
	FIX
	NEW_FEATURE
	NON_USER_FACING
	DEPENDENCY_BUMP
	HELM
	UPGRADE
)

var (
	_ChangelogEntryTypeToValue = map[string]ChangelogEntryType{
		"BREAKING_CHANGE": BREAKING_CHANGE,
		"FIX":             FIX,
		"NEW_FEATURE":     NEW_FEATURE,
		"NON_USER_FACING": NON_USER_FACING,
		"DEPENDENCY_BUMP": DEPENDENCY_BUMP,
		"HELM":            HELM,
		"UPGRADE":         UPGRADE,
	}

	_ChangelogEntryValueToType = map[ChangelogEntryType]string{
		BREAKING_CHANGE: "BREAKING_CHANGE",
		FIX:             "FIX",
		NEW_FEATURE:     "NEW_FEATURE",
		NON_USER_FACING: "NON_USER_FACING",
		DEPENDENCY_BUMP: "DEPENDENCY_BUMP",
		HELM:            "HELM",
		UPGRADE:         "UPGRADE",
	}
)

func (clt ChangelogEntryType) String() string {
	return [...]string{"BREAKING_CHANGE", "FIX", "NEW_FEATURE", "NON_USER_FACING", "DEPENDENCY_BUMP", "HELM", "UPGRADE"}[clt]
}

func (clt ChangelogEntryType) BreakingChange() bool {
	return clt == BREAKING_CHANGE
}

func (clt ChangelogEntryType) NewFeature() bool {
	return clt == NEW_FEATURE
}

func (clt ChangelogEntryType) MarshalJSON() ([]byte, error) {
	if s, ok := interface{}(clt).(fmt.Stringer); ok {
		return json.Marshal(s.String())
	}
	s, ok := _ChangelogEntryValueToType[clt]
	if !ok {
		return nil, fmt.Errorf("invalid ChangelogEntry type: %d", clt)
	}
	return json.Marshal(s)
}

func (clt *ChangelogEntryType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("ChangelogEntryType should be a string, got %s", data)
	}
	v, ok := _ChangelogEntryTypeToValue[s]
	if !ok {
		return fmt.Errorf("invalid ChangelogEntryType %q", s)
	}
	*clt = v
	return nil
}
