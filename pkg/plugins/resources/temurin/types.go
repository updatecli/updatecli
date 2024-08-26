package temurin

type releaseInformation struct {
	Releases []string `json:"releases"`
}

type apiInfoReleases struct {
	MostRecentLTS            int   `json:"most_recent_lts"`
	MostRecentFeatureRelease int   `json:"most_recent_feature_release"`
	AvailableLTSReleases     []int `json:"available_lts_releases"`
	AvailableReleases        []int `json:"available_releases"`
}

type parsedVersion struct {
	Major    int `json:"major"`
	Minor    int `json:"minor"`
	Security int `json:"security"`
}
