package actors

import (
	"testing"
)

var manager *ChannelManager

func init() {
	manager = NewChannelManager()
	manager.data["a"] = map[string]interface{}{
		"b": map[string]interface{}{
			"c": map[string]interface{}{
				"d": "a.b.c.d=value",
			},
		},
	}
}

func TestChannelResolveNonExisting(t *testing.T) {
	_, err := manager.ResolveVariable("@DoesNotExist@")
	if err == nil {
		t.Log("Did not fail to lookup variable")
		t.Fail()
	}
}

func TestChannelResolveNonSpec(t *testing.T) {
	input := "IsNotASPec@"
	result, err := manager.ResolveVariable(input)
	if err != nil {
		t.Log("Did not fail to lookup variable")
		t.Fail()
	}
	if svalue, ok := result.(string); !ok {
		t.Log("Resulting value is not a string")
		t.Fail()
	} else if svalue != input {
		t.Logf("Expected: %s got %s", input, svalue)
		t.Fail()
	}
}

func TestChannelResolve(t *testing.T) {
	result, err := manager.ResolveVariable("@a.b.c.d@")
	if err != nil {
		t.Logf("Failed to lookup variable: %s", err.Error())
		t.Fail()
		return
	}
	value, ok := result.(string)
	if !ok {
		t.Logf("Failed to resolve variable - Expected type: string got: %T (%v)", result, value)
		t.Fail()
	} else if value != "a.b.c.d=value" {
		t.Logf("Failed to resolve variable - Expected value \"a.b.c.d=value\" got: %s", value)
		t.Fail()
	}
}

func TestChannelResolve2(t *testing.T) {
	result, err := manager.ResolveVariable("@a.b.c@")
	if err != nil {
		t.Logf("Failed to lookup variable: %s", err.Error())
		t.Fail()
		return
	}
	value, ok := result.(map[string]interface{})
	if !ok {
		t.Logf("Failed to resolve variable - Expected type: map[string]interface{} got: %T (%v)", value, value)
		t.Fail()
	} else if value["d"] != "a.b.c.d=value" {
		t.Logf("Failed to resolve variable - Expected value \"a.b.c.d=value\" got: %s", value["d"])
		t.Fail()
	}
}

func TestChannelAssignToVariables(t *testing.T) {
	err := manager.AssignToVariable("@b.a.b.c.d@", "b.a.b.c.d=value")
	if err != nil {
		t.Logf("Failed to lookup variable: %s", err.Error())
		t.Fail()
		return
	}
	result, err := manager.ResolveVariable("@b.a.b.c.d@")
	value, ok := result.(string)
	if !ok {
		t.Logf("Failed to resolve variable - Expected type: string got: %T (%v)", result, value)
		t.Fail()
	} else if value != "b.a.b.c.d=value" {
		t.Logf("Failed to resolve variable - Expected value \"a.b.c.d=value\" got: %s", value)
		t.Fail()
	}
}

func TestChannelAssignToVariables2(t *testing.T) {
	err := manager.AssignToVariable("@c@", map[string]interface{}{"value": "value"})
	if err != nil {
		t.Logf("Failed to lookup variable: %s", err.Error())
		t.Fail()
		return
	}
	result, err := manager.ResolveVariable("@c.value@")
	value, ok := result.(string)
	if !ok {
		t.Logf("Failed to resolve variable - Expected type: string got: %T (%v)", result, value)
		t.Fail()
	} else if value != "value" {
		t.Logf("Failed to resolve variable - Expected value \"a.b.c.d=value\" got: %s", value)
		t.Fail()
	}
}

func TestChannelAssignToVariableBadSpec(t *testing.T) {
	if err := manager.AssignToVariable("@@", nil); err == nil {
		t.Log("Expected assignment to '@@' to fail")
		t.Fail()
	}
	if err := manager.AssignToVariable("@", nil); err == nil {
		t.Log("Expected assignment to '@' to fail")
		t.Fail()
	}
	if err := manager.AssignToVariable("x@x", nil); err == nil {
		t.Log("Expected assignment to 'x@x' to fail")
		t.Fail()
	}
}
