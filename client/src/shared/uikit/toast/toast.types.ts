import { TOAST_SEVERITY } from './toast.constants'

export type TToastRef = {
	show: (
		severity: TOAST_SEVERITY,
		message: string,
		detail: string,
		life?: number,
	) => void
}
