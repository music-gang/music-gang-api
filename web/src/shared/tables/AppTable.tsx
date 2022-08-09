import { children, For, JSXElement, Match, mergeProps, Switch } from "solid-js"


type AppTableProps<T> = {
    childs?: readonly T[]
    empty?: JSXElement | string
    headers?: Array<{ label: string, hidden: bool, width?: string }>
    onItemClick?: (item: T, index: number) => void
    builder?: (item: T, index: number) => JSXElement
    width?: string,
    hoverable?: bool
}

const AppTable = <T,>(props: AppTableProps<T>) => {

    const merged = mergeProps({
        empty: "No items found",
        headers: [],
        childs: [],
        onItemClick: null,
        builder: (item: unknown) => <>{item}</>,
        width: "w-72",
        hoverable: true,
    }, props)

    const empty = children(() => merged.empty)

    return (
        <>
            <Switch fallback={empty()}>
                <Match when={merged.childs.length > 0}>
                    <div class={["relative overflow-x-auto overflow-y-auto shadow-sm sm:rounded-lg", merged.width]
                        .join(" ")}>
                        <table class="w-full table-fixed text-sm text-left text-gray-500 dark:text-gray-400">
                            <thead class="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-800 dark:text-gray-400">
                                <tr>
                                    <For each={merged.headers}>{(item) =>
                                        <th scope="col" class={[
                                            "px-6 py-3",
                                            item.width ? item.width : "",
                                        ].join(" ")}>
                                            <span class={item.hidden ? "sr-only" : ""}>{item.label}</span>
                                        </th>
                                    }</For>
                                </tr>
                            </thead>
                            <tbody>
                                <For each={merged.childs}>{(item, index) =>
                                    <tr class={[
                                        "bg-white border-b dark:bg-gray-800 dark:border-gray-700",
                                        merged.onItemClick ? "cursor-pointer" : "",
                                        merged.hoverable ? "hover:bg-gray-50 dark:hover:bg-gray-900" : "",
                                    ].join(" ")}
                                        onClick={() => {
                                            if (merged.onItemClick) {
                                                merged.onItemClick(item, index())
                                            }
                                        }}>
                                        {merged.builder(item, index())}
                                    </tr>
                                }</For>
                            </tbody>
                        </table>
                    </div>
                </Match>
            </Switch>
        </>
    )
}

export default AppTable
