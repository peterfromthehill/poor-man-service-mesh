package classifiers

import (
	"poor-man-service-mesh/pkg/dpi/types"
	"regexp"
	"strings"
)

// HTTPClassifier struct
type HTTPClassifier struct{}

type HTTPClient struct {
	Method string
	Host   string
	Path   string
}

var httpVerbs = []string{
	"OPTIONS",
	"GET",
	"HEAD",
	"POST",
	"PUT",
	"DELETE",
	"TRACE",
	"CONNECT",
}

var regex *regexp.Regexp

func init() {
	var regexStr = "^(" + strings.Join(httpVerbs, "|") + ") ([^\\s]+) " +
		"HTTP/[12](.[01])?\r\n(.*\r\n)*\r\n"
	// regex should match the first line of all HTTP requests
	regex, _ = regexp.Compile(regexStr)
}

// HeuristicClassify for HTTPClassifier
func (classifier HTTPClassifier) HeuristicClassify(flow *types.Flow) (bool, interface{}) {
	httpC := HTTPClient{}
	return checkFlowPayload(flow, func(payload []byte) bool {
		detected := regex.Match(payload)
		if detected {
			result := regex.FindStringSubmatch(string(payload))
			if len(result) > 1 {
				httpC.Method = result[1]
				httpC.Path = result[2]
			}
			payloadStr := string(payload)
			payloadStr = strings.ReplaceAll(payloadStr, "\r", "")
			for i, line := range strings.Split(payloadStr, "\n") {
				if i != 0 {
					if len(line) > 0 && strings.HasPrefix(strings.ToUpper(line), "HOST:") {
						httpC.Host = strings.Split(line, " ")[1]
					}
				}
			}
		}
		return detected
	}), httpC
}

// GetProtocol returns the corresponding protocol
func (classifier HTTPClassifier) GetProtocol() types.Protocol {
	return types.HTTP
}
