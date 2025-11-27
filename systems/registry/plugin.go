package registry

type FeatureKind int

const (
	FeatureNone FeatureKind = iota
	FeatureTreeNodeMangler
	FeatureEdict
)
