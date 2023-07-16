package codegen

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/labstack/echo/v4"

	"github.tesla.cn/itapp/lines/filex"
)

const (
	// These allow the case statements to be sorted later:
	//prefixMostSpecific  = "3"
	//prefixLessSpecific  = "6"
	prefixLeastSpecific = "9"
	responseTypeSuffix  = "RespByPen"
)

const (
	TagXInternalApi = "x-internal-api"
	XExtends        = "x-extends"
	XExtendsUrls    = "urls"
)

var (
	contentTypesJSON = []string{echo.MIMEApplicationJSON, "text/x-json"}
	contentTypesYAML = []string{"application/yaml", "application/x-yaml", "text/yaml", "text/x-yaml"}
	contentTypesXML  = []string{echo.MIMEApplicationXML, echo.MIMETextXML}
)

// This function takes an array of Parameter definition, and generates a valid
// Go parameter declaration from them, eg:
// ", foo int, bar string, baz float32". The preceding comma is there to save
// a lot of work in the template engine.
func genParamArgs(params []ParameterDefinition) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		paramName := p.GoVariableName()
		parts[i] = fmt.Sprintf("%s %s", paramName, p.TypeDef())
	}
	return ", " + strings.Join(parts, ", ")
}

func genParamArgsCustom(params []ParameterDefinition) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		paramName := p.GoVariableName()
		parts[i] = fmt.Sprintf("%s %s", paramName, p.TypeDef())
	}
	return strings.Join(parts, ", ")
}

// This is another variation of the function above which generates only the
// parameter names:
// ", foo, bar, baz"
func genParamNames(params []ParameterDefinition) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = p.GoVariableName()
	}
	return ", " + strings.Join(parts, ", ")
}

func genParamNamesCustom(params []ParameterDefinition) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = p.GoVariableName()
	}
	return strings.Join(parts, ", ")
}

func genParamObject(op OperationDefinition) string {
	if op.RequiresParamObject() {
		paramObject := fmt.Sprintf("params model.%sParams", op.OperationId)
		if len(op.PathParams) > 0 {
			paramObject = ", " + paramObject
		}
		return paramObject
	}
	return ""
}

func genParamObjectName(op OperationDefinition) string {
	if op.RequiresParamObject() {
		paramObjectName := "params"
		if len(op.PathParams) > 0 {
			paramObjectName = ", " + paramObjectName
		}
		return paramObjectName
	}
	return ""
}

func genParamFmtString(path string) string {
	return ReplacePathParamsWithStr(path)
}

// genResponseTypeName creates the name of generated response types (given the operationID):
func genResponseTypeName(operationID string) string {
	return fmt.Sprintf("%s%s", UppercaseFirstCharacter(operationID), responseTypeSuffix)
}

// genResponsePayload generates the payload returned at the end of each client request function
func genResponsePayload(operationID string) string {
	var buffer = bytes.NewBufferString("")

	// Here is where we build up a response:
	_, _ = fmt.Fprintf(buffer, "&%s{\n", genResponseTypeName(operationID))
	_, _ = fmt.Fprintf(buffer, "Body: bodyBytes,\n")
	_, _ = fmt.Fprintf(buffer, "HTTPResponse: rsp,\n")
	_, _ = fmt.Fprintf(buffer, "}")

	return buffer.String()
}

// genResponseUnmarshal generates unmarshaling steps for structured response payloads
func genResponseUnmarshal(op *OperationDefinition) string {
	var buffer = bytes.NewBufferString("")

	var handledCaseClauses = make(map[string]string)
	var unhandledCaseClauses = make(map[string]string)

	// Get the type definitions from the operation:
	typeDefinitions, err := op.GetResponseTypeDefinitions()
	if err != nil {
		panic(err)
	}

	// Add a case for each possible response:
	responses := op.Spec.Responses
	for _, typeDefinition := range typeDefinitions {

		responseRef, ok := responses[typeDefinition.ResponseName]
		if !ok {
			continue
		}

		// We can't do much without a value:
		if responseRef.Value == nil {
			fmt.Fprintf(os.Stderr, "Response %s.%s has nil value\n", op.OperationId, typeDefinition.ResponseName)
			continue
		}

		// If there is no content-type then we have no unmarshaling to do:
		if len(responseRef.Value.Content) == 0 {
			caseAction := "break // No content-type"
			caseClauseKey := "case " + getConditionOfResponseName("rsp.StatusCode", typeDefinition.ResponseName) + ":"
			unhandledCaseClauses[prefixLeastSpecific+caseClauseKey] = fmt.Sprintf("%s\n%s\n", caseClauseKey, caseAction)
			continue
		}

		// If we made it this far then we need to handle unmarshaling for each content-type:
		sortedContentKeys := SortedContentKeys(responseRef.Value.Content)
		for _, contentTypeName := range sortedContentKeys {

			// We get "interface{}" when using "anyOf" or "oneOf" (which doesn't work with Go types):
			if typeDefinition.TypeName == "interface{}" {
				// Unable to unmarshal this, so we leave it out:
				continue
			}

			// Add content-types here (json / yaml / xml etc):
			switch {

			// JSON:
			case StringInArray(contentTypeName, contentTypesJSON):
				var caseAction string

				caseAction = fmt.Sprintf("var dest %s\n"+
					"if err := json.Unmarshal(bodyBytes, &dest); err != nil { \n"+
					" return nil, err \n"+
					"}\n"+
					"response.%s = &dest",
					typeDefinition.Schema.TypeDecl(),
					typeDefinition.TypeName)

				caseKey, caseClause := buildUnmarshalCase(typeDefinition, caseAction, "json")
				handledCaseClauses[caseKey] = caseClause

			// YAML:
			case StringInArray(contentTypeName, contentTypesYAML):
				var caseAction string
				caseAction = fmt.Sprintf("var dest %s\n"+
					"if err := yaml.Unmarshal(bodyBytes, &dest); err != nil { \n"+
					" return nil, err \n"+
					"}\n"+
					"response.%s = &dest",
					typeDefinition.Schema.TypeDecl(),
					typeDefinition.TypeName)
				caseKey, caseClause := buildUnmarshalCase(typeDefinition, caseAction, "yaml")
				handledCaseClauses[caseKey] = caseClause

			// XML:
			case StringInArray(contentTypeName, contentTypesXML):
				var caseAction string
				caseAction = fmt.Sprintf("var dest %s\n"+
					"if err := xml.Unmarshal(bodyBytes, &dest); err != nil { \n"+
					" return nil, err \n"+
					"}\n"+
					"response.%s = &dest",
					typeDefinition.Schema.TypeDecl(),
					typeDefinition.TypeName)
				caseKey, caseClause := buildUnmarshalCase(typeDefinition, caseAction, "xml")
				handledCaseClauses[caseKey] = caseClause

			// Everything else:
			default:
				caseAction := fmt.Sprintf("// Content-type (%s) unsupported", contentTypeName)
				caseClauseKey := "case " + getConditionOfResponseName("rsp.StatusCode", typeDefinition.ResponseName) + ":"
				unhandledCaseClauses[prefixLeastSpecific+caseClauseKey] = fmt.Sprintf("%s\n%s\n", caseClauseKey, caseAction)
			}
		}
	}

	// Now build the switch statement in order of most-to-least specific:
	// See: https://github.com/deepmap/oapi-codegen/issues/127 for why we handle this in two separate
	// groups.
	fmt.Fprintf(buffer, "switch {\n")
	for _, caseClauseKey := range SortedStringKeys(handledCaseClauses) {

		fmt.Fprintf(buffer, "%s\n", handledCaseClauses[caseClauseKey])
	}
	for _, caseClauseKey := range SortedStringKeys(unhandledCaseClauses) {

		fmt.Fprintf(buffer, "%s\n", unhandledCaseClauses[caseClauseKey])
	}
	fmt.Fprintf(buffer, "}\n")

	return buffer.String()
}

// buildUnmarshalCase builds an unmarshalling case clause for different content-types:
func buildUnmarshalCase(typeDefinition TypeDefinition, caseAction string, contentType string) (caseKey string, caseClause string) {
	caseKey = fmt.Sprintf("%s.%s.%s", prefixLeastSpecific, contentType, typeDefinition.ResponseName)
	caseClauseKey := getConditionOfResponseName("rsp.StatusCode", typeDefinition.ResponseName)
	caseClause = fmt.Sprintf("case strings.Contains(rsp.Header.Get(\"%s\"), \"%s\") && %s:\n%s\n", echo.HeaderContentType, contentType, caseClauseKey, caseAction)
	return caseKey, caseClause
}

// Return the statusCode comparison clause from the response name.
func getConditionOfResponseName(statusCodeVar, responseName string) string {
	switch responseName {
	case "default":
		return "true"
	case "1XX", "2XX", "3XX", "4XX", "5XX":
		return fmt.Sprintf("%s / 100 == %s", statusCodeVar, responseName[:1])
	default:
		return fmt.Sprintf("%s == %s", statusCodeVar, responseName)
	}
}

func genReqObject(op OperationDefinition) string {
	if len(op.Bodies) > 0 {
		var reqObject string
		if op.BodyRequired {
			reqObject = fmt.Sprintf("req model.%s", op.Bodies[0].Schema.RefType)
		} else {
			reqObject = fmt.Sprintf("req *model.%s", op.Bodies[0].Schema.RefType)
		}
		if len(op.PathParams) > 0 || op.RequiresParamObject() {
			reqObject = ", " + reqObject
		}
		return reqObject
	}
	return ""
}

func genReqObjectType(op OperationDefinition) string {
	if len(op.Bodies) > 0 {
		return op.Bodies[0].Schema.RefType
	}
	return ""
}

func genReqObjectName(op OperationDefinition) string {
	if len(op.Bodies) > 0 {
		reqObjectName := "req"
		if len(op.PathParams) > 0 || op.RequiresParamObject() {
			reqObjectName = ", " + reqObjectName
		}
		return reqObjectName
	}
	return ""
}

func importModel(op OperationDefinition) bool {
	return op.RequiresParamObject() || op.BodyRequired
}

func getResponseTypeDefinitions(op *OperationDefinition) []TypeDefinition {
	td, err := op.GetResponseTypeDefinitions()
	if err != nil {
		panic(err)
	}
	return td
}

// This outputs a string array
func toStringArray(sarr []string) string {
	return `[]string{"` + strings.Join(sarr, `","`) + `"}`
}

func stripNewLines(s string) string {
	r := strings.NewReplacer("\n", "")
	return r.Replace(s)
}

func bquote(s string) string {
	return fmt.Sprintf("`%s`", s)
}

func isRequiresParamObject(op OperationDefinition) bool {
	return op.RequiresParamObject()
}

var toBeReplace = regexp.MustCompile(`{(.+?)}`)

func toGinPath(s string) string {
	return toBeReplace.ReplaceAllString(s, ":$1")
}

func isInternalRoute(tags []string) bool {
	for _, v := range tags {
		if v == TagXInternalApi {
			return true
		}
	}

	return false
}

// This function map is passed to the template engine, and we can call each
// function here by keyName from the template code.
var TemplateFunctions = template.FuncMap{
	"genParamArgs":               genParamArgs,
	"genParamArgsCustom":         genParamArgsCustom,
	"genParamNames":              genParamNames,
	"genParamNamesCustom":        genParamNamesCustom,
	"genParamObject":             genParamObject,
	"genParamObjectName":         genParamObjectName,
	"genParamFmtString":          genParamFmtString,
	"genReqObject":               genReqObject,
	"genReqObjectType":           genReqObjectType,
	"genReqObjectName":           genReqObjectName,
	"genResponseTypeName":        genResponseTypeName,
	"getResponseTypeDefinitions": getResponseTypeDefinitions,
	"genResponsePayload":         genResponsePayload,
	"genResponseUnmarshal":       genResponseUnmarshal,
	"importModel":                importModel,
	"toStringArray":              toStringArray,
	"stripNewLines":              stripNewLines,
	"lcFirst":                    LowercaseFirstCharacter,
	"ucFirst":                    UppercaseFirstCharacter,
	"bquote":                     bquote,
	"isRequiresParamObject":      isRequiresParamObject,
	"toGinPath":                  toGinPath,
	"GoFileName":                 filex.GoFileName,
	"isInternalRoute":            isInternalRoute,
}
