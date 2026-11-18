package input

type VolumeState struct {
	Level int  `json:"level"`
	Muted bool `json:"muted"`
}

// UpdateVolume reads the endpoint when no change is requested.
type VolumeController interface {
	UpdateVolume(delta int, level *int, muted *bool) (VolumeState, error)
}

type mockDevice struct {
	namedAdapter
	state VolumeState
}

func (m *mockDevice) UpdateVolume(delta int, level *int, muted *bool) (VolumeState, error) {
	m.state.Level = min(100, max(0, m.state.Level+delta))
	if level != nil {
		m.state.Level = *level
	}
	if muted != nil {
		m.state.Muted = *muted
	}
	return m.state, nil
}
