package hueapi

import (
	"github.com/mlctrez/ehugo/ssdp"
	"strings"
)

func (h *HueApi) SSDPCallback(p *ssdp.Packet) {
	var err error
	if err = p.Parse(); err != nil {
		h.logger.Errorf("SSDPCallback parse error: %s", err)
		return
	}

	// limit to one device to avoid so much logging
	if !strings.HasPrefix(p.Client.String(), "10.0.0.45:") {
		return
	}

	if p.Method != "M-SEARCH" || strings.Contains(p.MIMEHeader.Get("St"), "dial-multiscreen-org") {
		return
	}
	h.logger.Infof("SSDPCallback client=%s method=%s headers=%+v", p.Client.String(), p.Method, p.MIMEHeader)

	for _, bridge := range h.bridges {
		if err = p.Reply(bridge); err != nil {
			h.logger.Errorf("ssdpCallback reply error: %s", err)
		}
	}
}
