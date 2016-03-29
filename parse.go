package semver

import (
	"errors"
	"fmt"
	"strings"
)

func rewriteRange(i string) string {
	m := constraintRangeRegex.FindAllStringSubmatch(i, -1)
	if m == nil {
		return i
	}
	o := i
	for _, v := range m {
		t := fmt.Sprintf(">= %s, <= %s", v[1], v[11])
		o = strings.Replace(o, v[0], t, 1)
	}

	return o
}

func parseConstraint(c string) (Constraint, error) {
	m := constraintRegex.FindStringSubmatch(c)
	if m == nil {
		return nil, fmt.Errorf("Malformed constraint: %s", c)
	}

	// Handle the full wildcard case first - easy!
	if isX(m[3]) {
		return any{}, nil
	}

	ver := m[2]
	var wildPatch, wildMinor bool
	if isX(strings.TrimPrefix(m[4], ".")) {
		wildPatch = true
		wildMinor = true
		ver = fmt.Sprintf("%s.0.0%s", m[3], m[6])
	} else if isX(strings.TrimPrefix(m[5], ".")) {
		wildPatch = true
		ver = fmt.Sprintf("%s%s.0%s", m[3], m[4], m[6])
	}

	v, err := NewVersion(ver)
	if err != nil {
		// The constraintRegex should catch any regex parsing errors. So,
		// we should never get here.
		return nil, errors.New("constraint Parser Error")
	}

	switch m[1] {
	case "^":
		// Caret always expands to a range
		return expandCaret(v), nil
	case "~":
		// Tilde always expands to a range
		return expandTilde(v, wildMinor), nil
	case "!=":
		// Not equals expands to a range if no element isX(); otherwise expands
		// to a union of ranges
		return expandNeq(v, wildMinor, wildPatch), nil
	case "", "=":
		if wildPatch || wildMinor {
			// Equalling a wildcard has the same behavior as expanding tilde
			return expandTilde(v, wildMinor), nil
		}
		return v, nil
	case ">":
		return expandGreater(v, false), nil
	case ">=", "=>":
		return expandGreater(v, true), nil
	case "<":
		return expandLess(v, wildMinor, wildPatch, false), nil
	case "<=", "=<":
		return expandLess(v, wildMinor, wildPatch, true), nil
	default:
		// Shouldn't be possible to get here, unless the regex is allowing
		// predicate we don't know about...
		return nil, fmt.Errorf("Unrecognized predicate %q", m[1])
	}
}

func expandCaret(v *Version) Constraint {
	maxv := &Version{
		major: v.major + 1,
		minor: 0,
		patch: 0,
	}

	return rangeConstraint{
		min:        v,
		max:        maxv,
		includeMin: true,
		includeMax: false,
	}
}

func expandTilde(v *Version, wildMinor bool) Constraint {
	if wildMinor {
		// When minor is wild on a tilde, behavior is same as caret
		return expandCaret(v)
	}

	maxv := &Version{
		major: v.major,
		minor: v.minor + 1,
		patch: 0,
	}

	return rangeConstraint{
		min:        v,
		max:        maxv,
		includeMin: true,
		includeMax: false,
	}
}

// expandNeq expands a "not-equals" constraint.
//
// If the constraint has any wildcards, it will expand into a unionConstraint
// (which is how we represent a disjoint set). If there are no wildcards, it
// will expand to a rangeConstraint with no min or max, but having the one
// exception.
func expandNeq(v *Version, wildMinor, wildPatch bool) Constraint {
	if !(wildMinor || wildPatch) {
		return rangeConstraint{
			excl: []*Version{v},
		}
	}

	// Create the low range with no min, and the max as the floor admitted by
	// the wildcard
	lr := rangeConstraint{
		max:        v,
		includeMax: false,
	}

	// The high range uses the derived version (bumped depending on where the
	// wildcards were) as the min, and is inclusive
	minv := &Version{
		major: v.major,
		minor: v.minor,
		patch: v.patch,
	}

	if wildMinor {
		minv.major++
	} else { // TODO should be an else if?
		minv.minor++
	}

	hr := rangeConstraint{
		min:        minv,
		includeMin: true,
	}

	return Union(lr, hr)
}

func expandGreater(v *Version, eq bool) Constraint {
	return rangeConstraint{
		min:        v,
		includeMin: eq,
	}
}

func expandLess(v *Version, wildMinor, wildPatch, eq bool) Constraint {
	v2 := &Version{
		major: v.major,
		minor: v.minor,
		patch: v.patch,
	}
	if wildMinor {
		v2.major++
	} else if wildPatch {
		v2.minor++
	}

	return rangeConstraint{
		max:        v2,
		includeMax: eq,
	}
}

func isX(x string) bool {
	l := strings.ToLower(x)
	return l == "x" || l == "*"
}
