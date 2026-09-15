package osv

import (
	"context"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/sirupsen/logrus"
)

// severityRanks orders GitHub advisory severities, MEDIUM is accepted as an alias of MODERATE.
var severityRanks = map[string]int{
	"LOW":      1,
	"MODERATE": 2,
	"MEDIUM":   2,
	"HIGH":     3,
	"CRITICAL": 4,
}

// severityRank returns the rank of a severity, or 0 when it is unknown.
func severityRank(severity string) int {
	return severityRanks[strings.ToUpper(strings.TrimSpace(severity))]
}

// vulnGroup is a single vulnerability, merging the OSV records that are aliases of each other,
// such as a GHSA advisory and its PYSEC twin.
type vulnGroup struct {
	// ID is the lowest record ID of the group.
	ID string
	// Aliases holds every other ID and alias of the group.
	Aliases []string
	Summary string
	// Severity is the highest severity of the group records, empty when unknown.
	Severity string
	// Fixed holds the versions fixing the vulnerability for the queried package.
	Fixed []string
}

// String describes the vulnerability on a single line.
func (g vulnGroup) String() string {
	var sb strings.Builder

	sb.WriteString(g.ID)
	if len(g.Aliases) > 0 {
		fmt.Fprintf(&sb, " (%s)", strings.Join(g.Aliases, ", "))
	}
	if g.Severity != "" {
		fmt.Fprintf(&sb, " [%s]", g.Severity)
	}
	if g.Summary != "" {
		fmt.Fprintf(&sb, " %s", g.Summary)
	}
	if len(g.Fixed) > 0 {
		fmt.Fprintf(&sb, " - fixed in: %s", strings.Join(g.Fixed, ", "))
	}

	return sb.String()
}

// knownVulnerabilities returns the vulnerabilities affecting the given version of the package,
// once the ignore and minseverity filters are applied.
func (o *Osv) knownVulnerabilities(ctx context.Context, version string) ([]vulnGroup, error) {
	vulns, err := o.query(ctx, version)
	if err != nil {
		return nil, err
	}

	return o.filter(o.groupVulnerabilities(vulns)), nil
}

// groupVulnerabilities merges records that are aliases of each other and drops withdrawn ones.
func (o *Osv) groupVulnerabilities(vulns []vulnerability) []vulnGroup {
	parent := map[string]string{}
	find := func(id string) string {
		if _, ok := parent[id]; !ok {
			parent[id] = id
		}
		for parent[id] != id {
			parent[id] = parent[parent[id]]
			id = parent[id]
		}
		return id
	}

	var active []vulnerability
	for _, v := range vulns {
		if v.Withdrawn != "" {
			logrus.Debugf("skipping withdrawn vulnerability %q", v.ID)
			continue
		}
		active = append(active, v)

		root := find(v.ID)
		for _, alias := range v.Aliases {
			if aliasRoot := find(alias); aliasRoot != root {
				parent[aliasRoot] = root
			}
		}
	}

	members := map[string][]vulnerability{}
	for _, v := range active {
		root := find(v.ID)
		members[root] = append(members[root], v)
	}

	groups := make([]vulnGroup, 0, len(members))
	for _, records := range members {
		slices.SortFunc(records, func(a, b vulnerability) int { return strings.Compare(a.ID, b.ID) })

		group := vulnGroup{ID: records[0].ID}
		names := map[string]struct{}{}
		fixed := map[string]struct{}{}

		for _, record := range records {
			names[record.ID] = struct{}{}
			for _, alias := range record.Aliases {
				names[alias] = struct{}{}
			}

			if group.Summary == "" {
				group.Summary = record.Summary
			}

			if severity := record.severity(); severityRank(severity) > severityRank(group.Severity) {
				group.Severity = strings.ToUpper(severity)
			}

			for _, a := range record.Affected {
				if !o.matchesPackage(a.Package) {
					continue
				}
				for _, r := range a.Ranges {
					// GIT ranges are expressed in commits, not versions.
					if r.Type == "GIT" {
						continue
					}
					for _, event := range r.Events {
						if event.Fixed != "" {
							fixed[event.Fixed] = struct{}{}
						}
					}
				}
			}
		}

		delete(names, group.ID)
		group.Aliases = slices.Sorted(maps.Keys(names))
		group.Fixed = slices.Sorted(maps.Keys(fixed))
		groups = append(groups, group)
	}

	slices.SortFunc(groups, func(a, b vulnGroup) int { return strings.Compare(a.ID, b.ID) })

	return groups
}

// filter drops the ignored vulnerabilities and those below the minimum severity.
func (o *Osv) filter(groups []vulnGroup) []vulnGroup {
	var result []vulnGroup

	for _, group := range groups {
		if o.isIgnored(group) {
			logrus.Debugf("ignoring vulnerability %q", group.ID)
			continue
		}

		if o.minSeverity > 0 {
			rank := severityRank(group.Severity)
			switch {
			case rank == 0:
				logrus.Debugf("vulnerability %q has an unknown severity, accounting for it", group.ID)
			case rank < o.minSeverity:
				logrus.Debugf("ignoring vulnerability %q with severity %s, below %s", group.ID, group.Severity, strings.ToUpper(o.spec.MinSeverity))
				continue
			}
		}

		result = append(result, group)
	}

	return result
}

// isIgnored returns true when the vulnerability ID or any of its aliases is ignored.
func (o *Osv) isIgnored(group vulnGroup) bool {
	if _, ok := o.ignore[strings.ToUpper(group.ID)]; ok {
		return true
	}
	for _, alias := range group.Aliases {
		if _, ok := o.ignore[strings.ToUpper(alias)]; ok {
			return true
		}
	}
	return false
}

// pypiNameSeparators matches the separators PEP 503 normalizes package names on.
var pypiNameSeparators = regexp.MustCompile(`[-_.]+`)

// matchesPackage returns true when an affected package of a record is the queried package.
// Records may list other packages, such as the Go standard library next to golang.org/x/net.
func (o *Osv) matchesPackage(p osvPackage) bool {
	// A package URL type without known OSV ecosystem can only be matched on the package URL itself.
	if o.ecosystem == "" {
		purl, _, _ := strings.Cut(p.Purl, "@")
		return purl != "" && purl == o.spec.Purl
	}

	// Some ecosystems carry a release suffix, such as "Debian:12".
	if p.Ecosystem != o.ecosystem && !strings.HasPrefix(p.Ecosystem, o.ecosystem+":") {
		return false
	}

	if o.ecosystem == ecosystemPyPI {
		return strings.EqualFold(
			pypiNameSeparators.ReplaceAllString(p.Name, "-"),
			pypiNameSeparators.ReplaceAllString(o.name, "-"),
		)
	}

	return p.Name == o.name
}
