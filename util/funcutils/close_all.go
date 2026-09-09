package funcutils

import "errors"

func CloseAll(closers ...func() error) error {
	var errs []error
	for _, closer := range closers {
		err := closer()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
