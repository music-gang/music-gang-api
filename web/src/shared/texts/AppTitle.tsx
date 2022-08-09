import { Component } from "solid-js"


export const AppH1: Component<{
    label: string,
}> = (props) => {

    const {
        label,
    } = props

    return (
        <>
            <h1 class="mb-4 text-xl font-medium text-gray-900 dark:text-white">{label}</h1>
        </>
    )
}

export const AppH3: Component<{
    label: string,
}> = (props) => {

    const {
        label,
    } = props

    return (
        <>
            <h3 class="mb-4 text-xl font-medium text-gray-900 dark:text-white">{label}</h3>
        </>
    )
}

