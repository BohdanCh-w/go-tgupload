package utils

import (
	"context"

	"github.com/Songmu/prompter"
)

func RepeatPrompt(ctx context.Context, prompt, def string) (string, error) {
	ch := make(chan string)
	defer close(ch)

	for {
		go func() {
			ch <- prompter.Prompt(prompt, def)
		}()

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case res := <-ch:
			if res != "" {
				return res, nil
			}
		}
	}
}
