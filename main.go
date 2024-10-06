package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"
)

// CodeElement represents a structured code element extracted from the SCIP index.
type CodeElement struct {
	Name      string            `json:"name"`
	Signature string            `json:"signature"`
	CodeType  string            `json:"code_type"`
	Docstring string            `json:"docstring"`
	Line      int32             `json:"line"`
	LineFrom  int32             `json:"line_from"`
	LineTo    int32             `json:"line_to"`
	Context   map[string]string `json:"context"`
}

func main() {
	// Parse command-line flags for input and output files.
	inputFile := flag.String("input", "index.scip", "Path to the input SCIP index file")
	outputFile := flag.String("output", "structure.json", "Path to the output JSON file")
	flag.Parse()

	// Read and parse the SCIP index file.
	index, err := readSCIPIndex(*inputFile)
	if err != nil {
		log.Fatalf("Error reading SCIP index: %v", err)
	}

	// Process the index to extract code elements.
	elements := processIndex(index)

	// Write the extracted elements to the output JSON file.
	if err := writeJSONOutput(*outputFile, elements); err != nil {
		log.Fatalf("Error writing JSON output: %v", err)
	}

	fmt.Printf("Successfully generated %s with %d code elements\n", *outputFile, len(elements))
}

// readSCIPIndex reads and unmarshals the SCIP index from the given file.
func readSCIPIndex(filename string) (*scip.Index, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read SCIP file: %w", err)
	}

	var index scip.Index
	if err := proto.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SCIP index: %w", err)
	}

	return &index, nil
}

// processIndex iterates through the SCIP index to extract code elements.
func processIndex(index *scip.Index) []CodeElement {
	var elements []CodeElement

	for _, doc := range index.Documents {
		for _, sym := range doc.Symbols {
			element, err := processSymbol(doc, sym)
			if err != nil {
				log.Printf("Skipping symbol due to error: %v", err)
				continue
			}

			if element.CodeType == "Unknown" {
				continue
			}

			elements = append(elements, element)
		}
	}

	return elements
}

// processSymbol extracts information from a SymbolInformation object to create a CodeElement.
func processSymbol(doc *scip.Document, sym *scip.SymbolInformation) (CodeElement, error) {
	log.Printf("Processing symbol: %s in file: %s", sym.Symbol, doc.RelativePath)

	line, lineFrom, lineTo, snippet, err := getSnippet(doc, sym)
	if err != nil {
		log.Printf("Warning: Could not extract snippet for symbol %s: %v", sym.Symbol, err)
		// Proceeding without snippet
	}

	return CodeElement{
		Name:      extractName(sym.Symbol),
		Signature: extractSignature(sym),
		CodeType:  extractCodeType(sym),
		Docstring: extractDocstring(sym),
		Line:      line,
		LineFrom:  lineFrom,
		LineTo:    lineTo,
		Context:   extractContext(doc, sym, snippet),
	}, nil
}

// extractName parses the symbol string to extract the symbol's name.
func extractName(symbol string) string {
	// Example symbol format: "kind package.module.name"
	parts := strings.Split(symbol, " ")
	if len(parts) > 1 {
		// Assuming the last part is the fully qualified name
		fullName := parts[len(parts)-1]
		nameParts := strings.Split(fullName, ".")
		return nameParts[len(nameParts)-1]
	}
	return symbol
}

// extractSignature retrieves the signature from the SignatureDocumentation field.
func extractSignature(sym *scip.SymbolInformation) string {
	if sym.SignatureDocumentation != nil {
		return sym.SignatureDocumentation.Text
	}
	return ""
}

// extractCodeType determines the type of the code element based on the SymbolKind.
func extractCodeType(sym *scip.SymbolInformation) string {
	switch sym.Kind {
	//case scip.SymbolInformation_UnspecifiedKind:
	//	return "Unspecified"
	case scip.SymbolInformation_AbstractMethod:
		return "Abstract Method"
	//case scip.SymbolInformation_Accessor:
	//	return "Accessor"
	case scip.SymbolInformation_Array:
		return "Array"
	//case scip.SymbolInformation_Assertion:
	//	return "Assertion"
	//case scip.SymbolInformation_AssociatedType:
	//	return "Associated Type"
	//case scip.SymbolInformation_Attribute:
	//	return "Attribute"
	//case scip.SymbolInformation_Axiom:
	//	return "Axiom"
	//case scip.SymbolInformation_Boolean:
	//	return "Boolean"
	case scip.SymbolInformation_Class:
		return "Class"
	case scip.SymbolInformation_Constant:
		return "Constant"
	//case scip.SymbolInformation_Constructor:
	//	return "Constructor"
	//case scip.SymbolInformation_Contract:
	//	return "Contract"
	//case scip.SymbolInformation_DataFamily:
	//	return "Data Family"
	//case scip.SymbolInformation_Delegate:
	//	return "Delegate"
	case scip.SymbolInformation_Enum:
		return "Enum"
	//case scip.SymbolInformation_EnumMember:
	//	return "Enum Member"
	case scip.SymbolInformation_Error:
		return "Error"
	//case scip.SymbolInformation_Event:
	//	return "Event"
	//case scip.SymbolInformation_Extension:
	//	return "Extension"
	//case scip.SymbolInformation_Fact:
	//	return "Fact"
	//case scip.SymbolInformation_Field:
	//	return "Field"
	//case scip.SymbolInformation_File:
	//	return "File"
	case scip.SymbolInformation_Function:
		return "Function"
	//case scip.SymbolInformation_Getter:
	//	return "Getter"
	//case scip.SymbolInformation_Grammar:
	//	return "Grammar"
	case scip.SymbolInformation_Instance:
		return "Instance"
	case scip.SymbolInformation_Interface:
		return "Interface"
	//case scip.SymbolInformation_Key:
	//	return "Key"
	//case scip.SymbolInformation_Lang:
	//	return "Language"
	//case scip.SymbolInformation_Lemma:
	//	return "Lemma"
	case scip.SymbolInformation_Library:
		return "Library"
	case scip.SymbolInformation_Macro:
		return "Macro"
	case scip.SymbolInformation_Method:
		return "Method"
	case scip.SymbolInformation_MethodAlias:
		return "Method Alias"
	//case scip.SymbolInformation_MethodReceiver:
	//	return "Method Receiver"
	//case scip.SymbolInformation_MethodSpecification:
	//	return "Method Specification"
	//case scip.SymbolInformation_Message:
	//	return "Message"
	//case scip.SymbolInformation_Mixin:
	//	return "Mixin"
	//case scip.SymbolInformation_Modifier:
	//	return "Modifier"
	case scip.SymbolInformation_Module:
		return "Module"
	//case scip.SymbolInformation_Namespace:
	//	return "Namespace"
	//case scip.SymbolInformation_Null:
	//	return "Null"
	//case scip.SymbolInformation_Number:
	//	return "Number"
	case scip.SymbolInformation_Object:
		return "Object"
	//case scip.SymbolInformation_Operator:
	//	return "Operator"
	case scip.SymbolInformation_Package:
		return "Package"
	//case scip.SymbolInformation_PackageObject:
	//	return "Package Object"
	//case scip.SymbolInformation_Parameter:
	//	return "Parameter"
	//case scip.SymbolInformation_ParameterLabel:
	//	return "Parameter Label"
	//case scip.SymbolInformation_Pattern:
	//	return "Pattern"
	//case scip.SymbolInformation_Predicate:
	//	return "Predicate"
	//case scip.SymbolInformation_Property:
	//	return "Property"
	//case scip.SymbolInformation_Protocol:
	//	return "Protocol"
	//case scip.SymbolInformation_ProtocolMethod:
	//	return "Protocol Method"
	//case scip.SymbolInformation_PureVirtualMethod:
	//	return "Pure Virtual Method"
	//case scip.SymbolInformation_Quasiquoter:
	//	return "Quasiquoter"
	//case scip.SymbolInformation_SelfParameter:
	//	return "Self Parameter"
	//case scip.SymbolInformation_Setter:
	//	return "Setter"
	case scip.SymbolInformation_Signature:
		return "Signature"
	//case scip.SymbolInformation_SingletonClass:
	//	return "Singleton Class"
	//case scip.SymbolInformation_SingletonMethod:
	//	return "Singleton Method"
	//case scip.SymbolInformation_StaticDataMember:
	//	return "Static Data Member"
	//case scip.SymbolInformation_StaticEvent:
	//	return "Static Event"
	//case scip.SymbolInformation_StaticField:
	//	return "Static Field"
	//case scip.SymbolInformation_StaticMethod:
	//	return "Static Method"
	//case scip.SymbolInformation_StaticProperty:
	//	return "Static Property"
	case scip.SymbolInformation_StaticVariable:
		return "Static Variable"
	//case scip.SymbolInformation_String:
	//	return "String"
	case scip.SymbolInformation_Struct:
		return "Struct"
	//case scip.SymbolInformation_Subscript:
	//	return "Subscript"
	//case scip.SymbolInformation_Tactic:
	//	return "Tactic"
	//case scip.SymbolInformation_Theorem:
	//	return "Theorem"
	//case scip.SymbolInformation_ThisParameter:
	//	return "This Parameter"
	case scip.SymbolInformation_Trait:
		return "Trait"
	case scip.SymbolInformation_TraitMethod:
		return "Trait Method"
	case scip.SymbolInformation_Type:
		return "Type"
	case scip.SymbolInformation_TypeAlias:
		return "Type Alias"
	case scip.SymbolInformation_TypeClass:
		return "Type Class"
	case scip.SymbolInformation_TypeClassMethod:
		return "Type Class Method"
	//case scip.SymbolInformation_TypeFamily:
	//	return "Type Family"
	//case scip.SymbolInformation_TypeParameter:
	//	return "Type Parameter"
	case scip.SymbolInformation_Union:
		return "Union"
	//case scip.SymbolInformation_Value:
	//	return "Value"
	//case scip.SymbolInformation_Variable:
	//	return "Variable"
	default:
		return "Unknown"
	}
}

// extractDocstring retrieves the docstring from the documentation field.
func extractDocstring(sym *scip.SymbolInformation) string {
	return strings.Join(sym.Documentation, "\n")
}

// extractLineFrom retrieves the starting line from the occurrence with the Definition role.
func extractLineFrom(doc *scip.Document, sym *scip.SymbolInformation) int32 {
	for _, occ := range doc.Occurrences {
		if occ.Symbol == sym.Symbol && hasRole(occ.SymbolRoles, scip.SymbolRole_Definition) {
			if len(occ.Range) > 0 {
				return occ.Range[0]
			}
		}
	}
	return 0
}

// extractLineTo retrieves the ending line from the occurrence with the Definition role.
// If endLine is not provided, it attempts to determine it.
func extractLineTo(doc *scip.Document, sym *scip.SymbolInformation) int32 {
	for _, occ := range doc.Occurrences {
		if occ.Symbol == sym.Symbol && hasRole(occ.SymbolRoles, scip.SymbolRole_Definition) {
			if len(occ.Range) >= 4 {
				return occ.Range[2]
			} else if len(occ.Range) >= 3 {
				// When range has 3 elements, endLine is inferred to be the same as startLine
				return occ.Range[0]
			}
		}
	}
	return 0
}

// extractContext gathers additional context information for the code element.
func extractContext(doc *scip.Document, sym *scip.SymbolInformation, snippet string) map[string]string {
	ctx := map[string]string{
		"file_path": doc.RelativePath,
		"file_name": filepath.Base(doc.RelativePath),
	}

	parts := strings.Split(sym.Symbol, " ")
	if len(parts) > 3 {
		ctx["module"] = parts[3]
	}

	for i, part := range parts {
		if part == "impl" && i+1 < len(parts) {
			ctx["struct_name"] = parts[i+1]
			break
		}
	}

	if snippet != "" {
		ctx["snippet"] = snippet
	}

	return ctx
}

// getSnippet extracts the code snippet from the document based on the symbol's range.
// It adjusts for zero-based indexing and attempts to find the end of multi-line symbols.
func getSnippet(doc *scip.Document, sym *scip.SymbolInformation) (line int32, lineFrom int32, lineTo int32, snippet string, err error) {
	var startLine, endLine int32

	for _, occ := range doc.Occurrences {
		if occ.Symbol == sym.Symbol && hasRole(occ.SymbolRoles, scip.SymbolRole_Definition) {
			if len(occ.Range) >= 2 {
				startLine = occ.Range[0]
				if len(occ.Range) >= 4 {
					endLine = occ.Range[2]
				} else {
					endLine = occ.Range[0]
				}
			}
			break
		}
	}

	if startLine == 0 && endLine == 0 {
		return 0, 0, 0, "", fmt.Errorf("definition occurrence not found or invalid range")
	}

	startLineZero := startLine - 1
	endLineZero := endLine - 1

	if startLineZero < 0 {
		log.Printf("Warning: startLineZero (%d) is less than 0. Adjusting to 0.", startLineZero)
		startLineZero = 0
	}
	if endLineZero < startLineZero {
		log.Printf("Warning: endLineZero (%d) is less than startLineZero (%d). Adjusting to startLineZero.", endLineZero, startLineZero)
		endLineZero = startLineZero
	}

	text, err := getText(doc)
	if err != nil {
		return line, lineFrom, lineTo, snippet, err
	}
	lines := strings.Split(text, "\n")

	if int(startLineZero) >= len(lines) {
		return line, lineFrom, lineTo, snippet, fmt.Errorf("startLineZero (%d) exceeds total lines (%d)", startLineZero, len(lines))
	}

	if endLineZero == startLineZero {
		detectedEndLine, err := findEndLine(lines, int(startLineZero))
		if err != nil {
			log.Printf("Warning: Could not determine end line for symbol %s: %v", sym.Symbol, err)
			detectedEndLine = int(startLineZero)
		}
		endLineZero = int32(detectedEndLine)
	}

	if int(endLineZero) >= len(lines) {
		log.Printf("Warning: endLineZero (%d) exceeds total lines (%d). Adjusting to last line.", endLineZero, len(lines))
		endLineZero = int32(len(lines) - 1)
	}

	if startLineZero < 0 || endLineZero < startLineZero || int(endLineZero) >= len(lines) {
		return line, lineFrom, lineTo, snippet, fmt.Errorf("invalid line range: start=%d, end=%d", startLineZero, endLineZero)
	}

	snippetLines := lines[startLineZero : endLineZero+1]
	snippet = strings.Join(snippetLines, "\n")

	line = startLine
	lineFrom = startLine
	lineTo = endLine

	return
}

// getText retrieves the text from doc.Text or reads it from the file path.
// It also caches the file content to optimize performance.
var fileCache = make(map[string]string)

func getText(doc *scip.Document) (string, error) {
	if text, exists := fileCache[doc.RelativePath]; exists {
		return text, nil
	}

	if len(doc.Text) > 0 {
		fileCache[doc.RelativePath] = doc.Text
		return doc.Text, nil
	}

	absPath, err := filepath.Abs(doc.RelativePath)
	if err != nil {
		return "", fmt.Errorf("unable to determine absolute path: %v", err)
	}
	fileData, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("unable to read file %s: %w", absPath, err)
	}
	text := string(fileData)
	fileCache[doc.RelativePath] = text
	return text, nil
}

// findEndLine attempts to find the end line of a symbol by matching braces.
// This is a simplistic approach and may need enhancements for complex cases.
func findEndLine(lines []string, start int) (int, error) {
	openBraces := 0
	for i := start; i < len(lines); i++ {
		line := lines[i]
		openBraces += strings.Count(line, "{")
		openBraces -= strings.Count(line, "}")
		if openBraces <= 0 && i != start {
			return i, nil
		}
	}
	return len(lines) - 1, fmt.Errorf("could not find closing brace")
}

// hasRole checks if the given symbolRoles bitmask includes the targetRole.
func hasRole(symbolRoles int32, targetRole scip.SymbolRole) bool {
	return symbolRoles&int32(targetRole) != 0
}

// writeJSONOutput writes the extracted code elements to a JSON file.
func writeJSONOutput(filename string, elements []CodeElement) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create JSON output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(elements); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
