import { children, For, JSXElement, Match, mergeProps, Switch } from "solid-js"


type AppListProps<T> = {
    childs?: readonly T[]
    empty?: JSXElement | string
    onItemClick?: (item: T, index: number) => void
    builder?: (item: T, index: number) => JSXElement
    width?: string
}

const AppList = <T,>(props: AppListProps<T>) => {

    const merged = mergeProps({
        empty: "",
        childs: [],
        onItemClick: () => { },
        builder: (item: unknown) => <>{item}</>,
        width: "w-72",
    }, props)

    const empty = children(() => merged.empty)

    const defaultClassFirstItem = "w-full px-4 py-2 border-b border-gray-200 rounded-t-lg dark:border-gray-600"
    const defaultClassMidItem = "w-full px-4 py-2 border-b border-gray-200 dark:border-gray-600"
    const defaultClassLastItem = "w-full px-4 py-2 rounded-b-lg"

    return (
        <>
            <Switch>
                <Match when={merged.childs.length === 0}>
                    {empty()}
                </Match>
                <Match when={merged.childs.length > 0}>
                    <ul
                        class={[
                            "text-sm font-medium text-gray-900 bg-white border border-gray-200 rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white",
                            merged.width
                        ].join(" ")
                        }>
                        <For each={merged.childs}>{(item, index) =>
                            <Switch>
                                <Match when={index() === 0}>
                                    <li
                                        onClick={() => { merged.onItemClick(item, index()) }}
                                        class={[
                                            defaultClassFirstItem,
                                            "cursor-pointer"
                                        ].join(" ")
                                        }>{merged.builder(item, index())}</li>
                                </Match>
                                <Match when={index() > 0 && index() < merged.childs.length - 1}>
                                    <li
                                        onClick={() => { merged.onItemClick(item, index()) }}
                                        class={[
                                            defaultClassMidItem,
                                            "cursor-pointer"
                                        ].join(" ")}>{merged.builder(item, index())}</li>
                                </Match>
                                <Match when={index() === merged.childs.length - 1}>
                                    <li
                                        onClick={() => { merged.onItemClick(item, index()) }}
                                        class={[
                                            defaultClassLastItem,
                                            "cursor-pointer"
                                        ].join(" ")}>{merged.builder(item, index())}</li>
                                </Match>
                            </Switch>
                        }</For>
                    </ul>
                </Match>
            </Switch>
        </>
    )
}

export default AppList