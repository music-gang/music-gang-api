import { children, Component, createSignal, For, JSXElement, Match, mergeProps, Show, Switch } from "solid-js"

export const newHTMLInputElement = (): HTMLInputElement => {
    return {} as HTMLInputElement
}

export const newHTMLTextAreaElement = (): HTMLTextAreaElement => {
    return {} as HTMLTextAreaElement
}

export const newHTMLSelectElement = (): HTMLSelectElement => {
    return {} as HTMLSelectElement
}

type InputType = "text" | "password" | "email" | "number" | "tel" | "url" | "search" | "date" | "time" | "datetime-local" | "month" | "week" |
    "color" | "file" | "image" | "hidden" | "textarea" | "select" | "option"

const AppInput: Component<{
    id: string,
    ref?: HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement,
    type?: InputType,
    placeholder?: string,
    required?: bool,
    disabled?: bool,
    readonly?: bool,
    label?: string,
    height?: string,
    flow?: "row" | "col" | "row-reverse" | "col-reverse"
    align?: "start" | "center" | "end" | "baseline" | "stretch",
    justify?: "start" | "center" | "end" | "between" | "around" | "evenly",
    message?: JSXElement | string | (() => JSXElement | string),
    error?: bool,
    errorMessage?: JSXElement | string | (() => JSXElement | string),
    success?: bool,
    successMessage?: JSXElement | string,
    value?: string | number | string[],
    min?: string | number,
    step?: string | number,
    rows?: number | string,
    options?: {
        label: string,
        value: string | number | string[] | undefined,
        selected?: bool
    }[],
    validateRule?: (value: string | number | string[]) => bool,
    onEnterKey?: (e: KeyboardEvent) => void,
    onBlur?: (e: FocusEvent) => void,
    onTyping?: (e: KeyboardEvent) => void,
    onChange?: (e: Event) => void,
    onClick?: (e: MouseEvent) => void,
    onInput?: (e: Event) => void,
}> = (props) => {

    const merged = mergeProps({
        ref: newHTMLInputElement(),
        type: "text" as InputType,
        placeholder: "",
        required: false,
        disabled: false,
        readonly: false,
        label: "",
        height: "",
        flow: "col",
        align: "start",
        justify: "start",
        message: null,
        error: false,
        success: false,
        value: null,
        rows: 5,
        options: [],
        validateRule: () => true,
        onEnterKey: () => { },
        onBlur: () => { },
        onTyping: () => { },
        onChange: () => { },
        onClick: () => { },
        onInput: () => { },
    }, props)

    const [touched, setTouched] = createSignal(false)
    const [invalid, setInvalid] = createSignal(false)

    const flowClass = "flex-" + merged.flow
    const alignClass = "items-" + merged.align
    const justifyClass = "justify-" + merged.justify

    const errorMessage = children(() => merged.errorMessage)
    const successMessage = children(() => merged.successMessage)
    const message = children(() => merged.message)

    return (
        <>
            <div class={'flex ' + flowClass + " " + alignClass + " " + justifyClass}>
                <Show when={merged.label}>
                    <label for={merged.id} class={[
                        "cursor-pointer block mb-2 text-sm font-medium text-gray-900 dark:text-gray-300",
                        (merged.error || (touched() && invalid())
                            ? "text-red-700 dark:text-red-500" : ""),
                        (merged.success ? "text-green-700 dark:text-green-500" : "")
                    ].join(" ")}>{merged.label}</label>
                </Show>
                <Switch fallback={<>
                    <input
                        type={merged.type}
                        id={merged.id}
                        ref={merged.ref as HTMLInputElement}
                        onKeyUp={(e) => {
                            if (e.key === "Enter") {
                                merged.onEnterKey(e)
                                return
                            }
                            setInvalid(false)
                            merged.onTyping(e)
                        }}
                        onBlur={(e) => {
                            merged.onBlur(e)
                            setTouched(true)
                            setInvalid(!merged.validateRule(e.currentTarget.value))
                        }}
                        min={merged.min}
                        step={merged.step}
                        value={merged.value ?? ""}
                        onChange={merged.onChange}
                        onClick={merged.onClick}
                        disabled={merged.disabled}
                        readOnly={merged.readonly}
                        class={[
                            "border text-gray-900 text-sm rounded-lg focus:ring-primary-500 focus:border-primary-500 block w-full p-2.5 dark:placeholder-gray-300 dark:text-gray-300 dark:focus:ring-primary-500 dark:focus:border-primary-500 bg-gray-50 border-gray-300 dark:border-gray-600 dark:bg-gray-600",
                            (merged.error || (touched() && invalid())
                                ? "bg-red-50 border-red-500 dark:bg-red-200 dark:border-red-600 dark:placeholder-gray-700 dark:text-gray-700 focus:ring-red-500 focus:border-red-500 dark:focus:ring-red-500"
                                : ""),
                            (merged.success
                                ? "bg-green-50 border-green-500 dark:bg-green-100 dark:border-green-600 dark:placeholder-gray-700 dark:text-gray-700 focus:ring-green-500 focus:border-green-500 dark:focus:ring-green-500"
                                : ""),
                            merged.height,
                        ].join(" ")}
                        placeholder={merged.placeholder}
                        onInput={merged.onInput}
                        required={merged.required} />
                </>}>
                    <Match when={merged.type === "select"}>
                        <select id="countries"
                            class={[
                                "border text-gray-900 text-sm rounded-lg focus:ring-primary-500 focus:border-primary-500 block w-full p-2.5 dark:placeholder-gray-300 dark:text-gray-300 dark:focus:ring-primary-500 dark:focus:border-primary-500 bg-gray-50 border-gray-300 dark:border-gray-600 dark:bg-gray-600",
                                (merged.error || (touched() && invalid())
                                    ? "bg-red-50 border-red-500 dark:bg-red-200 dark:border-red-600 dark:placeholder-gray-700 dark:text-gray-700 focus:ring-red-500 focus:border-red-500 dark:focus:ring-red-500"
                                    : ""),
                                (merged.success
                                    ? "bg-green-50 border-green-500 dark:bg-green-100 dark:border-green-600 dark:placeholder-gray-700 dark:text-gray-700 focus:ring-green-500 focus:border-green-500 dark:focus:ring-green-500"
                                    : ""),
                                merged.height,
                            ].join(" ")}
                            value={merged.value ?? ""}
                            disabled={merged.disabled || merged.readonly}
                            onInput={merged.onInput}
                        >
                            <For each={merged.options}>{(item) => <>
                                <option
                                    value={item.value}
                                    selected={item.selected}
                                >{item.label}</option>
                            </>}</For>
                        </select>
                    </Match>
                    <Match when={merged.type === "textarea"}>
                        <textarea
                            id={merged.id}
                            ref={merged.ref as HTMLTextAreaElement}
                            onKeyUp={(e) => {
                                if (e.key === "Enter") {
                                    merged.onEnterKey(e)
                                    return
                                }
                                setInvalid(false)
                                merged.onTyping(e)
                            }}
                            onBlur={(e) => {
                                merged.onBlur(e)
                                setTouched(true)
                                setInvalid(!merged.validateRule(e.currentTarget.value))
                            }}
                            value={merged.value ?? ""}
                            onChange={merged.onChange}
                            onClick={merged.onClick}
                            disabled={merged.disabled}
                            readOnly={merged.readonly}
                            rows={merged.rows}
                            onInput={merged.onInput}
                            class={[
                                "border text-gray-900 text-sm rounded-lg focus:ring-primary-500 focus:border-primary-500 block w-full p-2.5 dark:placeholder-gray-300 dark:text-gray-300 dark:focus:ring-primary-500 dark:focus:border-primary-500 bg-gray-50 border-gray-300 dark:border-gray-600 dark:bg-gray-600",
                                (merged.error || (touched() && invalid())
                                    ? "bg-red-50 border-red-500 dark:bg-red-200 dark:border-red-600 dark:placeholder-gray-700 dark:text-gray-700 focus:ring-red-500 focus:border-red-500 dark:focus:ring-red-500"
                                    : ""),
                                (merged.success
                                    ? "bg-green-50 border-green-500 dark:bg-green-100 dark:border-green-600 dark:placeholder-gray-700 dark:text-gray-700 focus:ring-green-500 focus:border-green-500 dark:focus:ring-green-500"
                                    : ""),
                                merged.height,
                            ].join(" ")}
                            placeholder={merged.placeholder}
                            required={merged.required}
                        ></textarea>
                    </Match>
                </Switch>
                <Switch fallback={<>
                    <Show when={message()}>
                        <p class="mt-2 text-sm text-primary-600 dark:text-gray-50">{message()}</p>
                    </Show>
                </>}>
                    <Match when={(merged.error || (touched() && invalid())) && merged.errorMessage}>
                        <p class="mt-2 text-sm text-red-600 dark:text-red-500">{errorMessage()}</p>
                    </Match>
                    <Match when={merged.success && merged.successMessage}>
                        <p class="mt-2 text-sm text-green-600 dark:text-green-500">{successMessage()}</p>
                    </Match>
                </Switch>
            </div>
        </>
    )
}

export default AppInput
