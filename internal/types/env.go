package types

import (
	"fmt"
	"strings"
)

const (
	EnvDEV Env = "dev"
	EnvQA  Env = "qa"
	EnvUAT Env = "uat"
)

type Env string

var DeployTargetEnvs = []string{string(EnvQA), string(EnvUAT)}

func EnvFromString(s string) (Env, error) {
	vv := Env(s)
	if vv != EnvDEV && vv != EnvQA && vv != EnvUAT {
		return "", fmt.Errorf("%q is incorrect envoirenment", s)
	}

	return vv, nil
}

func EnvFromSuffix(s string) (Env, error) {
	switch s {
	case "":
		return EnvDEV, nil
	case "qa":
		return EnvQA, nil
	case "uat":
		return EnvUAT, nil
	}

	return "", fmt.Errorf("%q is incorrect envoirenment", s)
}

func (e Env) String() string {
	return strings.ToUpper(string(e))
}
