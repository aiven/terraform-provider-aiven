package vpc

import (
	"fmt"
	"strings"
)

type peeringVPCID struct {
	projectName      string
	vpcID            string
	peerCloudAccount string
	peerVPC          string
	peerRegion       *string
}

// parsePeerVPCID splits string id like "my-project/id/id/my-vpc" + optional "/region"
func parsePeerVPCID(src string) (*peeringVPCID, error) {
	chunks := strings.Split(src, "/")
	length := len(chunks)
	if length < 4 || 5 < length {
		return nil, fmt.Errorf("expected unix path-like string with 4-5 chunks, got %d", length)
	}

	pID := &peeringVPCID{
		projectName:      chunks[0],
		vpcID:            chunks[1],
		peerCloudAccount: chunks[2],
		peerVPC:          chunks[3],
	}

	if len(chunks) == 5 {
		pID.peerRegion = &chunks[4]
	}
	return pID, nil
}

func ConvertStateInfoToMap(s *map[string]any) map[string]string {
	if s == nil || len(*s) == 0 {
		return nil
	}

	r := make(map[string]string)
	for k, v := range *s {
		if _, ok := v.(string); ok {
			r[k] = v.(string)
		} else {
			r[k] = fmt.Sprintf("%+v", v)
		}
	}

	return r
}

func validateVPCID(i any, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	if len(strings.Split(v, "/")) != 2 {
		errors = append(errors, fmt.Errorf("invalid %v, expected <project_name>/<vpc_id>", k))
		return warnings, errors
	}

	return warnings, errors
}
