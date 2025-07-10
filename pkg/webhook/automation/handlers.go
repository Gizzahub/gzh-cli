package automation

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v66/github"
	"go.uber.org/zap"
)

// CreateIssueHandler handles creating GitHub issues
type CreateIssueHandler struct {
	client *github.Client
	logger *zap.Logger
}

// NewCreateIssueHandler creates a new issue creation handler
func NewCreateIssueHandler(client *github.Client, logger *zap.Logger) *CreateIssueHandler {
	return &CreateIssueHandler{
		client: client,
		logger: logger,
	}
}

// Execute creates a new issue
func (h *CreateIssueHandler) Execute(ctx context.Context, event *Event, action Action) error {
	if event.Repository == nil {
		return fmt.Errorf("repository information not available")
	}

	title, _ := action.Parameters["title"].(string)
	body, _ := action.Parameters["body"].(string)
	labels, _ := action.Parameters["labels"].([]string)
	assignees, _ := action.Parameters["assignees"].([]string)

	// Template variable substitution
	title = h.substituteVariables(title, event)
	body = h.substituteVariables(body, event)

	issue := &github.IssueRequest{
		Title:     &title,
		Body:      &body,
		Labels:    &labels,
		Assignees: &assignees,
	}

	owner := event.Repository.GetOwner().GetLogin()
	repo := event.Repository.GetName()

	createdIssue, _, err := h.client.Issues.Create(ctx, owner, repo, issue)
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	h.logger.Info("Created issue",
		zap.String("repo", repo),
		zap.Int("number", createdIssue.GetNumber()),
		zap.String("title", title))

	return nil
}

// ValidateParameters validates the action parameters
func (h *CreateIssueHandler) ValidateParameters(params map[string]interface{}) error {
	if _, ok := params["title"]; !ok {
		return fmt.Errorf("title parameter is required")
	}
	if _, ok := params["body"]; !ok {
		return fmt.Errorf("body parameter is required")
	}
	return nil
}

// substituteVariables replaces template variables with event data
func (h *CreateIssueHandler) substituteVariables(template string, event *Event) string {
	replacements := map[string]string{
		"{{event.type}}":     event.Type,
		"{{event.action}}":   event.Action,
		"{{event.id}}":       event.ID,
		"{{repo.name}}":      event.Repository.GetName(),
		"{{repo.full_name}}": event.Repository.GetFullName(),
		"{{sender.login}}":   event.Sender.GetLogin(),
	}

	result := template
	for key, value := range replacements {
		result = strings.ReplaceAll(result, key, value)
	}

	return result
}

// AddLabelHandler handles adding labels to issues/PRs
type AddLabelHandler struct {
	client *github.Client
	logger *zap.Logger
}

// NewAddLabelHandler creates a new label handler
func NewAddLabelHandler(client *github.Client, logger *zap.Logger) *AddLabelHandler {
	return &AddLabelHandler{
		client: client,
		logger: logger,
	}
}

// Execute adds labels to an issue or PR
func (h *AddLabelHandler) Execute(ctx context.Context, event *Event, action Action) error {
	if event.Repository == nil {
		return fmt.Errorf("repository information not available")
	}

	labels, ok := action.Parameters["labels"].([]string)
	if !ok {
		return fmt.Errorf("labels parameter must be a string array")
	}

	number, ok := h.getIssueOrPRNumber(event)
	if !ok {
		return fmt.Errorf("could not determine issue/PR number from event")
	}

	owner := event.Repository.GetOwner().GetLogin()
	repo := event.Repository.GetName()

	_, _, err := h.client.Issues.AddLabelsToIssue(ctx, owner, repo, number, labels)
	if err != nil {
		return fmt.Errorf("failed to add labels: %w", err)
	}

	h.logger.Info("Added labels",
		zap.String("repo", repo),
		zap.Int("number", number),
		zap.Strings("labels", labels))

	return nil
}

// ValidateParameters validates the action parameters
func (h *AddLabelHandler) ValidateParameters(params map[string]interface{}) error {
	if _, ok := params["labels"]; !ok {
		return fmt.Errorf("labels parameter is required")
	}
	return nil
}

// getIssueOrPRNumber extracts issue/PR number from event
func (h *AddLabelHandler) getIssueOrPRNumber(event *Event) (int, bool) {
	// This would need to inspect the event payload based on event type
	// For now, return a placeholder
	return 0, false
}

// CreateCommentHandler handles adding comments to issues/PRs
type CreateCommentHandler struct {
	client *github.Client
	logger *zap.Logger
}

// NewCreateCommentHandler creates a new comment handler
func NewCreateCommentHandler(client *github.Client, logger *zap.Logger) *CreateCommentHandler {
	return &CreateCommentHandler{
		client: client,
		logger: logger,
	}
}

// Execute adds a comment to an issue or PR
func (h *CreateCommentHandler) Execute(ctx context.Context, event *Event, action Action) error {
	if event.Repository == nil {
		return fmt.Errorf("repository information not available")
	}

	body, ok := action.Parameters["body"].(string)
	if !ok {
		return fmt.Errorf("body parameter is required")
	}

	number, ok := h.getIssueOrPRNumber(event)
	if !ok {
		return fmt.Errorf("could not determine issue/PR number from event")
	}

	// Template variable substitution
	handler := &CreateIssueHandler{} // Reuse substitution logic
	body = handler.substituteVariables(body, event)

	owner := event.Repository.GetOwner().GetLogin()
	repo := event.Repository.GetName()

	comment := &github.IssueComment{
		Body: &body,
	}

	_, _, err := h.client.Issues.CreateComment(ctx, owner, repo, number, comment)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	h.logger.Info("Created comment",
		zap.String("repo", repo),
		zap.Int("number", number))

	return nil
}

// ValidateParameters validates the action parameters
func (h *CreateCommentHandler) ValidateParameters(params map[string]interface{}) error {
	if _, ok := params["body"]; !ok {
		return fmt.Errorf("body parameter is required")
	}
	return nil
}

// getIssueOrPRNumber extracts issue/PR number from event
func (h *CreateCommentHandler) getIssueOrPRNumber(event *Event) (int, bool) {
	// This would need to inspect the event payload based on event type
	// For now, return a placeholder
	return 0, false
}

// MergePRHandler handles automatic PR merging
type MergePRHandler struct {
	client *github.Client
	logger *zap.Logger
}

// NewMergePRHandler creates a new PR merge handler
func NewMergePRHandler(client *github.Client, logger *zap.Logger) *MergePRHandler {
	return &MergePRHandler{
		client: client,
		logger: logger,
	}
}

// Execute merges a pull request
func (h *MergePRHandler) Execute(ctx context.Context, event *Event, action Action) error {
	if event.Repository == nil {
		return fmt.Errorf("repository information not available")
	}

	prNumber, ok := h.getPRNumber(event)
	if !ok {
		return fmt.Errorf("could not determine PR number from event")
	}

	owner := event.Repository.GetOwner().GetLogin()
	repo := event.Repository.GetName()

	// Get merge method from parameters
	mergeMethod, _ := action.Parameters["merge_method"].(string)
	if mergeMethod == "" {
		mergeMethod = "merge" // default
	}

	commitMessage, _ := action.Parameters["commit_message"].(string)

	options := &github.PullRequestOptions{
		MergeMethod: mergeMethod,
	}

	result, _, err := h.client.PullRequests.Merge(ctx, owner, repo, prNumber, commitMessage, options)
	if err != nil {
		return fmt.Errorf("failed to merge PR: %w", err)
	}

	h.logger.Info("Merged pull request",
		zap.String("repo", repo),
		zap.Int("number", prNumber),
		zap.String("sha", result.GetSHA()))

	return nil
}

// ValidateParameters validates the action parameters
func (h *MergePRHandler) ValidateParameters(params map[string]interface{}) error {
	if method, ok := params["merge_method"].(string); ok {
		validMethods := []string{"merge", "squash", "rebase"}
		valid := false
		for _, vm := range validMethods {
			if method == vm {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid merge_method: must be one of merge, squash, rebase")
		}
	}
	return nil
}

// getPRNumber extracts PR number from event
func (h *MergePRHandler) getPRNumber(event *Event) (int, bool) {
	// This would need to inspect the event payload based on event type
	// For now, return a placeholder
	return 0, false
}

// NotificationHandler handles sending notifications
type NotificationHandler struct {
	logger   *zap.Logger
	webhooks map[string]string // notification type -> webhook URL
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		logger:   logger,
		webhooks: make(map[string]string),
	}
}

// RegisterWebhook registers a webhook URL for a notification type
func (h *NotificationHandler) RegisterWebhook(notificationType, url string) {
	h.webhooks[notificationType] = url
}

// Execute sends a notification
func (h *NotificationHandler) Execute(ctx context.Context, event *Event, action Action) error {
	notificationType, _ := action.Parameters["type"].(string)
	if notificationType == "" {
		notificationType = "default"
	}

	message, _ := action.Parameters["message"].(string)
	if message == "" {
		return fmt.Errorf("message parameter is required")
	}

	// Template variable substitution
	handler := &CreateIssueHandler{} // Reuse substitution logic
	message = handler.substituteVariables(message, event)

	// Send notification based on type
	webhookURL, exists := h.webhooks[notificationType]
	if !exists {
		return fmt.Errorf("no webhook configured for notification type: %s", notificationType)
	}

	// Here you would implement the actual notification sending logic
	// For example, sending to Slack, Discord, email, etc.
	h.logger.Info("Sent notification",
		zap.String("type", notificationType),
		zap.String("webhook", webhookURL),
		zap.String("message", message))

	return nil
}

// ValidateParameters validates the action parameters
func (h *NotificationHandler) ValidateParameters(params map[string]interface{}) error {
	if _, ok := params["message"]; !ok {
		return fmt.Errorf("message parameter is required")
	}
	return nil
}

// RunWorkflowHandler handles triggering GitHub Actions workflows
type RunWorkflowHandler struct {
	client *github.Client
	logger *zap.Logger
}

// NewRunWorkflowHandler creates a new workflow handler
func NewRunWorkflowHandler(client *github.Client, logger *zap.Logger) *RunWorkflowHandler {
	return &RunWorkflowHandler{
		client: client,
		logger: logger,
	}
}

// Execute triggers a workflow
func (h *RunWorkflowHandler) Execute(ctx context.Context, event *Event, action Action) error {
	if event.Repository == nil {
		return fmt.Errorf("repository information not available")
	}

	workflowFile, ok := action.Parameters["workflow_file"].(string)
	if !ok {
		return fmt.Errorf("workflow_file parameter is required")
	}

	ref, _ := action.Parameters["ref"].(string)
	if ref == "" {
		ref = event.Repository.GetDefaultBranch()
	}

	inputs, _ := action.Parameters["inputs"].(map[string]interface{})

	owner := event.Repository.GetOwner().GetLogin()
	repo := event.Repository.GetName()

	// Create workflow dispatch event
	_, err := h.client.Actions.CreateWorkflowDispatchEventByFileName(
		ctx, owner, repo, workflowFile,
		github.CreateWorkflowDispatchEventRequest{
			Ref:    ref,
			Inputs: inputs,
		})
	if err != nil {
		return fmt.Errorf("failed to trigger workflow: %w", err)
	}

	h.logger.Info("Triggered workflow",
		zap.String("repo", repo),
		zap.String("workflow", workflowFile),
		zap.String("ref", ref))

	return nil
}

// ValidateParameters validates the action parameters
func (h *RunWorkflowHandler) ValidateParameters(params map[string]interface{}) error {
	if _, ok := params["workflow_file"]; !ok {
		return fmt.Errorf("workflow_file parameter is required")
	}
	return nil
}
