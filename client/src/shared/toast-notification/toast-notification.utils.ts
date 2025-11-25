import { TOAST_SEVERITY, TToastRef } from '../uikit/toast'

let toastRef: TToastRef | null = null

export const setToastNotificationRef = (ref: TToastRef | null) => {
	toastRef = ref
}

export const showToastNotification = ({
	message,
	header,
	severity,
}: {
	message: string
	header: string
	severity: TOAST_SEVERITY
}) => {
	if (toastRef) {
		toastRef.show(severity, header, message)
	}
}
