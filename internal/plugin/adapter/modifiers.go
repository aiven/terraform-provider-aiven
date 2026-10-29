package adapter

import "fmt"

type MapModifier func(d ResourceData, dto map[string]any) error

func ComposeMapModifiers(modifiers ...MapModifier) MapModifier {
	return func(d ResourceData, dto map[string]any) error {
		for _, modifier := range modifiers {
			if err := modifier(d, dto); err != nil {
				return err
			}
		}
		return nil
	}
}

// RenameFields renames field names terraform name -> dto name
func RenameFields(fields map[string]string) MapModifier {
	return func(_ ResourceData, dto map[string]any) error {
		for was, want := range fields {
			v, ok := dto[was]
			if ok {
				dto[want] = v
				delete(dto, was)
			}
		}
		return nil
	}
}

// SetFixedValues sets request fields from key-value pairs, overriding the plan.
// Used for fields removed from the schema, e.g. the discriminator of a polymorphic endpoint.
func SetFixedValues(kv ...string) MapModifier {
	return func(_ ResourceData, dto map[string]any) error {
		if len(kv)%2 != 0 {
			return fmt.Errorf("SetFixedValues expects key-value pairs, got %d arguments", len(kv))
		}
		for i := 0; i < len(kv); i += 2 {
			dto[kv[i]] = kv[i+1]
		}
		return nil
	}
}
