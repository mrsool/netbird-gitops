package data

// PostureCheck holds NetBird PostureCheck object
type PostureCheck struct {
	ID          string              `json:"id"`
	Name        string              `yaml:"name" json:"name"`
	Description string              `yaml:"description" json:"description"`
	Checks      PostureCheckDetails `yaml:"checks" json:"checks"`
}

// PostureCheckDetails different checks in posture check
type PostureCheckDetails struct {
	NBVersionCheck        MinVersionDescriptor     `yaml:"nb_version_check" json:"nb_version_check"`
	OSVersionCheck        OSVersionCheckObj        `yaml:"os_version_check" json:"os_version_check"`
	GeoLocationCheck      GeoLocationCheckObj      `yaml:"geo_location_check" json:"geo_location_check"`
	PeerNetworkRangeCheck PeerNetworkRangeCheckObj `yaml:"peer_network_range_check" json:"peer_network_range_check"`
	ProcessCheck          ProcessCheckObj          `yaml:"process_check" json:"process_check"`
}

// OSVersionCheckObj Different OS types version checks
type OSVersionCheckObj struct {
	Android MinVersionDescriptor       `yaml:"android" json:"android"`
	IOS     MinVersionDescriptor       `yaml:"ios" json:"ios"`
	Darwin  MinVersionDescriptor       `yaml:"darwin" json:"darwin"`
	Linux   MinKernelVersionDescriptor `yaml:"linux" json:"linux"`
	Windows MinKernelVersionDescriptor `yaml:"windows" json:"windows"`
}

// MinVersionDescriptor descriptor for generic min version
type MinVersionDescriptor struct {
	MinVersion string `yaml:"min_version" json:"min_version"`
}

// MinKernelVersionDescriptor descriptor for kernel min version
type MinKernelVersionDescriptor struct {
	MinKernelVersion string `yaml:"min_kernel_version" json:"min_kernel_version"`
}

// GeoLocationCheckObj posture check geo location check
type GeoLocationCheckObj struct {
	Locations []GeoLocation `yaml:"locations" json:"locations"`
}

// GeoLocation descriptor for a geolocation
type GeoLocation struct {
	CountryCode string `yaml:"country_code" json:"country_code"`
	CityName    string `yaml:"city_name" json:"city_name"`
}

// PeerNetworkRangeCheckObj posture check network range check
type PeerNetworkRangeCheckObj struct {
	Ranges []string `yaml:"ranges" json:"ranges"`
}

// ProcessCheckObj posture check process checklist
type ProcessCheckObj struct {
	Processes []OSProcess `yaml:"processes" json:"processes"`
}

// OSProcess posture check for different paths for OS
type OSProcess struct {
	LinuxPath   string `yaml:"linux_path" json:"linux_path"`
	MacPath     string `yaml:"mac_path" json:"mac_path"`
	WindowsPath string `yaml:"windows_path" json:"windows_path"`
}
