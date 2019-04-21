package clicore

import (
	"strings"
)

type MockWriteSyncer struct {
	inputs     []string
	syncCount  uint
	inputCount uint
}

func (m *MockWriteSyncer) Write(in []byte) (int, error) {
	m.inputCount++
	m.inputs = append(m.inputs, string(in))
	return len(in), nil
}
func (m *MockWriteSyncer) Sync() error {
	m.syncCount++
	return nil
}
func (m *MockWriteSyncer) Summarize() (string, uint, uint) {
	return strings.Join(m.inputs, "\n"), m.inputCount, m.syncCount
}

type MockTargets struct {
	Stdout  *MockWriteSyncer
	Stderr  *MockWriteSyncer
	FileLog *MockWriteSyncer
}

func NewMockTargets() MockTargets {
	return MockTargets{
		Stdout:  &MockWriteSyncer{},
		Stderr:  &MockWriteSyncer{},
		FileLog: &MockWriteSyncer{},
	}
}
