package config

import (
	"testing"
)

func TestFileConfigGetInt(t *testing.T) {
	c := NewUtil(Options{
		FilePath: "../test/config.yaml",
		LogLevel: 100, // turn off logging
	})

	// Read an int value
	if i, ok := c.GetInt("integer-value"); !(ok && i == 36) {
		t.Errorf("Test failed: reading int, expected: %d, got: %d", 36, i)
	}

	// Read an int value, stored as string
	if i, ok := c.GetInt("not-integer-value"); !(!ok && i == 0) {
		t.Errorf("Test failed: reading int, expected: %d, got: %d", 0, i)
		// TODO: quoted integer is string, not int!
	}

	// Read a float value as int, truncating decimal part
	if i, ok := c.GetInt("float-value"); !(ok && i == 11) {
		t.Errorf("Test failed: reading int, expected: %d, got: %d", 11, i)
	}

	// Read a float value as int, stored as string, truncating decimal part
	if i, ok := c.GetInt("not-float-value"); !(!ok && i == 0) {
		t.Errorf("Test failed: reading int, expected: %d, got: %d", 0, i)
	}

	// Read a negative int value
	if i, ok := c.GetInt("negative-integer-value"); !(ok && i == -4) {
		t.Errorf("Test failed: reading int, expected: %d, got: %d", -4, i)
	}
}

func TestFileConfigGetString(t *testing.T) {
	c := NewUtil(Options{
		FilePath: "../test/config.yaml",
		LogLevel: 100, // turn off logging
	})

	// Read a string value, quoted
	if s, ok := c.GetString("string-value"); !(ok && s == "hey ho") {
		t.Errorf("Test failed: reading string, expected: %s, got: %s", "hey ho", s)
	}

	// Read a string value, unquoted
	if s, ok := c.GetString("unq-string-value"); !(ok && s == "considered") {
		t.Errorf("Test failed: reading string, expected: %s, got: %s", "considered", s)
	}

	// Read a string value, empty
	if s, ok := c.GetString("empty-string-value"); !(ok && s == "") {
		t.Errorf("Test failed: reading string, expected: %s, got: %s", "", s)
	}

	// Read a float value, stored as float
	if s, ok := c.GetString("float-value"); !(!ok && s == "") {
		t.Errorf("Test failed: reading string, expected: %s, got: %s", "", s)
	}
}

func TestFileConfigGetFloat(t *testing.T) {
	c := NewUtil(Options{
		FilePath: "../test/config.yaml",
		LogLevel: 100, // turn off logging
	})

	// Read a float value
	if f, ok := c.GetFloat("float-value"); !(ok && f == 11.65425) {
		t.Errorf("Test failed: reading float, expected: %f, got: %f", 11.65425, f)
	}

	// Read a float value, stored as string
	if f, ok := c.GetFloat("not-float-value"); !(!ok && f == 0.0) {
		t.Errorf("Test failed: reading float, expected: %f, got: %f", 0.0, f)
		// TODO: quoted float is string, not float!
	}

	// Read a float value, stored as int
	if f, ok := c.GetFloat("integer-value"); !(ok && f == 36.0) {
		t.Errorf("Test failed: reading float, expected: %f, got: %f", 36.0, f)
	}

	// Read a negative float value
	if f, ok := c.GetFloat("negative-float-value"); !(ok && f == -0.411) {
		t.Errorf("Test failed: reading float, expected: %f, got: %f", -0.411, f)
	}
}

func TestFileConfigGetBool(t *testing.T) {
	c := NewUtil(Options{
		FilePath: "../test/config.yaml",
		LogLevel: 100, // turn off logging
	})

	// Read a bool value
	if b, ok := c.GetBool("boolean-value-1"); !(ok && b) {
		t.Errorf("Test failed: reading bool, expected: %v, got: %v", true, b)
	}
	if b, ok := c.GetBool("boolean-value-2"); !(ok && b) {
		t.Errorf("Test failed: reading bool, expected: %v, got: %v", true, b)
	}
	if b, ok := c.GetBool("boolean-value-3"); !(ok && b) {
		t.Errorf("Test failed: reading bool, expected: %v, got: %v", true, b)
	}

	// Read a bool value, stored as string
	if b, ok := c.GetBool("not-boolean-value"); !(!ok && !b) {
		t.Errorf("Test failed: reading bool, expected: %v, got: %v", false, b)
	}
	// Read a bool value, stored as int
	if b, ok := c.GetBool("not-boolean-value-2"); !(!ok && !b) {
		t.Errorf("Test failed: reading bool, expected: %v, got: %v", false, b)
	}
}

func TestFileConfigUseCase(t *testing.T) {
	c := NewUtil(Options{
		FilePath: "../test/config.yaml",
		LogLevel: 100, // turn off logging
	})

	if s, ok := c.GetString("some-config.protocol"); !(ok && s == "tcp") {
		t.Errorf("Test failed: reading string, expected: %s, got: %s", "tcp", s)
	}
	if s, ok := c.GetString("some-config.address.ip"); !(ok && s == "127.0.0.2") {
		t.Errorf("Test failed: reading string, expected: %s, got: %s", "127.0.0.2", s)
	}
	if i, ok := c.GetInt("some-config.address.port"); !(ok && i == 3000) {
		t.Errorf("Test failed: reading int, expected: %d, got: %d", 3000, i)
	}
}

func TestFileConfigBundle(t *testing.T) {
	type someConfig struct {
		Protocol string
		Address  struct {
			IP   string `mapstructure:"ip"`
			Port int
		}
		Version  string
		SomeBool bool `mapstructure:"some-boolean"`
	}

	sc := someConfig{}

	NewBundle("some-config", &sc, Options{
		FilePath: "../test/config.yaml",
		LogLevel: 100, // turn off logging
	})

	if sc.Protocol != "tcp" {
		t.Errorf("Test failed: bundle, expected: %s, got: %s", "tcp", sc.Protocol)
	}
	if sc.Address.IP != "127.0.0.2" {
		t.Errorf("Test failed: bundle, expected: %s, got: %s", "127.0.0.2", sc.Address.IP)
	}
	if sc.Address.Port != 3000 {
		t.Errorf("Test failed: bundle, expected: %d, got: %d", 3000, sc.Address.Port)
	}
	if sc.SomeBool != true {
		t.Errorf("Test failed: bundle, expected: %v, got: %v", true, sc.SomeBool)
	}
}

func TestFileConfigDeep(t *testing.T) {
	c := NewUtil(Options{
		FilePath: "../test/config.yaml",
		LogLevel: 100, // turn off logging
	})

	if i, ok := c.GetInt("deep-config.l1.l2.l_3.l-4.l 5.6l"); !(ok && i == 6) {
		t.Errorf("Test failed: deep, expected: %d, got: %d", 6, i)
	}
}
