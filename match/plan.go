package match

import "gopkg.in/yaml.v2"

func LoadPlan(yml []byte) ([]*ReviewSession, error) {
	var sessionArray []*ReviewSession
	err := yaml.Unmarshal(yml, &sessionArray)
	if err != nil {
		return nil, err
	}

	return sessionArray, nil
}