package definition

import (
	"fmt"
	"strings"
)

// Definition is the definition of one or more services.
// In templates, it is usually accessed via the `def` variable.
//  Package name is <%= def.PackageName %>
type Definition struct {
	Services       []Service
	PackageName    string
	PackageComment string

	Comments map[string]string
}

func (d Definition) String() string {
	s := "package " + d.PackageName + "\n\n"
	for i := range d.Services {
		s += d.Services[i].String()
	}
	return s
}

// Structure gets a Structure by name.
func (d Definition) Structure(name string) *Structure {
	for _, service := range d.Services {
		for _, structure := range service.Structures {
			if structure.Name == name {
				return &structure
			}
		}
	}
	return nil
}

// Service describes a logically grouped set of endpoints.
type Service struct {
	Name       string
	Comment    string
	Methods    []Method
	Structures []Structure
}

// EnsureStructure adds the Structure to the service if it isn't
// already there.
func (s *Service) EnsureStructure(structure Structure) {
	for i := range s.Structures {
		if s.Structures[i].Name == structure.Name {
			return
		}
	}
	s.Structures = append(s.Structures, structure)
}

func (s Service) String() string {
	var str string
	if s.Comment != "" {
		str += "// " + s.Comment + "\n"
	}
	str += "type " + s.Name + " interface {\n"
	for i := range s.Methods {
		str += "\t" + s.Methods[i].String()
	}
	str += "}\n\n"
	for i := range s.Structures {
		str += s.Structures[i].String()
	}
	return str
}

// Method is a single method.
type Method struct {
	Name              string
	Comment           string
	RequestStructure  Structure
	ResponseStructure Structure
}

func (m Method) String() string {
	var str string
	if m.Comment != "" {
		str += "// " + m.Comment + "/n"
	}
	str += m.Name + "(*" + m.RequestStructure.Name + ") *" + m.ResponseStructure.Name + "\n"
	return str
}

// Structure describes a data structure.
type Structure struct {
	Name       string
	Comment    string
	Fields     []Field
	IsImported bool

	IsRequestObject  bool
	IsResponseObject bool
}

func (s Structure) String() string {
	var str string
	if s.Comment != "" {
		str += "// " + s.Comment + "\n"
	}
	str += "type " + s.Name + " struct {\n"
	for i := range s.Fields {
		str += "\t" + s.Fields[i].String() + "\n"
	}
	str += "}\n\n"
	return str
}

// HasFields gets whether the Structure has any fields or not.
func (s Structure) HasFields() bool {
	return len(s.Fields) > 0
} // TODO: test

// HasField gets whether the Structure has a specific field or not.
func (s Structure) HasField(field string) bool {
	for _, f := range s.Fields {
		if f.Name == field {
			return true
		}
	}
	return false
} // TODO: test

// FieldsOfType gets all Field objects that have a specific type.
func (s Structure) FieldsOfType(typename string) []Field {
	var fields []Field
	for _, field := range s.Fields {
		if field.Type.Name == typename {
			fields = append(fields, field)
		}
	}
	return fields
}

// Field describes a structure field.
type Field struct {
	Name    string
	Comment string
	Type    Type
}

func (f Field) String() string {
	return fmt.Sprintf("%s %s", f.Name, f.Type.code())
}

// IsExported gets whether the field is exported or not.
func (f Field) IsExported() bool {
	return strings.ToUpper(f.Name[0:1]) == f.Name[0:1]
}

// Type describes the type of a Field.
type Type struct {
	Name       string
	IsMultiple bool
	IsStruct   bool
	IsImported bool
}

func (t Type) code() string {
	str := t.Name
	if t.IsMultiple {
		str = "[]" + str
	}
	return str
}
