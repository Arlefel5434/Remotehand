package input

import (
	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wca/pkg/wca"
	"math"
	"runtime"
	"unsafe"
)

func (windowsPointer) UpdateVolume(delta int, level *int, muted *bool) (state VolumeState, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err = ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		if result, ok := err.(*ole.OleError); !ok || result.Code() != 1 {
			return
		}
	}
	defer ole.CoUninitialize()
	var devices *wca.IMMDeviceEnumerator
	if err = wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &devices); err != nil {
		return
	}
	defer devices.Release()
	var device *wca.IMMDevice
	if err = devices.GetDefaultAudioEndpoint(wca.ERender, wca.EMultimedia, &device); err != nil {
		return
	}
	defer device.Release()
	var endpoint *wca.IAudioEndpointVolume
	if err = device.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &endpoint); err != nil {
		return
	}
	defer endpoint.Release()
	var scalar float32
	if level != nil {
		if err = endpoint.SetMasterVolumeLevelScalar(float32(*level)/100, nil); err != nil {
			return
		}
	} else if delta != 0 {
		if err = endpoint.GetMasterVolumeLevelScalar(&scalar); err != nil {
			return
		}
		if err = endpoint.SetMasterVolumeLevelScalar(min(1, max(0, scalar+float32(delta)/100)), nil); err != nil {
			return
		}
	}
	if muted != nil {
		if err = endpoint.SetMute(*muted, nil); err != nil {
			return
		}
	}
	if err = endpoint.GetMasterVolumeLevelScalar(&scalar); err != nil {
		return
	}
	// Core Audio writes a four-byte BOOL through the wrapper's *bool parameter.
	var muteFlag uint32
	if err = endpoint.GetMute((*bool)(unsafe.Pointer(&muteFlag))); err != nil {
		return
	}
	return VolumeState{Level: int(math.Round(float64(scalar) * 100)), Muted: muteFlag != 0}, nil
}
