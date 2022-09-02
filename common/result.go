package common

// Result is a result of an operation.
// It usefull for the result of an operation and passed to a channel.
type Result struct {
	// Ok is the success result of the operation.
	ok interface{}
	// Err is the error result of the operation.
	err error
}

// Err creates a new Result with an error.
func Err(err error) Result {
	return Result{
		err: err,
	}
}

// Ok creates a new Result with an ok.
func Ok(ok interface{}) Result {
	return Result{
		ok: ok,
	}
}

// IsErr returns true if an error occured.
func (r Result) IsErr() bool {
	return r.err != nil
}

// IsOk returns true if the operation is successful.
func (r Result) IsOk() bool {
	return r.err == nil
}

// Unwrap returns the result of the operation.
// If the operation is not successful, it panics.
// It is raccomanded to check if an error occured before calling this method.
func (r Result) Unwrap() interface{} {
	if r.err != nil {
		panic(r.err)
	}
	return r.ok
}

// UnWrapErr returns the error of the operation.
// If the operation is successful, it panics.
// It is raccomanded to check if an error occured before calling this method.
func (r Result) UnWrapErr() error {
	if r.err == nil {
		panic("unwrap error on ok result")
	}
	return r.err
}

// UnwrapOr returns the result of the operation or the default value.
func (r Result) UnwrapOr(val interface{}) interface{} {
	if r.err != nil {
		return val
	}
	return r.ok
}

func (r Result) UnwrapOrDefault() interface{} {
	return r.UnwrapOr(nil)
}

// UnwrapOrElse returns the result of the operation or, if the operation is not successful, uses the passed closure to return a ok result.
func (r Result) UnwrapOrElse(op func(err error) interface{}) interface{} {
	if r.err != nil {
		return op(r.err)
	}
	return r.ok
}
