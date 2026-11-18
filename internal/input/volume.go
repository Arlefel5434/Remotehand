package input

type VolumeState struct {
	Level int  `json:"level"`
	Muted bool `json:"muted"`
}

// UpdateVolume reads the endpoint when delta is zero and muted is nil.
type VolumeController interface {
	UpdateVolume(delta int, muted *bool) (VolumeState, error)
}

type mockDevice struct {
	namedAdapter
	state VolumeState
}

func (m *mockDevice) UpdateVolume(delta int, muted *bool) (VolumeState, error) {
	m.state.Level = min(100, max(0, m.state.Level+delta))
	if muted != nil {
		m.state.Muted = *muted
	}
	return m.state, nil
}
