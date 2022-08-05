/* @refresh reload */
import 'flowbite'

import './index.css'
import { render } from 'solid-js/web'

import App from './App'
import toast from 'solid-toast'
import AuthStore from './app/stores/AuthStore'
import queueManager, { Job } from './app/managers/QueueManager'
import { Err, ErrorCode, newError, Ok } from './app/util'


const onSessionExpired = (): void => {
    toast.error("Session expired. Please login again.")
    AuthStore.mutations.setUserApp(null)
}

queueManager.registerBehavior(ErrorCode.ENOTAUTHENTICATED,
    new Job(
        async () => {
            await AuthStore.actions.refreshToken()
            if (AuthStore.state.userApp) {
                return Ok(true)
            }
            return Err(newError(ErrorCode.ENOTAUTHENTICATED, "Failed to refresh token, should logout"))
        },
        {
            rejector: onSessionExpired,
        }
    )
)
queueManager.registerBehavior(ErrorCode.EUNAUTHORIZED,
    new Job(
        async () => {
            await AuthStore.actions.refreshToken()
            if (AuthStore.state.userApp) {
                return Ok(true)
            }
            return Err(newError(ErrorCode.EUNAUTHORIZED, "Failed to refresh token, should logout"))
        },
        {
            rejector: onSessionExpired,
        }
    )
)

AuthStore.actions.fetchUser()

render(() => <App />, document.getElementById('root') as HTMLElement)
