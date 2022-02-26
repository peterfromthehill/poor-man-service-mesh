package types

type Module interface {
	Initialize() error
	Destroy() error
	ClassifyFlow(flow *Flow) ClassificationResult
}
