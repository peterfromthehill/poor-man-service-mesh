package dpi

import (
	"poor-man-service-mesh/pkg/dpi/classifiers"
	"time"

	"poor-man-service-mesh/pkg/dpi/types"
)

var activatedModules []types.Module
var moduleList = []types.Module{
	classifiers.NewClassifierModule(),
}
var cacheExpiration = 5 * time.Minute

// Initialize initializes the library and the selected modules.
func Initialize() (errs []error) {

	types.InitCache(cacheExpiration)
	for _, module := range moduleList {
		activated := false
		for _, activeModule := range activatedModules {
			if activeModule == module {
				activated = true
				break
			}
		}
		if !activated {
			err := module.Initialize()
			if err == nil {
				activatedModules = append(activatedModules, module)
			} else {
				errs = append(errs, err)
			}
		}
	}
	return
}

// Destroy frees all allocated resources and deactivates the active modules.
func Destroy() (errs []error) {
	types.DestroyCache()
	newActivatedModules := make([]types.Module, 0)
	for _, module := range activatedModules {
		err := module.Destroy()
		if err != nil {
			newActivatedModules = append(newActivatedModules, module)
			errs = append(errs, err)
		}
	}
	activatedModules = newActivatedModules
	return
}

func GetPacket(payload []byte) *types.Packet {
	return &types.Packet{
		Payload: payload,
	}
}

// ClassifyFlow takes a Flow and tries to classify it with all of the activated
// modules in order, until one of them manages to classify it. It returns
// the detected protocol as well as the source that made the classification.
// If no classification is made, the protocol Unknown is returned.
func ClassifyFlow(packet *types.Packet) []types.Protocol {
	var result []types.Protocol
	for _, module := range activatedModules {
		resultTmp := module.ClassifyFlow(packet)
		result = append(result, resultTmp.ClassificationResult...)
	}
	return result
}
