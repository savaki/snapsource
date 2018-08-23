package generate

import (
	"errors"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

// String returns a pointer to the provide string or nil if the zero value was passed in
func String(in string) *string {
	if in == "" {
		return nil
	}

	return &in
}

func filename(in *descriptor.FileDescriptorProto) *string {
	if in.Name != nil {
		name := *in.Name
		ext := filepath.Ext(name)
		name = name[0 : len(name)-len(ext)]
		return String(name + ".pb.snap.go")
	}

	return String("events.pb.snap.go")
}

func packageName(in *descriptor.FileDescriptorProto) (string, error) {
	if in.Package != nil {
		return *in.Package, nil
	}

	if in.Name != nil {
		name := *in.Name
		ext := filepath.Ext(name)
		return name[0 : len(name)-len(ext)], nil
	}

	return "", errors.New("unable to determine package name")
}

func base(in string) string {
	idx := strings.LastIndex(in, ".")
	if idx == -1 {
		return in
	}
	return in[idx+1:]
}

func lower(in string) string {
	return strings.ToLower(in)
}

func camel(in string) string {
	segments := strings.Split(in, "_")
	capped := make([]string, 0, len(segments))

	for _, segment := range segments {
		if segment == "" {
			continue
		}
		capped = append(capped, strings.ToUpper(segment[0:1])+segment[1:])
	}
	return strings.Join(capped, "")
}

func typ(in interface{}) interface{} {
	switch v := in.(type) {
	case *descriptor.FieldDescriptorProto_Type:
		switch *v {
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			return "bool"
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			return "[]byte"
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
			return "float64"
		case descriptor.FieldDescriptorProto_TYPE_INT32:
			return "int32"
		case descriptor.FieldDescriptorProto_TYPE_INT64:
			return "int64"
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			return "string"
		case descriptor.FieldDescriptorProto_TYPE_UINT32:
			return "uint32"
		case descriptor.FieldDescriptorProto_TYPE_UINT64:
			return "uint64"
		default:
			return nil
		}
	default:
		return nil
	}
}

func other(fields []*descriptor.FieldDescriptorProto) interface{} {
	results := make([]*descriptor.FieldDescriptorProto, 0, len(fields))

	if fields != nil {
		for _, field := range fields {
			if v := strings.ToLower(*field.Name); v == "id" || v == "version" || v == "at" {
				continue
			}
			results = append(results, field)
		}
	}

	return results
}

func name(field *descriptor.FieldDescriptorProto) string {
	name := gogoproto.GetCustomName(field)
	if name != "" {
		return name
	}

	return camel(*field.Name)
}

func id(in *descriptor.DescriptorProto) string {
	for _, field := range in.Field {
		fieldName := *field.Name
		if strings.ToLower(fieldName) != "id" {
			continue
		}

		name := gogoproto.GetCustomName(field)
		if name != "" {
			return name
		}
		if fieldName == "ID" {
			return fieldName
		}

		return "Id"
	}

	return "Id"
}

func newTemplate(content string) (*template.Template, error) {
	fn := map[string]interface{}{
		"base":  base,
		"lower": lower,
		"camel": camel,
		"type":  typ,
		"other": other,
		"id":    id,
		"name":  name,
	}

	return template.New("page").Funcs(fn).Parse(content)
}

// findContainerMessage returns the message that contains all the other message types
func findContainerMessage(in *descriptor.FileDescriptorProto) (*descriptor.DescriptorProto, error) {
outer:
	for _, message := range in.MessageType {
		for index, field := range message.Field {
			if index > 0 {
				return nil, errors.New("not found")
			}
			if *field.Name != "type" || *field.Number != int32(1) {
				continue outer
			}
			return message, nil
		}
	}

	return nil, nil
}
