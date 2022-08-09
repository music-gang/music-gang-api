import { children, Component, JSXElement, mergeProps } from "solid-js"


const AppText: Component<{ class?: string, text: string | JSXElement }> = (props) => {

    const merged = mergeProps({ text: "", class: "" }, props)

    const childText = children(() => merged.text)

    let defaultTextClass = "text-gray-900 dark:text-gray-50"

    if (merged.class) {
        defaultTextClass += " " + merged.class
    }

    return (
        <>
            <span class={defaultTextClass}>{childText()}</span>
        </>
    )
}

export default AppText