package kubernetes

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
)

func Retry[T any](cb func() (T, error)) (T, error) {
	var value T

	err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		if _, ok := err.(errors.APIStatus); !ok {
			return true
		}

		if ok := errors.IsTimeout(err); ok {
			return true
		}

		if ok := errors.IsServerTimeout(err); ok {
			return true
		}

		if ok := errors.IsServiceUnavailable(err); ok {
			return true
		}

		return false
	}, func() error {
		v, err := cb()
		if err != nil {
			return err
		}

		value = v

		return nil
	})

	return value, err
}
