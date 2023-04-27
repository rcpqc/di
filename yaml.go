package di

type YamlResolver struct {
	Resolve(rv reflect.Value, node any) error
}


func (o *YamlResolver) Resolve(rv reflect.Value, node any) error {
	if reflect.PointerTo(rv.Type()).Implements(unmarshalerType) && rv.CanAddr() {
		return rv.Addr().Interface().(unmarshaler).UnmarshalYAML(node)
	}
	switch rv.Kind() {
	case reflect.Ptr:
		return resolvePtr(rv, node)
	case reflect.Array, reflect.Slice:
		return resolveSlice(rv, node)
	case reflect.Map:
		return resolveMap(rv, node)
	case reflect.Struct:
		return resolveStruct(rv, node)
	case reflect.Interface:
		return resolveInterface(rv, node)
	case reflect.Chan, reflect.UnsafePointer, reflect.Func:
		return fmt.Errorf("rv.Kind == %v not supported", rv)
	default:
		return resolveNormal(rv, node)
	}
	return nil
}


func resolvePtr(rv reflect.Value, node *yaml.Node) error {
	if rv.IsNil() {
		rv.Set(reflect.New(rv.Type().Elem()))
	}
	return resolve(rv.Elem(), node)
}

func resolveSlice(rv reflect.Value, node *yaml.Node) error {
	nodes := []yaml.Node{}
	if err := node.Decode(&nodes); err != nil {
		return err
	}
	rv.Set(reflect.MakeSlice(rv.Type(), len(nodes), len(nodes)))
	for i, node := range nodes {
		elem := reflect.New(rv.Type().Elem()).Elem()
		if err := resolve(elem, &node); err != nil {
			return err
		}
		rv.Index(i).Set(elem)
	}
	return nil
}

func resolveMap(rv reflect.Value, node *yaml.Node) error {
	nodes := map[string]yaml.Node{}
	if err := node.Decode(&nodes); err != nil {
		return err
	}
	rv.Set(reflect.MakeMap(rv.Type()))
	for key, node := range nodes {
		elem := reflect.New(rv.Type().Elem()).Elem()
		if err := resolve(elem, &node); err != nil {
			return err
		}
		rv.SetMapIndex(reflect.ValueOf(key), elem)
	}
	return nil
}

func resolveStruct(rv reflect.Value, node *yaml.Node) error {
	fields := map[string]yaml.Node{}
	if err := node.Decode(&fields); err != nil {
		return err
	}
	for i := 0; i < rv.NumField(); i++ {
		if !rv.Field(i).CanSet() {
			continue
		}
		tag, inline := tagParse(rv.Type().Field(i).Tag.Get("yaml"))

		if tag == "-" {
			continue
		}
		if inline {
			if err := resolve(rv.Field(i), node); err != nil {
				return err
			}
		} else {
			if tag == "" {
				tag = strings.ToLower(rv.Type().Field(i).Name)
			}
			field, ok := fields[tag]
			if !ok {
				continue
			}
			if err := resolve(rv.Field(i), &field); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolveNormal(rv reflect.Value, node *yaml.Node) error {
	elemPtr := reflect.New(rv.Type())
	if err := node.Decode(elemPtr.Interface()); err != nil {
		return err
	}
	rv.Set(elemPtr.Elem())
	return nil
}

func resolveInterface(rv reflect.Value, node *yaml.Node) error {
	if !rv.Type().Implements(ObjectType) {
		return resolveNormal(rv, node)
	}
	o, err := newObject(node)
	if err != nil {
		return err
	}
	rv.Set(reflect.ValueOf(o))
	return nil
}

type unmarshaler interface{ UnmarshalYAML(value *yaml.Node) error }

var unmarshalerType = reflect.TypeOf((*unmarshaler)(nil)).Elem()
