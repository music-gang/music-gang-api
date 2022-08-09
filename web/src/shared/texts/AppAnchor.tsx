import { Component, Show } from "solid-js"


const AppAnchor: Component<{
    label: string,
    href?: string,
    onClick?: () => void,
    target?: "_blank" | "_self" | "_parent" | "_top" | "",
    externalLabel?: string,
}> = (props) => {

    const {
        label,
        href = "#",
        onClick = () => { },
        target = "",
        externalLabel = "",
    } = props

    return (
        <>
            <Show when={externalLabel}>
                {externalLabel}
            </Show>&nbsp;<a
                href={href}
                target={target}
                onClick={onClick}
                class="text-sm text-primary-700 hover:underline dark:text-primary-500">{label}</a>
        </>
    )
}

export default AppAnchor