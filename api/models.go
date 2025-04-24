package api

// Statistics about the 4G or 5G connection signals
type SignalStats struct {
	AntennaUsed string   `json:"antennaUsed"` // "Internal_directional"
	Bands       []string `json:"bands"`       // "b66", "n41", etc.
	Bars        float64  `json:"bars"`
	Cid         int      `json:"cid"`
	ENBID       int      `json:"eNBID"` // Only set for 4G
	GNBID       int      `json:"gNBID"` // Only set for 5G
	Rsrp        int      `json:"rsrp"`
	Rsrq        int      `json:"rsrq"`
	Rssi        int      `json:"rssi"`
	Sinr        int      `json:"sinr"`
}

type Device struct {
	FriendlyName    string `json:"friendlyName"`
	HardwareVersion string `json:"hardwareVersion"`
	Index           int    `json:"index"`
	IsEnabled       bool   `json:"isEnabled"`
	IsMeshSupported bool   `json:"isMeshSupported"`
	MacID           string `json:"macId"`
	Manufacturer    string `json:"manufacturer"`
	ManufacturerOUI string `json:"manufacturerOUI"`
	Model           string `json:"model"`
	Name            string `json:"name"`
	Role            string `json:"role"`
	Serial          string `json:"serial"`
	SoftwareVersion string `json:"softwareVersion"`
	Type            string `json:"type"`
	UpdateState     string `json:"updateState"`
}

type Generic struct {
	Apn          string `json:"apn"`
	HasIPv6      bool   `json:"hasIPv6"`
	Registration string `json:"registration"`
	Roaming      bool   `json:"roaming"`
}

type Signal struct {
	FourG   SignalStats `json:"4g"`
	FiveG   SignalStats `json:"5g"`
	Generic Generic     `json:"generic"`
}

type Time struct {
	DaylightSavings struct {
		IsUsed bool `json:"isUsed"`
	} `json:"daylightSavings"`
	LocalTime     int    `json:"localTime"`
	LocalTimeZone string `json:"localTimeZone"`
	UpTime        int    `json:"upTime"`
}

type GatewayResponse struct {
	Device Device `json:"device"`
	Signal Signal `json:"signal"`
	Time   Time   `json:"time"`
}

type authToken struct {
	Expiration       int64  `json:"expiration"`
	RefreshCountLeft int8   `json:"refreshCountLeft"`
	RefreshCountMax  int8   `json:"refreshCountMax"`
	Token            string `json:"token"`
}

type authResponse struct {
	Auth authToken `json:"auth"`
}
