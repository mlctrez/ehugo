package hueapi

import "fmt"

// https://developers.meethue.com/develop/hue-api/lights-api/

type LightState struct {
	On        bool      `json:"on"`
	Bri       uint8     `json:"bri"`
	Hue       uint16    `json:"hue"`
	Sat       uint8     `json:"sat"`
	Effect    string    `json:"effect"`
	XY        []float64 `json:"xy"`
	Ct        uint16    `json:"ct"`
	Alert     string    `json:"alert"`
	ColorMode string    `json:"colormode"`
	Mode      string    `json:"mode"`
	Reachable bool      `json:"reachable"`
}

type LightInfo struct {
	State            LightState `json:"state"`
	Type             string     `json:"type"`
	Name             string     `json:"name"`
	ModelID          string     `json:"modelid"`
	ManufacturerName string     `json:"manufacturername"`
	UniqueID         string     `json:"uniqueid"`
	SWVersion        string     `json:"swversion"`
}

func (l *LightInfo) Defaults(id string) {
	if l.Type == "" {
		l.Type = "Extended color light"
	}
	if l.ModelID == "" {
		l.ModelID = "LCT007"
	}
	if l.ManufacturerName == "" {
		l.ManufacturerName = "Philips"
	}
	if l.UniqueID == "" {
		l.UniqueID = fmt.Sprintf("00:17:88:01:00:bd:c7:b9-%02s", id)
	}
	if l.SWVersion == "" {
		l.SWVersion = "66012040"
	}

	l.State = LightState{
		On:        true,
		Bri:       254,
		Hue:       0,
		Sat:       0,
		Effect:    "none",
		XY:        []float64{0.0, 0.0},
		Ct:        366,
		Alert:     "none",
		ColorMode: "ct",
		Mode:      "homeautomation",
		Reachable: true,
	}

}

type StateChange struct {
	On     *bool     `json:"on,omitempty"`
	Bri    *uint8    `json:"bri,omitempty"`
	Hue    *uint16   `json:"hue,omitempty"`
	Sat    *uint8    `json:"sat,omitempty"`
	Effect *string   `json:"effect,omitempty"`
	XY     []float64 `json:"xy,omitempty"`
	Ct     *uint16   `json:"ct,omitempty"`
	Alert  *string   `json:"alert,omitempty"`
}

func (l *LightInfo) ApplyStateChange(id string, change *StateChange) []map[string]interface{} {
	var response []map[string]interface{}

	basePath := fmt.Sprintf("/lights/%s/state", id)

	if change.On != nil {
		l.State.On = *change.On
		response = append(response, map[string]interface{}{
			"success": map[string]interface{}{basePath + "/on": change.On},
		})
	}

	if change.Bri != nil {
		l.State.Bri = *change.Bri
		response = append(response, map[string]interface{}{
			"success": map[string]interface{}{basePath + "/bri": change.Bri},
		})
	}
	if change.Hue != nil {
		l.State.Hue = *change.Hue
		l.State.ColorMode = "hs"
		response = append(response, map[string]interface{}{
			"success": map[string]interface{}{basePath + "/hue": change.Hue},
		})
	}
	if change.Sat != nil {
		l.State.Sat = *change.Sat
		l.State.ColorMode = "hs"
		response = append(response, map[string]interface{}{
			"success": map[string]interface{}{basePath + "/sat": change.Sat},
		})
	}
	if change.Ct != nil {
		l.State.Ct = *change.Ct
		l.State.ColorMode = "ct"
		response = append(response, map[string]interface{}{
			"success": map[string]interface{}{basePath + "/ct": change.Ct},
		})
	}

	return response
}
