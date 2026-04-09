// Package interactive provides a unified interface for interactive CLI input,
// abstracting away survey/v2 complexity.
//
// The central [Prompter] interface supports text input, yes/no confirmation,
// single-selection lists, and multiline editing. [TerminalPrompter] implements
// it via survey/v2; [NonInteractivePrompter] returns errors for CI/script use.
//
// Helper functions ([SelectProject], [SelectSuite], [SelectRun], [SelectSection])
// wrap API calls and format domain objects into human-readable choice lists.
// Commands retrieve the Prompter from context via [PrompterFromContext],
// enabling clean separation of business logic from user interaction.
package interactive
