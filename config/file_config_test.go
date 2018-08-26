package config

import (
	"testing"
)

func fileAssert(t *testing.T, expected interface{}, got interface{}) {
	t.Errorf("expected=%v, got=%v", expected, got)
}

func TestFileConfigGetInt(t *testing.T) {
	c := NewUtil(Options{
		FilePath: "../test/config.yaml",
		LogLevel: 100, // turn off logging
	})
	if i, ok := c.GetInt("integer-value"); !(ok && i == 36) {
		fileAssert(t, 36, i)
	}
	if i, ok := c.GetInt("not-integer-value"); !(!ok && i == 0) {
		// quoted integer is string, not int
		fileAssert(t, 0, i)
	}
	if i, ok := c.GetInt("negative-integer-value"); !(ok && i == -4) {
		fileAssert(t, -4, i)
	}
	if i, ok := c.GetInt("float-value"); !(ok && i == 11) {
		// decimal part truncated
		fileAssert(t, 11, i)
	}
	if i, ok := c.GetInt("not-float-value"); !(!ok && i == 0) {
		// quoted float is string, not float
		fileAssert(t, 0, i)
	}
}

func TestFileConfigGetString(t *testing.T) {
	c := NewUtil(Options{
		FilePath: "../test/config.yaml",
		LogLevel: 100, // turn off logging
	})
	if s, ok := c.GetString("string-value"); !(ok && s == "hey ho") {
		fileAssert(t, "hey ho", s)
	}
	if s, ok := c.GetString("unq-string-value"); !(ok && s == "considered") {
		fileAssert(t, "considered", s)
	}
	if s, ok := c.GetString("empty-string-value"); !(ok && s == "") {
		fileAssert(t, "", s)
	}
	if s, ok := c.GetString("float-value"); !(!ok && s == "") {
		// float value is not string
		fileAssert(t, "", s)
	}
}

func TestFileConfigGetFloat(t *testing.T) {
	c := NewUtil(Options{
		FilePath: "../test/config.yaml",
		LogLevel: 100, // turn off logging
	})
	if f, ok := c.GetFloat("float-value"); !(ok && f == 11.65425) {
		fileAssert(t, 11.65425, f)
	}
	if f, ok := c.GetFloat("not-float-value"); !(!ok && f == 0.0) {
		// quoted float is string, not float
		fileAssert(t, 0.0, f)
	}
	if f, ok := c.GetFloat("negative-float-value"); !(ok && f == -0.411) {
		fileAssert(t, -0.411, f)
	}
	if f, ok := c.GetFloat("integer-value"); !(ok && f == 36.0) {
		fileAssert(t, 36.0, f)
	}
}

func TestFileConfigGetBool(t *testing.T) {
	c := NewUtil(Options{
		FilePath: "../test/config.yaml",
		LogLevel: 100, // turn off logging
	})
	if b, ok := c.GetBool("boolean-value-1"); !(ok && b) {
		fileAssert(t, true, b)
	}
	if b, ok := c.GetBool("boolean-value-2"); !(ok && b) {
		fileAssert(t, true, b)
	}
	if b, ok := c.GetBool("boolean-value-3"); !(ok && b) {
		fileAssert(t, true, b)
	}
	if b, ok := c.GetBool("not-boolean-value"); !(!ok && !b) {
		// quoted value is string, not bool
		fileAssert(t, false, b)
	}
	if b, ok := c.GetBool("not-boolean-value-2"); !(!ok && !b) {
		// integer value is int, not bool
		fileAssert(t, false, b)
	}
}

func TestFileConfigUseCase(t *testing.T) {
	c := NewUtil(Options{
		FilePath: "../test/config.yaml",
		LogLevel: 100, // turn off logging
	})
	if s, ok := c.GetString("some-config.protocol"); !(ok && s == "tcp") {
		fileAssert(t, "tcp", s)
	}
	if s, ok := c.GetString("some-config.address.ip"); !(ok && s == "127.0.0.2") {
		fileAssert(t, "127.0.0.2", s)
	}
	if i, ok := c.GetInt("some-config.address.port"); !(ok && i == 3000) {
		fileAssert(t, 3000, i)
	}
}

func TestFileConfigBundle(t *testing.T) {
	type someConfig struct {
		Protocol string
		Address  struct {
			IP   string `config:"ip"`
			Port int
		}
		Version  string // `config:",watch"`
		SomeBool bool   `config:"some-boolean,watch"`
	}

	sc := someConfig{}

	NewBundle("some-config", &sc, Options{
		FilePath: "../test/config.yaml",
		LogLevel: 100, // turn off logging
	})

	if sc.Protocol != "tcp" {
		fileAssert(t, "tcp", sc.Protocol)
	}
	if sc.Address.IP != "127.0.0.2" {
		fileAssert(t, "127.0.0.2", sc.Address.IP)
	}
	if sc.Address.Port != 3000 {
		fileAssert(t, 3000, sc.Address.Port)
	}
	if sc.SomeBool != true {
		fileAssert(t, true, sc.SomeBool)
	}
}

func TestFileConfigDeep(t *testing.T) {
	c := NewUtil(Options{
		FilePath: "../test/config.yaml",
		LogLevel: 100, // turn off logging
	})
	if i, ok := c.GetInt("deep-config.l1.l2.l_3.l-4.l 5.6l"); !(ok && i == 6) {
		fileAssert(t, 6, i)
	}
}
