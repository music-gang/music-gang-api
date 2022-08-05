import User, { UserApp } from "../entity/User"
import { Err, Error, Ok, Result } from "../util"

const authService = () => ({
    getUser: async (): Promise<Result<User, Error>> => {

        const response = await fetch("/v1/user", {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                // "Authorization": `Bearer ${AuthStore.getters.tokenPair.access_token}`,
            },
        })

        if (response.status === 200) {
            return Ok((await response.json() as { user: User }).user)
        }

        return Err(await response.json() as unknown as Error)
    },
    login: async (username: string, password: string): Promise<Result<UserApp, Error>> => {

        const response = await fetch("/v1/auth/login", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({ username, password }),

        })

        if (response.status === 200) {
            const tokenPair = await response.json() as UserApp
            return Ok(tokenPair)
        }

        return Err(await response.json() as unknown as Error)
    },
    logout: async ({ accessToken, refreshToken }: {
        accessToken?: string, refreshToken?: string
    }): Promise<Result<null, Error>> => {

        const response = await fetch("/v1/auth/logout", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                access_token: accessToken,
                refresh_token: refreshToken,
            }),
        })

        if (response.status === 200) {
            return Ok(null)
        }

        return Err(await response.json() as unknown as Error)
    },
    refreshToken: async (refreshToken: string): Promise<Result<UserApp, Error>> => {

        const response = await fetch("/v1/auth/refresh", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ refresh_token: refreshToken }),
        })

        if (response.status === 200) {
            return Ok(await response.json() as UserApp)
        }

        return Err(await response.json() as unknown as Error)
    }
})

export default authService()