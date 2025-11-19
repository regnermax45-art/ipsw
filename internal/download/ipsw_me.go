package download

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const ipswMeAPI = "https://api.ipsw.me/v4/"

// Device struct
type Device struct {
	Name        string `json:"name,omitempty"`
	Identifier  string `json:"identifier,omitempty"`
	BoardConfig string `json:"boardconfig,omitempty"`
	Platform    string `json:"platform,omitempty"`
	CpID        int    `json:"cpid,omitempty"`
	BdID        int    `json:"bdid,omitempty"`
	Firmwares   []IPSW `json:"firmwares,omitempty"`
}

// IPSW struct
type IPSW struct {
	Identifier  string    `json:"identifier,omitempty"`
	Version     string    `json:"version,omitempty"`
	BuildID     string    `json:"buildid,omitempty"`
	SHA1        string    `json:"sha1sum,omitempty"`
	MD5         string    `json:"md5sum,omitempty"`
	FileSize    int       `json:"filesize,omitempty"`
	URL         string    `json:"url,omitempty"`
	ReleaseDate time.Time `json:"releasedate"`
	UploadDate  time.Time `json:"uploaddate"`
	Signed      bool      `json:"signed,omitempty"`
}

// iPhone SE2/SE3 device identifiers
const (
	iPhoneSE2Identifier = "iPhone12,8" // iPhone SE (2nd generation)
	iPhoneSE3Identifier = "iPhone14,6" // iPhone SE (3rd generation)
)

// DeviceMapping represents a mapping between SE2 and SE3
type DeviceMapping struct {
	SE2Identifier string `json:"se2_identifier"`
	SE3Identifier string `json:"se3_identifier"`
	Compatible    bool   `json:"compatible"`
}

// GetAllDevices returns a list of all devices
func GetAllDevices() ([]Device, error) {
	devices := []Device{}

	res, err := http.Get(ipswMeAPI + "devices")
	if err != nil {
		return devices, err
	}
	defer res.Body.Close()
	
	if res.StatusCode != http.StatusOK {
		return devices, fmt.Errorf("api returned status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return devices, err
	}

	err = json.Unmarshal(body, &devices)
	if err != nil {
		return devices, err
	}

	return devices, nil
}

// GetDevice returns a device from its identifier
func GetDevice(identifier string) (Device, error) {
	d := Device{}

	res, err := http.Get(ipswMeAPI + "device/" + identifier)
	if err != nil {
		return d, err
	}
	defer res.Body.Close()
	
	if res.StatusCode != http.StatusOK {
		return d, fmt.Errorf("api returned status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return d, err
	}

	err = json.Unmarshal(body, &d)
	if err != nil {
		return d, err
	}

	return d, nil
}

// GetDeviceIPSWs returns a device's IPSWs from its identifier
func GetDeviceIPSWs(identifier string) ([]IPSW, error) {
	d, err := GetDevice(identifier)
	if err != nil {
		return nil, err
	}
	return d.Firmwares, nil
}

// GetAllIPSW finds all IPSW files for a given iOS version
func GetAllIPSW(version string) ([]IPSW, error) {
	ipsws := []IPSW{}

	res, err := http.Get(ipswMeAPI + "ipsw/" + version)
	if err != nil {
		return ipsws, err
	}
	defer res.Body.Close()
	
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return ipsws, err
	}

	err = json.Unmarshal(body, &ipsws)
	if err != nil {
		return ipsws, err
	}

	return ipsws, nil
}

// GetIPSW will get an IPSW when supplied an identifier and build ID
func GetIPSW(identifier, buildID string) (IPSW, error) {
	i := IPSW{}

	res, err := http.Get(ipswMeAPI + "ipsw/" + identifier + "/" + buildID)
	if err != nil {
		return i, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return i, fmt.Errorf("api returned status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return i, err
	}

	err = json.Unmarshal(body, &i)
	if err != nil {
		return i, err
	}

	return i, nil
}

// GetVersion returns the iOS version for a given build ID
func GetVersion(buildID string) (string, error) {
	devices, err := GetAllDevices()
	if err != nil {
		return "", fmt.Errorf("failed to get all devices from ipsw.me API: %v", err)
	}

	for i := len(devices) - 1; i >= 0; i-- {
		var dev Device
		res, err := http.Get(ipswMeAPI + "device/" + devices[i].Identifier)
		if err != nil {
			continue // Skip on error and try next device
		}
		
		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			continue
		}

		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			continue
		}

		err = json.Unmarshal(body, &dev)
		if err != nil {
			continue
		}

		for _, ipsw := range dev.Firmwares {
			if ipsw.BuildID == buildID {
				return ipsw.Version, nil
			}
		}
	}

	return "", fmt.Errorf("build did not match a version in the ipsw.me API")
}

// GetBuildID returns the BuildID for a given version and identifier
func GetBuildID(version, identifier string) (string, error) {
	var ipsws []IPSW

	res, err := http.Get(ipswMeAPI + "ipsw/" + version)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api returned status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &ipsws)
	if err != nil {
		return "", err
	}

	for _, i := range ipsws {
		if i.Identifier == identifier {
			return i.BuildID, nil
		}
	}
	return "", fmt.Errorf("no build found for version %s and device %s", version, identifier)
}

// GetSE2ToSE3Mapping returns the device mapping between SE2 and SE3
func GetSE2ToSE3Mapping() DeviceMapping {
	return DeviceMapping{
		SE2Identifier: iPhoneSE2Identifier,
		SE3Identifier: iPhoneSE3Identifier,
		Compatible:    true, // Generally compatible for porting purposes
	}
}

// GetCompatibleIPSWs returns IPSWs that are compatible between SE2 and SE3
func GetCompatibleIPSWs(version string) ([]IPSW, error) {
	compatibleIPSWs := []IPSW{}
	
	// Get SE2 IPSWs
	se2IPSWs, err := GetDeviceIPSWs(iPhoneSE2Identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get SE2 IPSWs: %v", err)
	}
	
	// Get SE3 IPSWs  
	se3IPSWs, err := GetDeviceIPSWs(iPhoneSE3Identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get SE3 IPSWs: %v", err)
	}
	
	// Find compatible versions (same iOS version)
	versionMap := make(map[string]bool)
	for _, ipsw := range se2IPSWs {
		versionMap[ipsw.Version] = true
	}
	
	for _, ipsw := range se3IPSWs {
		if versionMap[ipsw.Version] {
			compatibleIPSWs = append(compatibleIPSWs, ipsw)
		}
	}
	
	return compatibleIPSWs, nil
}

// GetSE3IPSWForSE2Version finds the SE3 IPSW that matches an SE2 iOS version
func GetSE3IPSWForSE2Version(se2Version string) (IPSW, error) {
	se3IPSWs, err := GetDeviceIPSWs(iPhoneSE3Identifier)
	if err != nil {
		return IPSW{}, fmt.Errorf("failed to get SE3 IPSWs: %v", err)
	}
	
	for _, ipsw := range se3IPSWs {
		if ipsw.Version == se2Version {
			return ipsw, nil
		}
	}
	
	return IPSW{}, fmt.Errorf("no SE3 IPSW found for SE2 version %s", se2Version)
}

// IsSE2Device checks if the identifier is iPhone SE2
func IsSE2Device(identifier string) bool {
	return identifier == iPhoneSE2Identifier
}

// IsSE3Device checks if the identifier is iPhone SE3  
func IsSE3Device(identifier string) bool {
	return identifier == iPhoneSE3Identifier
}

// ConvertSE2ToSE3Identifier converts SE2 identifier to SE3 if needed
func ConvertSE2ToSE3Identifier(identifier string) string {
	if identifier == iPhoneSE2Identifier {
		return iPhoneSE3Identifier
	}
	return identifier
}

// Release struct for releases endpoint
type Release struct {
	Version    string    `json:"version"`
	BuildID    string    `json:"buildid"`
	Released   time.Time `json:"released"`
	Beta       bool      `json:"beta"`
	RC         bool      `json:"rc"`
	Signed     bool      `json:"signed"`
	DeviceIDs  []string  `json:"deviceIds"`
}

// GetReleases returns all iOS releases
func GetReleases() ([]Release, error) {
	releases := []Release{}

	res, err := http.Get(ipswMeAPI + "releases")
	if err != nil {
		return releases, err
	}
	defer res.Body.Close()
	
	if res.StatusCode != http.StatusOK {
		return releases, fmt.Errorf("api returned status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return releases, err
	}

	err = json.Unmarshal(body, &releases)
	if err != nil {
		return releases, err
	}

	return releases, nil
}
