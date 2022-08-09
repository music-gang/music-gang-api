import { children, Component, JSX } from "solid-js"

const Container: Component<{ children: JSX.Element }> = (props) => {
    const c = children(() => props.children)
    return (
        <>
            {c()}
        </>
    )
}

export default Container