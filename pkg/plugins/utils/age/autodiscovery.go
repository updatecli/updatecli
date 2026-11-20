package age

import "github.com/updatecli/updatecli/pkg/plugins/utils/version"

// AutodiscoverySpec defines the age filter of an autodiscovery crawler.
// It adds to Spec the settings that only make sense when the version in use is known.
type AutodiscoverySpec struct {
	Spec `yaml:",inline" mapstructure:",squash"`
	// "allowdowngrade" defines whether an update can lower the version in use.
	//
	// default: false
	//
	// remark:
	//   * by default, an update never lowers the version in use. While every newer
	//     release is too recent, the update is skipped.
	//   * when set to true, an update lowers the version in use when it doesn't match the
	//     age filter, to the newest release that does.
	//   * it only applies with a "minimum" or "maximum" age.
	//
	AllowDowngrade *bool `yaml:",omitempty"`
}

// IsZero returns true if no setting is defined, "allowdowngrade" included.
func (s AutodiscoverySpec) IsZero() bool {
	return s.Spec.IsZero() && s.AllowDowngrade == nil
}

// DowngradeAllowed returns true if an update can lower the version in use.
func (s AutodiscoverySpec) DowngradeAllowed() bool {
	return s.AllowDowngrade != nil && *s.AllowDowngrade
}

// MayDowngrade returns true if an age filter is defined and allows an update to lower the
// version in use, falling back to the newest release matching the age filter.
func (s AutodiscoverySpec) MayDowngrade() bool {
	return !s.Spec.IsZero() && s.DowngradeAllowed()
}

// NeverDowngrade returns true if an age filter is defined and an update must not lower the
// version in use. Without an age filter, nothing holds back the newest versions, so the version
// filters are left as they are.
func (s AutodiscoverySpec) NeverDowngrade() bool {
	return !s.Spec.IsZero() && !s.DowngradeAllowed()
}

// GreaterThanPattern returns the pattern of filter used to look for a version newer than
// currentVersion. With an age filter, it only accepts a version lower than currentVersion
// when a downgrade is allowed. Without one, it's the pattern of filter.GreaterThanPattern.
func (s AutodiscoverySpec) GreaterThanPattern(filter *version.Filter, currentVersion string) (string, error) {
	switch {
	case s.MayDowngrade():
		return filter.GreaterThanPatternAllowingDowngrade(currentVersion)
	case s.NeverDowngrade():
		return filter.GreaterThanPatternWithoutDowngrade(currentVersion)
	default:
		return filter.GreaterThanPattern(currentVersion)
	}
}
