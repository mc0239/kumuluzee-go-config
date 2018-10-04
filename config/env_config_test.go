package config

import (
	"testing"
)

func envAssert(t *testing.T, expected interface{}, got interface{}) {
	if expected != got {
		t.Errorf("expected=%v, got=%v", expected, got)
	}
}

func TestEnvKey(t *testing.T) {
	keys := []string{"kumuluzee", "KumuluzEE[0]", "lev1.lev2[5].LEV3", "v€ry-c00l"}
	expNorm := []string{"kumuluzee", "KumuluzEE_0_", "lev1_lev2_5__LEV3", "v_ry_c00l"}
	expNorm2 := []string{"KUMULUZEE", "KUMULUZEE_0_", "LEV1_LEV2_5__LEV3", "V_RY_C00L"}
	expLeg1 := []string{"KUMULUZEE", "KUMULUZEE0", "LEV1_LEV25_LEV3", "V€RYC00L"}
	expLeg2 := []string{"KUMULUZEE", "KUMULUZEE[0]", "LEV1_LEV2[5]_LEV3", "V€RY-C00L"}

	for i, keyName := range keys {
		envAssert(t, expNorm[i], normalizeKey(keyName))
		envAssert(t, expNorm2[i], normalizeKeyUpper(keyName))
		envAssert(t, expLeg1[i], parseKeyLegacy1(keyName))
		envAssert(t, expLeg2[i], parseKeyLegacy2(keyName))
	}
}
