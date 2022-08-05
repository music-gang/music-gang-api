
export type Option<T> = T | null

export enum ErrorCode {
    ECONFLICT = "conflict",        // conflict with current state
    EINTERNAL = "internal",        // internal error
    EINVALID = "invalid",         // invalid input
    ENOTFOUND = "not_found",       // resource not found
    ENOTIMPLEMENTED = "not_implemented", // feature not implemented
    EUNAUTHORIZED = "unauthorized",    // access denied
    EUNKNOWN = "unknown",         // unknown error
    EFORBIDDEN = "forbidden",       // access forbidden
    EEXISTS = "exists",          // resource already exists
    EMAXATTEMPTS = "max_attempts", // max attempts reached
    ENOTAUTHENTICATED = "not_authenticated", // not authenticated

    EMGVM = "mgvm",                // error code prefix for music gang virtual machine, it is assimilated to EINTERNAL
    EMGVM_LOWFUEL = "low_fuel",            // subcode for EMGVM, low fuel
    EMGVM_CORE_POOL_NOT_FOUND = "core_pool_not_found", // subcode for EMGVM, core pool not found
    EMGVM_CORE_POOL_TIMEOUT = "core_pool_timeout",   // subcode for EMGVM, core pool timeout

    EANCHORAGE = "anchorage" // error code prefix for anchorage contract executor, it is assimilated to EINTERNAL
}

export interface Error {
    code: ErrorCode
    message: string
    details: Option<string[]>
}

export const newError = (code: ErrorCode, message: string, details: Option<string[]> = null): Error => ({
    code,
    message,
    details
})

/**
 * Defines a result type that can be either a success or an error.
 */
export interface Result<T, E> {
    /**
     * Returns `true` if the result is a success.
     */
    isOk(): bool
    /**
     * Returns `true` if the result is an error.
     */
    isErr(): bool
    /**
     * Maps the success value of the result to a new value.
     * @param fn A function that takes the result value and returns a new result.
     */
    map<U>(fn: (value: T) => U): Result<U, E>
    /**
     * Maps the error value of the result to a new value.
     * @param fn A function that takes the result error and returns a new result.
     */
    mapErr<U>(fn: (value: E) => U): Result<T, U>
    /**
     * Transforms the result into a new result, if result is an error then returns the default value.
     * @param defaultValue The default value to return if the result is an error.
     * @param fn A function that takes the success value of the result and returns a new result.
     */
    mapOr<U>(defaultValue: U, fn: (value: T) => U): U
    /**
     * Transforms the result into a new result, if result is an error then uses the fallback function to create a new result, otherwise uses the fn function to create a new result.
     * @param fallback A function that takes the error value of the result and returns a new result.
     * @param fn A function that takes the success value of the result and returns a new result.
     */
    mapOrElse<U>(fallback: (err: E) => U, fn: (value: T) => U): U
    /**
     * Returns the contained success value or the default value.
     * It can thorw an error if the result is an error.
     */
    unwrap(): T
    /**
     * Returns the contained error value or the default value.
     * It can thorw an error if the result is a success.
     */
    unwrapErr(): E
    /**
     * Returns the contained success value or the default value.
     */
    unwrapOr(defaultValue: T): T
    /**
     * Returns the contained success value or applies the fallback function to the error value.
     */
    unwrapOrElse(fn: (err: E) => T): T
}

/**
 * Creates a new success result.
 * @param value The value to return if the result is a success.
 * @returns {Result<T, E>} A result that is a success with the given value.
 */
export const Ok = <T, E = Error>(value: T): Result<T, E> => ({
    isOk: () => true,
    isErr: () => false,
    map: <U>(fn: (value: T) => U): Result<U, E> => Ok(fn(value)),
    mapErr: <U>(): Result<T, U> => Ok(value),
    mapOr: <U>(defaultValue: U, fn: (value: T) => U): U => fn(value),
    mapOrElse: <U>(fallback: (err: E) => U, fn: (value: T) => U): U => fn(value),
    unwrap: () => value,
    unwrapErr: () => {
        throw new Error("Called unwrapError on Ok")
    },
    unwrapOr: () => value,
    unwrapOrElse: () => value
})

/**
 * Creates a new error result.
 * @param error The error to return if the result is an error.
 * @returns {Result<T, E>} A result that is an error with the given error.
 */
export const Err = <T, E = Error>(error: E): Result<T, E> => ({
    isOk: () => false,
    isErr: () => true,
    map: <U>(): Result<U, E> => Err(error),
    mapErr: <U>(fn: (value: E) => U): Result<T, U> => Err(fn(error)),
    mapOr: <U>(defaultValue: U): U => defaultValue,
    mapOrElse: <U>(fallback: (err: E) => U): U => fallback(error),
    unwrap: () => {
        throw new Error("Called unwrap on Err")
    },
    unwrapErr: () => error,
    unwrapOr: (defaultValue: T) => defaultValue,
    unwrapOrElse: (fn: (err: E) => T) => fn(error)
})

export type PaginateResponse<T> = {
    data: Array<T>
    total_results: number
    current_page: number
    items_per_page: number
}