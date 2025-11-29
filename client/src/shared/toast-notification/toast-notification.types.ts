import { type TOAST_SEVERITY } from '../uikit/toast'

export type TToastNotificationContext = {
	show: ({
		message,
		header,
		severity,
	}: {
		message: string
		header: string
		severity: TOAST_SEVERITY
	}) => void
}
