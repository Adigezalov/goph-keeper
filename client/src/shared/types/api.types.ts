export interface IResponse<T> {
	data: T
	status: number
}

export interface IApiError {
	code: string
	message?: string
	response?: {
		status: number
		data?: {
			message?: string
		}
	}
}
