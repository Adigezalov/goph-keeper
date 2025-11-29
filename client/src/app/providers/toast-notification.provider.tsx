import { ReactNode, memo, useCallback, useEffect, useRef } from 'react'

import {
	ToastNotificationContext,
	setToastNotificationRef,
} from '@shared/toast-notification'
import { TOAST_SEVERITY, TToastRef, Toast } from '@shared/uikit/toast'

type TProps = {
	children: ReactNode
}

export const ToastNotificationProvider = memo(({ children }: TProps) => {
	const toast = useRef<TToastRef>(null)

	useEffect(() => {
		setToastNotificationRef(toast.current)
	}, [])

	const show = useCallback(
		({
			message,
			header,
			severity,
		}: {
			message: string
			header: string
			severity: TOAST_SEVERITY
		}) => {
			toast.current?.show(severity, header, message)
		},
		[],
	)

	return (
		<ToastNotificationContext.Provider value={{ show }}>
			{children}
			<Toast ref={toast} />
		</ToastNotificationContext.Provider>
	)
})
