package actors

import (
	"fmt"
	"strings"
)

// ChannelLookupError is used when a channel or a value within can't be resolved
type ChannelLookupError string

func (l ChannelLookupError) Error() string {
	return string(l)
}

// ChannelMutabilityViolation is used when an attempt is made to overwrite an existing channel entry
type ChannelMutabilityViolation string

func (m ChannelMutabilityViolation) Error() string {
	return string(m)
}

// ChannelManager handles channel data storage and lookup
type ChannelManager struct {
	data map[string]interface{}
}

// NewChannelManager creates a new ChannelManager instance
func NewChannelManager() *ChannelManager {
	return NewChannelManagerWithInitialData(map[string]interface{}{})
}

// NewChannelManagerWithInitialData creates a new ChannelManager instance
func NewChannelManagerWithInitialData(data map[string]interface{}) *ChannelManager {
	return &ChannelManager{
		data: data,
	}
}

// ResolveVariable retrieves the variable assigned
func (manager *ChannelManager) ResolveVariable(spec string) (interface{}, error) {
	if strings.HasPrefix(spec, "@") && strings.HasSuffix(spec, "@") {
		return resolveChannelVariable(manager.data, strings.Split(strings.Trim(spec, "@"), ".")...)
	}
	return spec, nil
}

// AssignFiltered assigns the channel data filtered based on the channel definition
func (manager *ChannelManager) AssignFiltered(channels []channelDefinition, data map[string]interface{}) error {
	filtered, err := filterChannels(channels, data)
	if err != nil {
		return err
	}
	for k := range filtered {
		if _, ok := manager.data[k]; ok {
			return ChannelMutabilityViolation(fmt.Sprintf("Attempt to overwrite %s", k))
		}
		manager.data[k] = filtered[k]
	}
	return nil
}

// GetFiltered returns just the elements as specified in the definition
func (manager *ChannelManager) GetFiltered(channels []channelDefinition) (map[string]interface{}, error) {
	return filterChannels(channels, manager.data)
}

// AssignToVariable resolves the spec and assigns the passed variable to it
func (manager *ChannelManager) AssignToVariable(spec string, value interface{}) error {
	if strings.HasPrefix(spec, "@") && strings.HasSuffix(spec, "@") {
		parts := strings.Split(strings.Trim(spec, "@"), ".")
		if len(parts) == 1 && len(parts[0]) == 0 {
			return ChannelLookupError("Invalid spec to assign to")
		}

		var base interface{} = map[string]interface{}{}
		result := base.(map[string]interface{})
		if len(parts) > 1 {
			for _, k := range parts[1 : len(parts)-1] {
				result[k] = map[string]interface{}{}
				result = result[k].(map[string]interface{})
			}
			result[parts[len(parts)-1]] = value
		} else {
			base = value
		}
		manager.data[parts[0]] = base
		return nil
	}
	return ChannelLookupError("Failed to resolve variable to assign to")
}

func filterChannels(channels []channelDefinition, data map[string]interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	missing := []string{}
	for _, entry := range channels {
		if value, ok := data[entry.Name]; ok {
			result[entry.Name] = value
		} else {
			missing = append(missing, entry.Name)
		}
	}
	if len(missing) > 0 {
		return nil, ChannelLookupError(fmt.Sprintf("ERROR: Missing channel(s) %s", strings.Join(missing, ",")))
	}
	return result, nil
}

func resolveChannelVariable(data map[string]interface{}, elements ...string) (interface{}, error) {
	if len(elements) > 0 {
		if value, ok := data[elements[0]]; ok {
			if len(elements) == 1 {
				return value, nil
			}
			if mapping, ok := value.(map[string]interface{}); ok {
				return resolveChannelVariable(mapping, elements[1:]...)
			} else if !ok {
				return nil, ChannelLookupError("Expected a map while resolving channel variable")
			}
		}
	}
	return nil, ChannelLookupError("Resolving channel variable failed")
}
