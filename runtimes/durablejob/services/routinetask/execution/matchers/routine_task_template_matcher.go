package matchers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type RoutineTaskTemplateMatcherInterface interface {
	MatchString(value string, values map[string]string) string
	MatchArborizedEditableBlock(block cblocknote.ArborizedEditableBlock, values map[string]string) (cblocknote.ArborizedEditableBlock, *cexceptions.Exception)
}

type RoutineTaskTemplateMatcher struct{}

func NewRoutineTaskTemplateMatcher() RoutineTaskTemplateMatcherInterface {
	return RoutineTaskTemplateMatcher{}
}

func (m RoutineTaskTemplateMatcher) MatchString(value string, values map[string]string) string {
	if len(values) == 0 || !strings.Contains(value, "{{") {
		return value
	}
	matched := value
	for key, resolvedValue := range values {
		matched = strings.ReplaceAll(matched, "{{"+key+"}}", resolvedValue)
	}
	return matched
}

func (m RoutineTaskTemplateMatcher) MatchArborizedEditableBlock(
	block cblocknote.ArborizedEditableBlock,
	values map[string]string,
) (cblocknote.ArborizedEditableBlock, *cexceptions.Exception) {
	matchedChildren := make([]cblocknote.ArborizedEditableBlock, len(block.Children))
	for index, child := range block.Children {
		matchedChild, exception := m.MatchArborizedEditableBlock(child, values)
		if exception != nil {
			return cblocknote.ArborizedEditableBlock{}, exception
		}
		matchedChildren[index] = matchedChild
	}
	block.Children = matchedChildren

	rawBlock, err := json.Marshal(block)
	if err != nil {
		return cblocknote.ArborizedEditableBlock{}, cexceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	var blockMap map[string]any
	if err := json.Unmarshal(rawBlock, &blockMap); err != nil {
		return cblocknote.ArborizedEditableBlock{}, cexceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	shouldMatch := false
	if props, ok := blockMap["props"].(map[string]any); ok {
		if template, ok := props["template"].(bool); ok && template {
			shouldMatch = true
		}
		delete(props, "template")
	}

	if shouldMatch {
		if props, exists := blockMap["props"]; exists {
			blockMap["props"] = m.matchJSONValue(props, values)
		}
		if content, exists := blockMap["content"]; exists {
			blockMap["content"] = m.matchJSONValue(content, values)
		}
	}

	rawMatchedBlock, err := json.Marshal(blockMap)
	if err != nil {
		return cblocknote.ArborizedEditableBlock{}, cexceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	var matchedBlock cblocknote.ArborizedEditableBlock
	if err := json.Unmarshal(rawMatchedBlock, &matchedBlock); err != nil {
		return cblocknote.ArborizedEditableBlock{}, cexceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).
			WithOrigin(fmt.Errorf("invalid matched template block: %w", err))
	}

	return matchedBlock, nil
}

func (m RoutineTaskTemplateMatcher) matchJSONValue(value any, values map[string]string) any {
	switch typed := value.(type) {
	case string:
		return m.MatchString(typed, values)
	case []any:
		matched := make([]any, len(typed))
		for index, item := range typed {
			matched[index] = m.matchJSONValue(item, values)
		}
		return matched
	case map[string]any:
		matched := make(map[string]any, len(typed))
		for key, item := range typed {
			matched[key] = m.matchJSONValue(item, values)
		}
		return matched
	default:
		return value
	}
}
