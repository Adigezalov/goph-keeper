import { BASE_APP_URL } from '@shared/constants'

export const SECRETS_URL = {
	BASE: BASE_APP_URL + '/v1/secrets',
	BY_ID: (id: string) => BASE_APP_URL + `/v1/secrets/${id}`,
	SYNC: BASE_APP_URL + '/v1/secrets/sync',
}

