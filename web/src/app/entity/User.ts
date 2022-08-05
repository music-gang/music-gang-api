
export type UserApp = {
    user: User
    access_token: string
    refresh_token: string
    expires_in: number
    token_type: string
}

export const newUserApp = ({ user, access_token, refresh_token, expires_in, token_type }: UserApp): UserApp => ({
    user,
    access_token,
    refresh_token,
    expires_in,
    token_type,
})

type User = {
    id: number
    friedlyname: string
    username: string
    enabled: bool
}

export const newUser = (): User => ({
    id: 0,
    friedlyname: "",
    username: "",
    enabled: true,
})

export const isNewUser = (user: Option<User>): bool => user === null || user.id === 0

export default User