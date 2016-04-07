package semver

import (
	"bytes"
	"fmt"
)

var rangeErrs = [...]string{
	"%s is less than the minimum of %s",
	"%s is less than or equal to the minimum of %s",
	"%s is greater than the maximum of %s",
	"%s is greater than or equal to the maximum of %s",
	"%s is specifically disallowed in %s",
}

const (
	rerrLT = iota
	rerrLTE
	rerrGT
	rerrGTE
	rerrNE
)

type MatchFailure interface {
	error
	Pair() (v *Version, c Constraint)
}

type rangeMatchFailure struct {
	v   *Version
	rc  rangeConstraint
	typ int8
}

func (rce rangeMatchFailure) Error() string {
	return fmt.Sprintf(rangeErrs[rce.typ], rce.v, rce.rc)
}

func (rce rangeMatchFailure) Pair() (v *Version, r Constraint) {
	return rce.v, rce.rc
}

type versionMatchFailure struct {
	v, other *Version
}

func (vce versionMatchFailure) Error() string {
	return fmt.Sprintf("%s is not equal to %s", vce.v, vce.other)
}

func (vce versionMatchFailure) Pair() (v *Version, r Constraint) {
	return vce.v, vce.other
}

type MultiMatchFailure []MatchFailure

func (mmf MultiMatchFailure) Error() string {
	var buf bytes.Buffer

	for k, e := range mmf {
		if k < len(mmf)-1 {
			fmt.Fprintf(&buf, "%s\n", e)
		} else {
			fmt.Fprintf(&buf, e.Error())
		}
	}

	return buf.String()
}
