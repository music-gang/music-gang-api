import { children, Component, JSXElement, mergeProps, Show } from "solid-js"
import AppButton, { ButtonColor } from "../inputs/AppButton"


const AppPopUpModal: Component<{
    id: string,
    svg?: JSXElement,
    text: string | JSXElement,
    leftLabel?: string | JSXElement,
    leftColor?: ButtonColor,
    fullHeight?: boolean,
    maxWidth?: string,
    width?: string,
    onClose?: (e: MouseEvent) => void,
    onLeftClick?: (e: MouseEvent) => void,
    centerLabel?: string,
    centerColor?: ButtonColor,
    onCenterClick?: (e: MouseEvent) => void,
    rightLabel?: string | JSXElement,
    rightColor?: ButtonColor,
    onRightClick?: (e: MouseEvent) => void,
    showCloseButton?: bool,
}> = (props) => {

    const merged = mergeProps({
        svg: <></>,
        maxWidth: "md",
        width: "w-full",
        fullHeight: false,
        leftColor: "red" as ButtonColor,
        rightColor: "light" as ButtonColor,
        centerColor: "primary" as ButtonColor,
        showCloseButton: true,
    }, props)

    const leftLabel = children(() => merged.leftLabel)
    const centerLabel = children(() => merged.centerLabel)
    const rightLabel = children(() => merged.rightLabel)
    const svg = children(() => merged.svg)
    const text = children(() => merged.text)

    return (
        <>
            <div id={merged.id} tabindex="-1" class={[
                "overflow-y-auto overflow-x-hidden fixed top-0 right-0 left-0 z-50 md:inset-0 h-modal md:h-full justify-center items-center flex w-full",
            ].join(" ")}
            >
                <div class={[merged.width, "relative p-4", merged.fullHeight ? "min-h-full h-full" : "h-auto"].join(" ")}>
                    <div class={["relative bg-white rounded-lg shadow dark:bg-gray-700", merged.fullHeight ? "h-full" : ""]
                        .join(" ")}>
                        <Show when={merged.showCloseButton}>
                            <button
                                type="button"
                                onClick={merged.onClose}
                                class="absolute top-3 right-2.5 text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-800 dark:hover:text-white">
                                <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path></svg>
                            </button>
                        </Show>
                        <div class="p-6 text-center">
                            {svg()}
                            <h3 class="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">{text()}</h3>
                            <div class="flex justify-evenly">
                                <Show when={leftLabel()}>
                                    <AppButton
                                        type="button"
                                        color={merged.leftColor}
                                        onClick={merged.onLeftClick}
                                        label={leftLabel()} />
                                </Show>
                                <Show when={centerLabel()}>
                                    <AppButton
                                        type="button"
                                        color={merged.centerColor}
                                        onClick={merged.onCenterClick}
                                        label={centerLabel()} />
                                </Show>
                                <Show when={rightLabel()}>
                                    <AppButton
                                        type="button"
                                        color={merged.rightColor}
                                        onClick={merged.onRightClick}
                                        label={rightLabel()} />
                                </Show>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            <div modal-backdrop="" class="bg-gray-900 bg-opacity-50 dark:bg-opacity-80 fixed inset-0 z-40"></div>
        </>
    )
}

export default AppPopUpModal