import { api } from '@shared/api'
import { IResponse } from '@shared/types'

import { SERVER_STATUS_URL } from '../constants'

export const serverStatusApi = (): Promise<IResponse<{ status: 'ok' }>> => {
	return api.patch(SERVER_STATUS_URL.CHECK)
}
