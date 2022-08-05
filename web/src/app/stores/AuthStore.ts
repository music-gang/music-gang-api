import { createRoot } from "solid-js"
import { createStore, produce } from "solid-js/store"
import toast from "solid-toast"
import User, { newUserApp, UserApp } from "../entity/User"
import queueManager from "../managers/QueueManager"
import AuthService from "../services/AuthService"

type AuthStoreState = {
    userApp: Option<UserApp>,
    updating: bool
}

const createAuthStore = () => {

    const [state, setState] = createStore({
        userApp: null,
        updating: false,
    } as AuthStoreState)

    const getters = {
        get isLoggedIn() {
            return state.userApp != null
                && state.userApp.user != null
                && state.userApp.access_token
        },
        get tokenPair() {
            return {
                access_token: state.userApp?.access_token,
                refresh_token: state.userApp?.refresh_token,
            }
        },
        get user() {
            return state.userApp?.user
        }
    }

    const mutations = {
        setUpdating: (updating: bool) => {
            setState(
                produce((state) => {
                    state.updating = updating
                })
            )
        },
        setUser: (user: User) => {
            setState(
                produce((state) => {
                    if (state.userApp) {
                        state.userApp.user = user
                        localStorage.setItem("user", JSON.stringify(state.userApp))
                    }
                })
            )
        },
        setUserApp: (userApp: Option<UserApp>) => {
            setState(
                produce((state) => {
                    state.userApp = userApp
                    if (userApp) {
                        localStorage.setItem("user", JSON.stringify(userApp))
                    } else {
                        localStorage.removeItem("user")
                        queueManager.flush()
                    }
                })
            )
        }
    }

    const actions = {
        fetchUser: async () => {
            if (!getters.tokenPair.access_token) {
                return
            }
            queueManager.add(() => AuthService.getUser(),
                (user) => {
                    mutations.setUser(user)
                },
                () => {
                    mutations.setUserApp(null)
                }
            )
        },
        login: async (username: string, password: string) => {
            if (state.updating) {
                return
            }
            mutations.setUpdating(true)
            const tokenPair = await AuthService.login(username, password)
            if (tokenPair.isOk()) {
                mutations.setUserApp(tokenPair.unwrap())
            } else {
                mutations.setUserApp(null)
                toast.error(tokenPair.unwrapErr().message)
            }
            mutations.setUpdating(false)
        },
        logout: async () => {
            if (state.updating) {
                return
            }
            mutations.setUpdating(true)
            AuthService.logout({
                accessToken: getters.tokenPair.access_token,
                refreshToken: getters.tokenPair.refresh_token,
            })
            mutations.setUpdating(false)
            mutations.setUserApp(null)
        },
        refreshToken: async () => {
            if (state.updating || !getters.tokenPair.refresh_token) {
                return
            }
            mutations.setUpdating(true)
            const tokenPair = await AuthService.refreshToken(getters.tokenPair.refresh_token)
            if (tokenPair.isOk()) {
                mutations.setUserApp(tokenPair.unwrap())
            } else {
                mutations.setUserApp(null)
            }
            mutations.setUpdating(false)
        }
    }

    const init = () => {
        // check in local storage if user is logged in
        const user = localStorage.getItem('user')
        if (user != undefined && user != null && user != "" && user != "undefined" && user != "null") {
            mutations.setUserApp(newUserApp(JSON.parse(user)))
        } else {
            mutations.setUserApp(null)
        }
    }

    init()

    return {
        getters,
        state,
        mutations,
        actions,
    }
}

export default createRoot(createAuthStore)