package hueapi

import (
	"encoding/xml"
	"fmt"
)

type Device struct {
	XMLName     xml.Name `xml:"root"`
	XMLNS       string   `xml:"xmlns,attr"`
	SpecVersion struct {
		Major int `xml:"major"`
		Minor int `xml:"minor"`
	} `xml:"specVersion"`
	URLBase string `xml:"URLBase"`
	Device  struct {
		DeviceType       string `xml:"deviceType"`
		FriendlyName     string `xml:"friendlyName"`
		Manufacturer     string `xml:"manufacturer"`
		ManufacturerURL  string `xml:"manufacturerURL"`
		ModelDescription string `xml:"modelDescription"`
		ModelName        string `xml:"modelName"`
		ModelNumber      string `xml:"modelNumber"`
		ModelURL         string `xml:"modelURL"`
		SerialNumber     string `xml:"serialNumber"`
		UDN              string `xml:"UDN"`
	} `xml:"device"`
}

func NewDevice(urlBase, friendlyName, serialNumber, uuid string) *Device {
	desc := &Device{}
	desc.XMLNS = "urn:schemas-upnp-org:device-1-0"
	desc.SpecVersion.Major = 1
	desc.SpecVersion.Minor = 0
	desc.URLBase = urlBase
	desc.Device.DeviceType = "urn:schemas-upnp-org:device:Basic:1"
	desc.Device.FriendlyName = friendlyName
	desc.Device.Manufacturer = "Royal Philips Electronics"
	desc.Device.ManufacturerURL = "http://www.philips.com"
	desc.Device.ModelDescription = "Philips hue Personal Wireless Lighting"
	desc.Device.ModelName = "Philips hue bridge 2015"
	desc.Device.ModelNumber = "BSB002"
	desc.Device.ModelURL = "http://www.meethue.com"
	desc.Device.SerialNumber = serialNumber
	desc.Device.UDN = fmt.Sprintf("uuid:%s", uuid)
	return desc
}
