package adapter

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
