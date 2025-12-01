import { LOCAL_TOKEN_NAME } from './tokens.constants'

export const getAccessToken = () => {
	return localStorage.getItem(LOCAL_TOKEN_NAME.ACCESS)
}

export const removeTokens = () => {
	localStorage.removeItem(LOCAL_TOKEN_NAME.ACCESS)
}

export const setTokens = (accessToken: string) => {
	localStorage.setItem(LOCAL_TOKEN_NAME.ACCESS, accessToken)
}
