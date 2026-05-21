// Package claude provides an AI adapter that uses the Anthropic Claude API
// to extract structured data from raw CV text via the tool-use feature.
package claude

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"backend/internal/domain"
	"backend/pkg/log"
)

// ErrFillOutCV is returned when Claude fails to parse the CV for any reason.
var ErrFillOutCV = errors.New("failed to fill out CV using Claude")

// AgentClaude implements the interfaces.Agent port using the Anthropic Claude API.
type AgentClaude struct {
	client *anthropic.Client
	log    log.Logger
}

// NewAgentClaude creates a new AgentClaude.
// apiKey is required; baseURL is optional and used to override the Anthropic API endpoint (e.g. for testing).
func NewAgentClaude(apiKey, baseURL string, logger log.Logger) (*AgentClaude, error) {
	if apiKey == "" {
		return nil, errors.New("apiKey is empty")
	}

	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	client := anthropic.NewClient(opts...)
	return &AgentClaude{
		client: &client,
		log:    logger,
	}, nil
}

// cvData mirrors the domain.CV fields that Claude should populate.
// cvData is the internal DTO that maps Claude's tool-use JSON response to Go fields.
type cvData struct {
	FirstName      string   `json:"first_name"`
	LastName       string   `json:"last_name"`
	CVTitle        string   `json:"cv_title"`
	Specialization string   `json:"specialization"`
	Skills         []string `json:"skills"`
	WorkExperience int      `json:"work_experience_years"`
}

// fillCVTool is the Claude tool definition passed on every request.
// Claude is forced to call this tool to return structured CV data instead of free-form text.
var fillCVTool = anthropic.ToolParam{
	Name:        "fill_cv",
	Description: anthropic.String("Extract structured information from a CV raw text and return it as structured data."),
	InputSchema: anthropic.ToolInputSchemaParam{
		Properties: map[string]any{
			"first_name": map[string]any{
				"type":        "string",
				"description": "Candidate's first name.",
			},
			"last_name": map[string]any{
				"type":        "string",
				"description": "Candidate's last name.",
			},
			"cv_title": map[string]any{
				"type":        "string",
				"description": "The title or headline of the CV (e.g. 'Senior Software Engineer').",
			},
			"specialization": map[string]any{
				"type":        "string",
				"description": "Primary professional specialization or domain (e.g. 'Backend Development', 'Data Science').",
			},
			"skills": map[string]any{
				"type":        "array",
				"description": "List of technical and professional skills mentioned in the CV.",
				"items": map[string]any{
					"type": "string",
				},
			},
			"work_experience_years": map[string]any{
				"type":        "integer",
				"description": "Total years of work experience inferred from the CV.",
			},
		},
	},
}

// FillOutCv sends the CV's raw text to Claude and populates the structured fields in-place.
// It uses tool-use to force a structured JSON response; returns ErrFillOutCV if Claude
// does not call the expected tool or if the response cannot be unmarshalled.
func (a *AgentClaude) FillOutCv(ctx context.Context, cv *domain.CV) error {
	if cv.RawText == "" {
		return errors.Join(ErrFillOutCV, errors.New("CV raw text is empty"))
	}

	prompt := fmt.Sprintf(
		"You are an expert HR assistant. Extract all available information from the following CV text and call the fill_cv tool with the extracted data. If a field cannot be determined from the text, use an empty string or zero.\n\nCV text:\n%s",
		cv.RawText,
	)

	tools := []anthropic.ToolUnionParam{
		{OfTool: &fillCVTool},
	}

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Tools: tools,
	})
	if err != nil {
		a.log.Error("failed to call Claude API",
			log.FieldLogger{Key: "err", Value: err},
		)
		return errors.Join(ErrFillOutCV, err)
	}

	for _, block := range message.Content {
		toolUse, ok := block.AsAny().(anthropic.ToolUseBlock)
		if !ok || toolUse.Name != "fill_cv" {
			continue
		}

		var data cvData
		if err := json.Unmarshal([]byte(toolUse.JSON.Input.Raw()), &data); err != nil {
			a.log.Error("failed to unmarshal Claude tool response",
				log.FieldLogger{Key: "err", Value: err},
			)
			return errors.Join(ErrFillOutCV, err)
		}

		cv.FirstName = data.FirstName
		cv.LastName = data.LastName
		cv.CVTitle = data.CVTitle
		cv.Specialization = data.Specialization
		cv.WorkExperience = data.WorkExperience

		cv.Skills = make([]domain.Skill, len(data.Skills))
		for i, name := range data.Skills {
			cv.Skills[i] = domain.Skill{Name: name}
		}

		return nil
	}

	return errors.Join(ErrFillOutCV, errors.New("claude did not call the fill_cv tool"))
}
