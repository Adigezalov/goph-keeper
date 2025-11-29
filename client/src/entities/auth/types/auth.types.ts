export type TAuth = {
	email?: string
	password?: string
}

export type TVerifyEmail = {
	email: string
	code: string
}

export type TResendCode = {
	email: string
}