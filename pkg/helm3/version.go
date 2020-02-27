package helm3

import (
	"github.com/MChorfa/porter-helm3/pkg"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/porter/version"
)

func (m *Mixin) PrintVersion(opts version.Options) error {
	metadata := mixin.Metadata{
		Name: "helm3",
		VersionInfo: mixin.VersionInfo{
			Version: pkg.Version,
			Commit:  pkg.Commit,
			Author:  "Mohamed Chorfa",
		},
	}
	return version.PrintVersion(m.Context, opts, metadata)
}
