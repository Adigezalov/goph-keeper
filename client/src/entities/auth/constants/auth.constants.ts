import { BASE_APP_URL, REFRESH } from '@shared/constants'

export const AUTH_URL = {
	LOGIN: BASE_APP_URL + `/v1/user/login`,
	REGISTRATION: BASE_APP_URL + `/v1/user/register`,
	VERIFY_EMAIL: BASE_APP_URL + `/v1/user/verify-email`,
	RESEND_CODE: BASE_APP_URL + `/v1/user/resend-code`,
	LOGOUT: BASE_APP_URL + `/v1/user/logout`,
	LOGOUT_ALL: BASE_APP_URL + `/v1/user/logout-all`,
	REFRESH: REFRESH,
}
