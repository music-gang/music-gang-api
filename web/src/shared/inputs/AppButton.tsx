import { children, Component, JSXElement, mergeProps } from "solid-js"

export type ButtonType = "button" | "submit" | "reset" | undefined
export type ButtonColor = "red" | "primary" | "light" | "green"
type ButtonSize = "xs" | "sm" | "md" | "lg" | "xl"


const AppButton: Component<{
    label: string | JSXElement | (() => JSXElement | string),
    type?: ButtonType,
    color?: ButtonColor,
    disabled?: bool,
    size?: ButtonSize,
    title?: string,
    onClick?: (e: MouseEvent) => void,

}> = (props) => {

    const merged = mergeProps({
        type: "button" as ButtonType,
        color: "primary" as ButtonColor,
        disabled: false,
        size: "sm" as ButtonSize,
        title: "",
    }, props)

    const label = children(() => merged.label)

    let color = ""
    switch (merged.color) {
        case "red":
            color = "text-white bg-red-600 hover:bg-red-800 focus:ring-red-300 dark:focus:ring-red-800"
            break

        case "light":
            color = "text-gray-500 bg-white hover:bg-gray-100 focus:ring-gray-200 border border-gray-200 hover:text-gray-900 dark:bg-gray-700 dark:text-gray-300 dark:border-gray-500 dark:hover:text-white dark:hover:bg-gray-600 dark:focus:ring-gray-600"
            break

        case "green":
            color = "text-white bg-green-600 hover:bg-green-800 focus:ring-green-300 dark:focus:ring-green-800"
            break

        case "primary":
        default:
            color = "text-white bg-primary-600 hover:bg-primary-800 focus:ring-primary-300 dark:bg-primary-600 dark:hover:bg-primary-700 dark:focus:ring-primary-800"
            break
    }

    let size = ""
    switch (merged.size) {
        case "xs":
            size = "px-3 py-2 text-xs font-medium text-center text-white rounded-lg focus:ring-4 focus:outline-none"
            break
        case "sm":
            size = "px-4 py-2 text-sm font-medium text-center text-white rounded-lg focus:ring-4 focus:outline-none"
            break
        case "md":
            size = "px-4 py-2 text-md font-medium text-center text-white rounded-md focus:ring-4 focus:outline-none"
            break
        case "lg":
            size = "px-6 py-3 text-lg font-medium text-center text-white rounded-lg focus:ring-4 focus:outline-none"
            break
        case "xl":
            size = "px-8 py-4 text-xl font-medium text-center text-white rounded-xl focus:ring-4 focus:outline-none"
            break
    }

    return (
        <>
            <button
                type={merged.type}
                onclick={merged.onClick}
                disabled={merged.disabled}
                title={merged.title}
                class={[
                    merged.disabled ? "cursor-not-allowed" : "",
                    size,
                    color
                ].join(" ")}>{label()}</button>
        </>
    )
}

export default AppButton