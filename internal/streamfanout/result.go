package streamfanout

import "errors"

func CollectErrors(input <-chan error) []error {
	out := make([]error, 0)
	for err := range input {
		if err != nil && len(out) == 0 {
			out = append(out, err)
		}
	}
	return out
}

func JoinErrors(input []error) error {
	return errors.Join(input...)
}
