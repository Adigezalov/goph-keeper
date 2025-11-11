import { Toast as PrimeToast } from 'primereact/toast'
import { ForwardedRef, ReactNode, forwardRef, useImperativeHandle, useRef } from 'react'

import { TOAST_DEFAULT_TIME, TOAST_POSITION, TOAST_SEVERITY } from './toast.constants'
import { TToastRef } from './toast.types'

export const Toast = forwardRef((_, ref: ForwardedRef<TToastRef>): ReactNode => {
	const toastRef = useRef<PrimeToast>(null)

	useImperativeHandle(ref, () => ({
		show: (
			severity: TOAST_SEVERITY,
			message: string,
			detail: string,
			life = TOAST_DEFAULT_TIME,
		) => {
			toastRef.current?.show({
				severity,
				summary: message,
				detail,
				life,
			})
		},
	}))

	return <PrimeToast position={TOAST_POSITION.BOTTOM_RIGHT} ref={toastRef} />
})
