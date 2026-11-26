package kafkaschema

import (
	"cmp"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/bufbuild/protocompile/ast"
	"github.com/bufbuild/protocompile/options"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/bufbuild/protocompile/walk"
	gocmp "github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Compare two configured revisions. Comments and literal contents are part of the
// submitted document. Edits to them must reach the registry.
func equalSchemas(a, b string) bool {
	if a == b {
		return true
	}
	if left, ok := decodeJSON(a); ok {
		right, ok := decodeJSON(b)
		return ok && gocmp.Equal(left, right, gocmp.Comparer(equalNumbers))
	}
	left, err := parseProto(a)
	if err != nil {
		return false
	}
	right, err := parseProto(b)
	return err == nil && slices.Equal(protoTokens(left.AST()), protoTokens(right.AST()))
}

// Karapace fingerprints its rendered schema when reusing a historical version.
func sameRegistrationText(a, b string) bool {
	if a == b {
		return true
	}
	if left, ok := decodeJSON(a); ok {
		right, ok := decodeJSON(b)
		return ok && gocmp.Equal(left, right, gocmp.Comparer(sameRegistrationNumber))
	}
	if compiled, ok := binaryDescriptor(a); ok {
		// Descriptor equality alone also matches different registered texts, for
		// example a field with documentation or an explicit default JSON name.
		// Compare the spelling the registry generates from the binary input.
		if !sameProtobufDeclarations(a, b, protoregistry.GlobalFiles) {
			return false
		}
		a = binaryRegistrationSource(compiled)
	}
	left, err := parseProto(a)
	if err != nil {
		return false
	}
	right, err := parseProto(b)
	return err == nil && gocmp.Equal(registrationDeclarations(left.AST()), registrationDeclarations(right.AST()))
}

// Karapace deserializes a FileDescriptorProto into source before fingerprinting
// it. Preserve the exact field type spellings, statement boundaries and option
// order when checking a historical version before destroying a subject.
// https://github.com/Aiven-Open/karapace/blob/91f25f28494d238e50891b6ba94bfa56d3c0cc63/src/karapace/core/protobuf/serialization.py
func binaryRegistrationSource(file *descriptorpb.FileDescriptorProto) string {
	var source strings.Builder
	if file.Syntax != nil {
		fmt.Fprintf(&source, "syntax = %q;\n", file.GetSyntax())
	}
	if file.GetPackage() != "" {
		fmt.Fprintf(&source, "package %s;\n", file.GetPackage())
	}
	for i, name := range file.Dependency {
		qualifier := ""
		if slices.Contains(file.PublicDependency, int32(i)) {
			qualifier = " public"
		}
		fmt.Fprintf(&source, "import%s %q;\n", qualifier, name)
	}
	if file.Options != nil {
		options := file.Options.ProtoReflect()
		// This is the order used by Karapace's FileOptions deserializer.
		for _, name := range []protoreflect.Name{
			"java_package", "java_outer_classname", "optimize_for", "java_multiple_files", "go_package",
			"cc_generic_services", "java_generic_services", "py_generic_services", "java_generate_equals_and_hash",
			"deprecated", "java_string_check_utf8", "cc_enable_arenas", "objc_class_prefix", "csharp_namespace",
			"swift_prefix", "php_class_prefix", "php_namespace", "php_metadata_namespace", "ruby_package",
		} {
			field := options.Descriptor().Fields().ByName(name)
			if field == nil || !options.Has(field) {
				continue
			}
			value := options.Get(field).String()
			if field.Kind() == protoreflect.StringKind {
				value = strconv.Quote(value)
			}
			fmt.Fprintf(&source, "option %s = %s;\n", name, value)
		}
	}
	writeBinaryTypes(&source, file.MessageType, file.EnumType)
	return source.String()
}

func writeBinaryTypes(source *strings.Builder, messages []*descriptorpb.DescriptorProto, enums []*descriptorpb.EnumDescriptorProto) {
	for _, message := range messages {
		writeBinaryMessage(source, message)
	}
	for _, enum := range enums {
		fmt.Fprintf(source, "enum %s {\n", enum.GetName())
		for _, value := range enum.Value {
			fmt.Fprintf(source, "%s = %d", value.GetName(), value.GetNumber())
			if value.GetOptions().GetDeprecated() {
				source.WriteString(" [deprecated = true]")
			}
			source.WriteString(";\n")
		}
		source.WriteString("}\n")
	}
}

func writeBinaryMessage(source *strings.Builder, m *descriptorpb.DescriptorProto) {
	fmt.Fprintf(source, "message %s {\n", m.GetName())
	reserved := make([]string, 0, len(m.ReservedRange)+len(m.ReservedName))
	for _, r := range m.ReservedRange {
		first, last := r.GetStart(), r.GetEnd()-1
		value := strconv.Itoa(int(first))
		if first != last {
			value += " to " + strconv.Itoa(int(last))
		}
		reserved = append(reserved, value)
	}
	for _, name := range m.ReservedName {
		reserved = append(reserved, strconv.Quote(name))
	}
	if len(reserved) > 0 {
		fmt.Fprintf(source, "reserved %s;\n", strings.Join(reserved, ", "))
	}
	if m.GetOptions().GetMapEntry() {
		source.WriteString("option map_entry = true;\n")
	}
	if m.GetOptions().GetDeprecated() {
		source.WriteString("option deprecated = true;\n")
	}
	oneofs := make(map[int32][]*descriptorpb.FieldDescriptorProto)
	synthetic := make(map[int32]bool)
	for _, field := range m.Field {
		switch {
		case field.OneofIndex == nil:
			writeBinaryField(source, field)
		case field.GetProto3Optional():
			writeBinaryField(source, field)
			synthetic[field.GetOneofIndex()] = true
		default:
			index := field.GetOneofIndex()
			oneofs[index] = append(oneofs[index], field)
		}
	}
	for i, group := range m.OneofDecl {
		if synthetic[int32(i)] {
			continue
		}
		fmt.Fprintf(source, "oneof %s {\n", group.GetName())
		for _, field := range oneofs[int32(i)] {
			writeBinaryField(source, field)
		}
		source.WriteString("}\n")
	}
	writeBinaryTypes(source, m.NestedType, m.EnumType)
	source.WriteString("}\n")
}

func writeBinaryField(source *strings.Builder, field *descriptorpb.FieldDescriptorProto) {
	if field.GetProto3Optional() || field.GetLabel() != descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL {
		fmt.Fprint(source, strings.ToLower(strings.TrimPrefix(field.GetLabel().String(), "LABEL_")), " ")
	}
	fieldType := field.GetTypeName()
	if fieldType == "" {
		fieldType = protoreflect.Kind(field.GetType()).String()
	}
	fmt.Fprintf(source, "%s %s = %d;\n", fieldType, field.GetName(), field.GetNumber())
}

// Karapace prints top-level declarations in groups: messages precede enums,
// even if the submitted file interleaved them. Keep the full token stream with
// documentation for each group, preserving order inside it (including message
// indexes). SDK state only has the rendered source when first moving a subject.
func registrationDeclarations(file *ast.FileNode) map[string][]string {
	ends := make(map[ast.Token]string)
	if file.Syntax != nil {
		ends[file.Syntax.End()] = "syntax"
	}
	for _, decl := range file.Decls {
		kind := file.TokenInfo(decl.Start()).RawText()
		if imp, ok := decl.(*ast.ImportNode); ok && imp.Public != nil {
			kind = "import public"
		}
		ends[decl.End()] = kind
	}
	groups := make(map[string][]string)
	tokens := protoTokens(file)
	items := file.Items()
	start, pos := 0, 0
	for i, ok := items.First(); ok; i, ok = items.Next(i) {
		token, comment := file.GetItem(i)
		pos++
		if kind, ok := ends[token]; ok && !comment.IsValid() {
			groups[kind] = append(groups[kind], tokens[start:pos]...)
			start = pos
		}
	}
	groups[""] = tokens[start:]
	return groups
}

func decodeJSON(s string) (any, bool) {
	decoder := json.NewDecoder(strings.NewReader(s))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, false
	}
	return value, decoder.Decode(new(any)) == io.EOF
}

func equalNumbers(a, b json.Number) bool {
	if a == b {
		return true
	}
	digitsA, exponentA := decimalParts(a)
	digitsB, exponentB := decimalParts(b)
	return digitsA == digitsB && exponentA.Cmp(exponentB) == 0
}

// Karapace renders 1.50 and 15e-1 as 1.5, but distinguishes 1 from 1.0
// and preserves the sign of floating-point zero in a schema fingerprint.
func sameRegistrationNumber(a, b json.Number) bool {
	floating := strings.ContainsAny(a.String(), ".eE")
	if floating != strings.ContainsAny(b.String(), ".eE") {
		return false
	}
	if floating && strings.HasPrefix(a.String(), "-") != strings.HasPrefix(b.String(), "-") {
		return false
	}
	return equalNumbers(a, b)
}

// Compare significant digits and exponents separately so even large exponents
// require space proportional to the input.
func decimalParts(number json.Number) (string, *big.Int) {
	mantissa, power, _ := strings.Cut(strings.ToLower(number.String()), "e")
	exponent := new(big.Int)
	if power != "" {
		exponent.SetString(power, 10)
	}
	unsigned, negative := strings.CutPrefix(mantissa, "-")
	whole, fraction, _ := strings.Cut(unsigned, ".")
	digits := strings.TrimLeft(whole+fraction, "0")
	if digits == "" {
		return "0", exponent.SetInt64(0)
	}
	significant := strings.TrimRight(digits, "0")
	exponent.Add(exponent, big.NewInt(int64(len(digits)-len(significant)-len(fraction))))
	if negative {
		significant = "-" + significant
	}
	return significant, exponent
}

func parseProto(s string) (parser.Result, error) {
	handler := reporter.NewHandler(nil)
	file, err := parser.Parse("schema.proto", strings.NewReader(strings.ReplaceAll(s, "\r\n", "\n")), handler)
	if err != nil {
		return nil, err
	}
	// Karapace renders maps from a compiled descriptor as an explicit
	// synthetic entry and accepts that text as input. Keep all other
	// source-validation diagnostics when comparing configured revisions.
	handler = reporter.NewHandler(reporter.NewReporter(func(e reporter.ErrorWithPos) error {
		if strings.HasSuffix(e.Error(), "map_entry option should not be set explicitly; use map type instead") {
			return nil
		}
		return e
	}, nil))
	result, err := parser.ResultFromAST(file, true, handler)
	if errors.Is(err, reporter.ErrInvalidSource) && handler.ReporterError() == nil {
		return result, nil
	}
	return result, err
}

func protoTokens(file *ast.FileNode) []string {
	literals := make(map[ast.Token]string)
	trailingDocs := make(map[ast.Token]bool)
	_ = ast.Walk(file, &ast.SimpleVisitor{DoVisitNode: func(node ast.Node) error {
		switch n := node.(type) {
		case *ast.StringLiteralNode:
			literals[n.Start()] = strconv.Quote(n.AsString())
		case *ast.FieldNode, *ast.MapFieldNode, *ast.EnumValueNode, *ast.ReservedNode:
			// These declarations consume an inline comment as their own documentation.
			// After package, import, syntax, and options it belongs to the next declaration.
			trailingDocs[n.End()] = true
		}
		return nil
	}})
	var tokens []string
	var previousLine int
	var previousToken ast.Token
	items := file.Items()
	for i, ok := items.First(); ok; i, ok = items.Next(i) {
		info := file.ItemInfo(i)
		token, comment := file.GetItem(i)
		s := info.RawText()
		if literal, ok := literals[token]; ok && !comment.IsValid() {
			s = literal
		}
		if comment.IsValid() {
			// Karapace inserts a separator space when printing line comments.
			if body, lineComment := strings.CutPrefix(s, "//"); lineComment {
				s = "//" + strings.TrimPrefix(body, " ")
			}
			if info.Start().Line == previousLine && trailingDocs[previousToken] {
				s = "trailing " + s
			}
		}
		tokens = append(tokens, s)
		previousLine = info.End().Line
		previousToken = token
	}
	return tokens
}

// A verified registration can keep the submitted source even when Karapace
// rewrites documentation or renders a binary FileDescriptorProto as source.
// This comparison is only used for the response to that registration.
func sameProtobufDeclarations(a, b string, files *protoregistry.Files) bool {
	left, leftBinary, err := registryDescriptor(a)
	if err != nil {
		return false
	}
	right, rightBinary, err := registryDescriptor(b)
	if err != nil {
		return false
	}
	if proto.Equal(left, right) {
		return true
	}
	if leftBinary == rightBinary {
		return false
	}
	if leftBinary {
		restoreOptionalOneofs(left.MessageType, right.MessageType)
	} else {
		restoreOptionalOneofs(right.MessageType, left.MessageType)
	}
	// Compiled field types and names may be fully qualified in a binary
	// descriptor. Text-to-text comparisons never need to link external imports.
	left, err = linkedDescriptor(left, files)
	if err != nil {
		return false
	}
	right, err = linkedDescriptor(right, files)
	if err != nil {
		return false
	}
	if files != nil && slices.ContainsFunc(left.Dependency, func(name string) bool {
		_, err := files.FindFileByPath(name)
		return errors.Is(err, protoregistry.NotFound)
	}) {
		// The registry provides some imports without a subject or version,
		// such as confluent/type/decimal.proto. Source represents such a field
		// by its qualified type name. All pinned references have been linked.
		binary, source := left, right
		if rightBinary {
			binary, source = right, left
		}
		matchImportedFields(binary.Extension, source.Extension)
		matchImportedMessages(binary.MessageType, source.MessageType)
	}
	return proto.Equal(left, right)
}

// Karapace's binary export marks proto3 optional fields but omits their
// synthetic oneofs. Use the matching source field's oneof to link that binary.
func restoreOptionalOneofs(binary, source []*descriptorpb.DescriptorProto) {
	for i := range min(len(binary), len(source)) {
		b, s := binary[i], source[i]
		for j := range min(len(b.Field), len(s.Field)) {
			field, parsed := b.Field[j], s.Field[j]
			if field.GetProto3Optional() && field.OneofIndex == nil && parsed.GetProto3Optional() &&
				parsed.OneofIndex != nil && field.GetNumber() == parsed.GetNumber() {
				field.OneofIndex = proto.Int32(int32(len(b.OneofDecl))) //nolint:gosec // Descriptor indexes are int32.
				b.OneofDecl = append(b.OneofDecl, s.OneofDecl[parsed.GetOneofIndex()])
			}
		}
		restoreOptionalOneofs(b.NestedType, s.NestedType)
	}
}

func matchImportedFields(binary, source []*descriptorpb.FieldDescriptorProto) {
	for i := range min(len(binary), len(source)) {
		b, s := binary[i], source[i]
		name := b.GetTypeName()
		if s.Type == nil && strings.HasPrefix(name, ".") && s.GetTypeName() == name &&
			(b.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE || b.GetType() == descriptorpb.FieldDescriptorProto_TYPE_ENUM) {
			s.Type = b.Type
		}
	}
}

func matchImportedMessages(binary, source []*descriptorpb.DescriptorProto) {
	for i := range min(len(binary), len(source)) {
		matchImportedFields(binary[i].Field, source[i].Field)
		matchImportedFields(binary[i].Extension, source[i].Extension)
		matchImportedMessages(binary[i].NestedType, source[i].NestedType)
	}
}

func linkedDescriptor(desc *descriptorpb.FileDescriptorProto, files *protoregistry.Files) (*descriptorpb.FileDescriptorProto, error) {
	if files == nil {
		files = protoregistry.GlobalFiles
	}
	file, err := (protodesc.FileOptions{AllowUnresolvable: true}).New(desc, files)
	if err != nil {
		return nil, err
	}
	result := protodesc.ToFileDescriptorProto(file)
	setFieldJSONNames(result.MessageType, file.Messages())
	// Protobuf reflection no longer exposes weak imports.
	result.WeakDependency = desc.WeakDependency
	return result, nil
}

// A binary descriptor may omit json_name. The text parser materializes its
// default, so compare the names that protobuf actually uses for JSON encoding.
func setFieldJSONNames(messages []*descriptorpb.DescriptorProto, resolved protoreflect.MessageDescriptors) {
	for i, message := range messages {
		desc := resolved.Get(i)
		for j, field := range message.Field {
			field.JsonName = proto.String(desc.Fields().Get(j).JSONName())
		}
		setFieldJSONNames(message.NestedType, desc.Messages())
	}
}

func binaryDescriptor(s string) (*descriptorpb.FileDescriptorProto, bool) {
	desc := new(descriptorpb.FileDescriptorProto)
	data, err := base64.StdEncoding.DecodeString(s)
	return desc, err == nil && proto.Unmarshal(data, desc) == nil
}

func registryDescriptor(s string) (*descriptorpb.FileDescriptorProto, bool, error) {
	desc, binary := binaryDescriptor(s)
	if !binary {
		result, err := parseProto(s)
		if err != nil {
			return nil, false, err
		}
		if _, err := options.InterpretUnlinkedOptions(result); err != nil {
			return nil, false, err
		}
		desc = result.FileDescriptorProto()
	}
	// Source text carries no filename. Pick a common one that cannot be
	// mistaken for one of the references when linking a binary descriptor.
	name := "schema.proto"
	for slices.Contains(desc.Dependency, name) {
		name = "_" + name
	}
	desc.Name, desc.SourceCodeInfo = &name, nil
	// Karapace prints ordinary imports before public imports. Keep which
	// imports are public/weak when their positions change.
	deps := slices.Clone(desc.Dependency)
	slices.Sort(desc.Dependency)
	for _, indexes := range [][]int32{desc.PublicDependency, desc.WeakDependency} {
		for i, previous := range indexes {
			if previous < 0 || int(previous) >= len(deps) {
				return nil, false, fmt.Errorf("invalid import index %d", previous)
			}
			next, _ := slices.BinarySearch(desc.Dependency, deps[previous])
			indexes[i] = int32(next)
		}
		slices.Sort(indexes)
	}
	proto2Export := binary && (desc.GetSyntax() == "" || desc.GetSyntax() == "proto2")
	normalizeMessages(desc.MessageType, proto2Export)
	err := walk.DescriptorProtos(desc, normalizeDescriptorOptions)
	return desc, binary, err
}

// The registry omits empty RPC option blocks and writes deprecated=false in
// serialized enum values. Both representations describe the same behavior.
func normalizeDescriptorOptions(_ protoreflect.FullName, descriptor proto.Message) error {
	m := descriptor.ProtoReflect()
	field := m.Descriptor().Fields().ByName("options")
	if field == nil || !m.Has(field) {
		return nil
	}
	opts := m.Get(field).Message()
	if deprecated := opts.Descriptor().Fields().ByName("deprecated"); deprecated != nil && !opts.Get(deprecated).Bool() {
		opts.Clear(deprecated)
	}
	options := opts.Interface().(interface {
		GetUninterpretedOption() []*descriptorpb.UninterpretedOption
	})
	for _, opt := range options.GetUninterpretedOption() {
		if opt.AggregateValue == nil {
			continue
		}
		// Without the extension's descriptor an aggregate is stored as text.
		// Parse it so delimiters added by Karapace do not change its value.
		file, err := parser.Parse("option.proto", strings.NewReader("option (ext) = { "+opt.GetAggregateValue()+" };"), reporter.NewHandler(nil))
		if err != nil {
			return err
		}
		if len(file.Decls) != 1 {
			return errors.New("invalid aggregate option")
		}
		literal := file.Decls[0].(*ast.OptionNode).Val.(*ast.MessageLiteralNode)
		opt.AggregateValue = proto.String(strings.Join(aggregateFields(file, literal.Elements), " "))
	}
	if proto.Size(opts.Interface()) == 0 {
		m.Clear(field)
	}
	return nil
}

func aggregateFields(file *ast.FileNode, fields []*ast.MessageFieldNode) []string {
	// A repeated text-format field may be printed as one list. Distinct
	// fields can move around it, but the order of its own values matters.
	slices.SortStableFunc(fields, func(a, b *ast.MessageFieldNode) int {
		return strings.Compare(a.Name.Value(), b.Name.Value())
	})
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		elements, list := field.Val.Value().([]ast.ValueNode)
		if !list {
			elements = []ast.ValueNode{field.Val}
		}
		for _, elem := range elements {
			values = append(values, field.Name.Value()+":"+aggregateValue(file, elem))
		}
	}
	return values
}

func aggregateValue(file *ast.FileNode, node ast.ValueNode) string {
	switch v := node.Value().(type) {
	case string:
		return strconv.Quote(v)
	case []*ast.MessageFieldNode:
		return "{" + strings.Join(aggregateFields(file, v), " ") + "}"
	case []ast.ValueNode:
		elems := make([]string, 0, len(v))
		for _, elem := range v {
			elems = append(elems, aggregateValue(file, elem))
		}
		return "[" + strings.Join(elems, ",") + "]"
	default:
		return file.NodeInfo(node).RawText()
	}
}

func normalizeMessages(messages []*descriptorpb.DescriptorProto, proto2Export bool) {
	for _, m := range messages {
		if proto2Export {
			for _, f := range m.Field {
				if f.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL && f.OneofIndex == nil {
					// Karapace sets the proto3 presence flag on explicitly optional proto2 fields too.
					f.Proto3Optional = nil
				}
			}
		}
		// Ordinary fields and map entries precede oneof blocks and declared
		// nested messages in Karapace's source. Preserve declared message
		// order, which determines Confluent wire-format message indexes.
		group := func(f *descriptorpb.FieldDescriptorProto) int {
			if f.OneofIndex != nil && !f.GetProto3Optional() {
				// Each oneof prints its ordinary fields before its groups.
				first := 2*int(f.GetOneofIndex()) + 1
				if f.GetType() == descriptorpb.FieldDescriptorProto_TYPE_GROUP {
					return first + 1
				}
				return first
			}
			if f.GetType() == descriptorpb.FieldDescriptorProto_TYPE_GROUP {
				return 2*len(m.OneofDecl) + 1
			}
			return 0
		}
		slices.SortStableFunc(m.Field, func(a, b *descriptorpb.FieldDescriptorProto) int {
			return cmp.Compare(group(a), group(b))
		})
		slices.SortStableFunc(m.NestedType, func(a, b *descriptorpb.DescriptorProto) int {
			switch {
			case a.GetOptions().GetMapEntry() == b.GetOptions().GetMapEntry():
				return 0
			case a.GetOptions().GetMapEntry():
				return -1
			default:
				return 1
			}
		})
		normalizeMessages(m.NestedType, proto2Export)
	}
}
