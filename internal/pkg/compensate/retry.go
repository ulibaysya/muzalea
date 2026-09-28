package compensate

import "context"

func CallRetryably(f Func, n int) error {
	var err error
	for range n {
		if err = f(context.TODO()); err == nil {
			return nil
		}
	}
	return err
}
