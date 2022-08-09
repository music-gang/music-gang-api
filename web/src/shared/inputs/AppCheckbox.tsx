import { Component, mergeProps, Show } from "solid-js"
import { newHTMLInputElement } from "./AppInput"

const AppCheckbox: Component<{
    id: string,
    ref?: HTMLInputElement,
    label?: string,
    checked?: bool,
    onInput?: (e: Event) => void,
}> = (props) => {

    const merged = mergeProps({
        label: "",
        ref: newHTMLInputElement(),
        checked: false,
        onInput: () => { },
    }, props)

    return (
        <>
            <div class="flex items-start">
                <div class="flex items-center h-5">
                    <input
                        id={merged.id}
                        type="checkbox"
                        checked={merged.checked}
                        ref={merged.ref}
                        onInput={merged.onInput}
                        class="cursor-pointer w-4 h-4 bg-gray-50 rounded border border-gray-300 focus:ring-3 focus:ring-primary-300 dark:bg-gray-600 dark:border-gray-500 dark:focus:ring-primary-600 dark:ring-offset-gray-800"
                    />
                </div>
                <Show when={merged.label}>
                    <label
                        for={merged.id}
                        class="cursor-pointer ml-2 text-sm font-medium text-gray-900 dark:text-gray-300"
                    >{merged.label}</label>
                </Show>
            </div>
        </>
    )
}

export default AppCheckbox