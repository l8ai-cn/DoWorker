package runner

func insertPromptAfterFirst(args []string, prompt string) []string {
	if len(args) == 0 {
		return []string{prompt}
	}
	return append([]string{args[0], prompt}, args[1:]...)
}
